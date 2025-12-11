package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strpe-connect/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StripeConnectRepository interface {
	// Developer Wallet operations
	CreateDeveloperWallet(ctx context.Context, organizationID string) (*models.DeveloperWallet, error)
	GetDeveloperWalletByOrgID(ctx context.Context, organizationID string) (*models.DeveloperWallet, error)
	GetAllDeveloperWallets(ctx context.Context, limit, offset int) ([]*models.DeveloperWallet, error)
	UpdateStripeConnectAccountID(ctx context.Context, walletID, stripeAccountID string) error
	UpdateOnboardingStatus(ctx context.Context, walletID string, completed, payoutsEnabled, chargesEnabled bool) error
	UpdateWalletBalance(ctx context.Context, walletID string, amount float64) error
	GetWalletByID(ctx context.Context, walletID string) (*models.DeveloperWallet, error)

	// Withdrawal operations
	CreateWithdrawalRequest(ctx context.Context, withdrawal *models.WithdrawalRequest) error
	GetWithdrawalByID(ctx context.Context, withdrawalID string) (*models.WithdrawalRequest, error)
	UpdateWithdrawalStatus(ctx context.Context, withdrawalID, status string, stripeTransferID *string, failureReason *string) error
	GetWithdrawalsByOrgID(ctx context.Context, organizationID string, limit, offset int) ([]*models.WithdrawalRequest, error)
	GetPendingWithdrawalsTotal(ctx context.Context, walletID string) (float64, error)

	// Transaction operations
	CreateTransaction(ctx context.Context, tx *models.FunctionExecutionTransaction) error
	GetTransactionsByDeveloperOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.FunctionExecutionTransaction, error)
	GetTransactionsByUserOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.FunctionExecutionTransaction, error)
	GetTransactionByID(ctx context.Context, transactionID string) (*models.FunctionExecutionTransaction, error)
	GetConnectedDevelopersByUserOrg(ctx context.Context, userOrgID string) ([]*models.DeveloperWallet, error)

	// Account operations (user balance)
	GetAccountByOrgID(ctx context.Context, orgID string) (*Account, error)
	DeductUserBalance(ctx context.Context, accountID string, amount float64) error
}

type stripeConnectRepository struct {
	db *pgxpool.Pool
}

func NewStripeConnectRepository(db *pgxpool.Pool) StripeConnectRepository {
	return &stripeConnectRepository{db: db}
}

// Account represents user account (from existing schema)
type Account struct {
	ID             string  `db:"id"`
	OrganizationID string  `db:"organization_id"`
	AccountBalance float64 `db:"account_balance"`
}

// ================================
// DEVELOPER WALLET OPERATIONS
// ================================

func (r *stripeConnectRepository) CreateDeveloperWallet(ctx context.Context, organizationID string) (*models.DeveloperWallet, error) {
	wallet := &models.DeveloperWallet{
		ID:                  uuid.New().String(),
		OrganizationID:      organizationID,
		Balance:             0.00,
		TotalEarned:         0.00,
		TotalWithdrawn:      0.00,
		OnboardingCompleted: false,
		PayoutsEnabled:      false,
		ChargesEnabled:      false,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	query := `
		INSERT INTO tenant_schema.developer_wallets
		(id, organization_id, balance, total_earned, total_withdrawn, onboarding_completed, payouts_enabled, charges_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, organization_id, stripe_connect_account_id, balance, total_earned, total_withdrawn,
		          onboarding_completed, onboarding_url, payouts_enabled, charges_enabled, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		wallet.ID, wallet.OrganizationID, wallet.Balance, wallet.TotalEarned, wallet.TotalWithdrawn,
		wallet.OnboardingCompleted, wallet.PayoutsEnabled, wallet.ChargesEnabled,
		wallet.CreatedAt, wallet.UpdatedAt,
	).Scan(
		&wallet.ID, &wallet.OrganizationID, &wallet.StripeConnectAccountID, &wallet.Balance,
		&wallet.TotalEarned, &wallet.TotalWithdrawn, &wallet.OnboardingCompleted, &wallet.OnboardingURL,
		&wallet.PayoutsEnabled, &wallet.ChargesEnabled, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create developer wallet: %w", err)
	}

	return wallet, nil
}

func (r *stripeConnectRepository) GetDeveloperWalletByOrgID(ctx context.Context, organizationID string) (*models.DeveloperWallet, error) {
	wallet := &models.DeveloperWallet{}

	query := `
		SELECT id, organization_id, stripe_connect_account_id, balance, total_earned, total_withdrawn,
		       onboarding_completed, onboarding_url, payouts_enabled, charges_enabled, created_at, updated_at
		FROM tenant_schema.developer_wallets
		WHERE organization_id = $1
	`

	err := r.db.QueryRow(ctx, query, organizationID).Scan(
		&wallet.ID, &wallet.OrganizationID, &wallet.StripeConnectAccountID, &wallet.Balance,
		&wallet.TotalEarned, &wallet.TotalWithdrawn, &wallet.OnboardingCompleted, &wallet.OnboardingURL,
		&wallet.PayoutsEnabled, &wallet.ChargesEnabled, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("wallet not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *stripeConnectRepository) GetAllDeveloperWallets(ctx context.Context, limit, offset int) ([]*models.DeveloperWallet, error) {
	query := `
		SELECT dw.id, dw.organization_id, dw.stripe_connect_account_id, dw.balance, dw.total_earned, dw.total_withdrawn,
		       dw.onboarding_completed, dw.onboarding_url, dw.payouts_enabled, dw.charges_enabled, dw.created_at, dw.updated_at
		FROM tenant_schema.developer_wallets dw
		ORDER BY dw.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallets: %w", err)
	}
	defer rows.Close()

	wallets := []*models.DeveloperWallet{}
	for rows.Next() {
		wallet := &models.DeveloperWallet{}
		err := rows.Scan(
			&wallet.ID, &wallet.OrganizationID, &wallet.StripeConnectAccountID, &wallet.Balance,
			&wallet.TotalEarned, &wallet.TotalWithdrawn, &wallet.OnboardingCompleted, &wallet.OnboardingURL,
			&wallet.PayoutsEnabled, &wallet.ChargesEnabled, &wallet.CreatedAt, &wallet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return wallets, nil
}

func (r *stripeConnectRepository) GetWalletByID(ctx context.Context, walletID string) (*models.DeveloperWallet, error) {
	wallet := &models.DeveloperWallet{}

	query := `
		SELECT id, organization_id, stripe_connect_account_id, balance, total_earned, total_withdrawn,
		       onboarding_completed, onboarding_url, payouts_enabled, charges_enabled, created_at, updated_at
		FROM tenant_schema.developer_wallets
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, walletID).Scan(
		&wallet.ID, &wallet.OrganizationID, &wallet.StripeConnectAccountID, &wallet.Balance,
		&wallet.TotalEarned, &wallet.TotalWithdrawn, &wallet.OnboardingCompleted, &wallet.OnboardingURL,
		&wallet.PayoutsEnabled, &wallet.ChargesEnabled, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *stripeConnectRepository) UpdateStripeConnectAccountID(ctx context.Context, walletID, stripeAccountID string) error {
	query := `
		UPDATE tenant_schema.developer_wallets
		SET stripe_connect_account_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, stripeAccountID, walletID)
	if err != nil {
		return fmt.Errorf("failed to update stripe account ID: %w", err)
	}

	return nil
}

func (r *stripeConnectRepository) UpdateOnboardingStatus(ctx context.Context, walletID string, completed, payoutsEnabled, chargesEnabled bool) error {
	query := `
		UPDATE tenant_schema.developer_wallets
		SET onboarding_completed = $1, payouts_enabled = $2, charges_enabled = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, completed, payoutsEnabled, chargesEnabled, walletID)
	if err != nil {
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}

	return nil
}

func (r *stripeConnectRepository) UpdateWalletBalance(ctx context.Context, walletID string, amount float64) error {
	query := `
		UPDATE tenant_schema.developer_wallets
		SET balance = balance + $1,
		    total_earned = total_earned + CASE WHEN $1 > 0 THEN $1 ELSE 0 END,
		    total_withdrawn = total_withdrawn + CASE WHEN $1 < 0 THEN ABS($1) ELSE 0 END,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, amount, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return nil
}

// ================================
// WITHDRAWAL OPERATIONS
// ================================

func (r *stripeConnectRepository) CreateWithdrawalRequest(ctx context.Context, withdrawal *models.WithdrawalRequest) error {
	withdrawal.ID = uuid.New().String()
	withdrawal.RequestedAt = time.Now()
	withdrawal.CreatedAt = time.Now()
	withdrawal.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenant_schema.withdrawal_requests
		(id, developer_wallet_id, organization_id, amount, status, requested_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		withdrawal.ID, withdrawal.DeveloperWalletID, withdrawal.OrganizationID, withdrawal.Amount,
		withdrawal.Status, withdrawal.RequestedAt, withdrawal.CreatedAt, withdrawal.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create withdrawal request: %w", err)
	}

	return nil
}

func (r *stripeConnectRepository) GetWithdrawalByID(ctx context.Context, withdrawalID string) (*models.WithdrawalRequest, error) {
	withdrawal := &models.WithdrawalRequest{}

	query := `
		SELECT id, developer_wallet_id, organization_id, amount, status, stripe_transfer_id, stripe_payout_id,
		       failure_reason, requested_at, completed_at, created_at, updated_at
		FROM tenant_schema.withdrawal_requests
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, withdrawalID).Scan(
		&withdrawal.ID, &withdrawal.DeveloperWalletID, &withdrawal.OrganizationID, &withdrawal.Amount,
		&withdrawal.Status, &withdrawal.StripeTransferID, &withdrawal.StripePayoutID,
		&withdrawal.FailureReason, &withdrawal.RequestedAt, &withdrawal.CompletedAt,
		&withdrawal.CreatedAt, &withdrawal.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal: %w", err)
	}

	return withdrawal, nil
}

func (r *stripeConnectRepository) UpdateWithdrawalStatus(ctx context.Context, withdrawalID, status string, stripeTransferID *string, failureReason *string) error {
	query := `
		UPDATE tenant_schema.withdrawal_requests
		SET status = $1,
		    stripe_transfer_id = COALESCE($2, stripe_transfer_id),
		    failure_reason = COALESCE($3, failure_reason),
		    completed_at = CASE WHEN $1 IN ('completed', 'failed') THEN NOW() ELSE completed_at END,
		    updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, status, stripeTransferID, failureReason, withdrawalID)
	if err != nil {
		return fmt.Errorf("failed to update withdrawal status: %w", err)
	}

	return nil
}

func (r *stripeConnectRepository) GetWithdrawalsByOrgID(ctx context.Context, organizationID string, limit, offset int) ([]*models.WithdrawalRequest, error) {
	query := `
		SELECT id, developer_wallet_id, organization_id, amount, status, stripe_transfer_id, stripe_payout_id,
		       failure_reason, requested_at, completed_at, created_at, updated_at
		FROM tenant_schema.withdrawal_requests
		WHERE organization_id = $1
		ORDER BY requested_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, organizationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*models.WithdrawalRequest
	for rows.Next() {
		withdrawal := &models.WithdrawalRequest{}
		err := rows.Scan(
			&withdrawal.ID, &withdrawal.DeveloperWalletID, &withdrawal.OrganizationID, &withdrawal.Amount,
			&withdrawal.Status, &withdrawal.StripeTransferID, &withdrawal.StripePayoutID,
			&withdrawal.FailureReason, &withdrawal.RequestedAt, &withdrawal.CompletedAt,
			&withdrawal.CreatedAt, &withdrawal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, nil
}

func (r *stripeConnectRepository) GetPendingWithdrawalsTotal(ctx context.Context, walletID string) (float64, error) {
	var total float64

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM tenant_schema.withdrawal_requests
		WHERE developer_wallet_id = $1 AND status IN ('pending', 'processing')
	`

	err := r.db.QueryRow(ctx, query, walletID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending withdrawals: %w", err)
	}

	return total, nil
}

// ================================
// TRANSACTION OPERATIONS
// ================================

func (r *stripeConnectRepository) CreateTransaction(ctx context.Context, tx *models.FunctionExecutionTransaction) error {
	tx.ID = uuid.New().String()
	tx.ExecutedAt = time.Now()
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenant_schema.function_execution_transactions
		(id, function_id, user_organization_id, developer_organization_id, user_account_id, developer_wallet_id,
		 amount, platform_fee, net_amount, description, status, executed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(ctx, query,
		tx.ID, tx.FunctionID, tx.UserOrganizationID, tx.DeveloperOrganizationID, tx.UserAccountID,
		tx.DeveloperWalletID, tx.Amount, tx.PlatformFee, tx.NetAmount, tx.Description, tx.Status,
		tx.ExecutedAt, tx.CreatedAt, tx.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *stripeConnectRepository) GetTransactionsByDeveloperOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.FunctionExecutionTransaction, error) {
	query := `
		SELECT id, function_id, user_organization_id, developer_organization_id, user_account_id,
		       developer_wallet_id, amount, platform_fee, net_amount, description, status,
		       executed_at, created_at, updated_at
		FROM tenant_schema.function_execution_transactions
		WHERE developer_organization_id = $1
		ORDER BY executed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *stripeConnectRepository) GetTransactionsByUserOrg(ctx context.Context, orgID string, limit, offset int) ([]*models.FunctionExecutionTransaction, error) {
	query := `
		SELECT id, function_id, user_organization_id, developer_organization_id, user_account_id,
		       developer_wallet_id, amount, platform_fee, net_amount, description, status,
		       executed_at, created_at, updated_at
		FROM tenant_schema.function_execution_transactions
		WHERE user_organization_id = $1
		ORDER BY executed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *stripeConnectRepository) GetConnectedDevelopersByUserOrg(ctx context.Context, userOrgID string) ([]*models.DeveloperWallet, error) {
	query := `
		SELECT DISTINCT dw.id, dw.organization_id, dw.stripe_connect_account_id, dw.balance, dw.total_earned, dw.total_withdrawn,
		       dw.onboarding_completed, dw.onboarding_url, dw.payouts_enabled, dw.charges_enabled, dw.created_at, dw.updated_at,
		       SUM(tx.amount) as total_paid_to_developer,
		       COUNT(tx.id) as transaction_count
		FROM tenant_schema.developer_wallets dw
		INNER JOIN tenant_schema.function_execution_transactions tx
			ON tx.developer_organization_id = dw.organization_id
		WHERE tx.user_organization_id = $1
		GROUP BY dw.id, dw.organization_id, dw.stripe_connect_account_id, dw.balance, dw.total_earned, dw.total_withdrawn,
		         dw.onboarding_completed, dw.onboarding_url, dw.payouts_enabled, dw.charges_enabled, dw.created_at, dw.updated_at
		ORDER BY total_paid_to_developer DESC
	`

	rows, err := r.db.Query(ctx, query, userOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query connected developers: %w", err)
	}
	defer rows.Close()

	wallets := []*models.DeveloperWallet{}
	for rows.Next() {
		wallet := &models.DeveloperWallet{}
		var totalPaid float64
		var txCount int

		err := rows.Scan(
			&wallet.ID, &wallet.OrganizationID, &wallet.StripeConnectAccountID, &wallet.Balance,
			&wallet.TotalEarned, &wallet.TotalWithdrawn, &wallet.OnboardingCompleted, &wallet.OnboardingURL,
			&wallet.PayoutsEnabled, &wallet.ChargesEnabled, &wallet.CreatedAt, &wallet.UpdatedAt,
			&totalPaid, &txCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return wallets, nil
}

func (r *stripeConnectRepository) GetTransactionByID(ctx context.Context, transactionID string) (*models.FunctionExecutionTransaction, error) {
	tx := &models.FunctionExecutionTransaction{}

	query := `
		SELECT id, function_id, user_organization_id, developer_organization_id, user_account_id,
		       developer_wallet_id, amount, platform_fee, net_amount, description, status,
		       executed_at, created_at, updated_at
		FROM tenant_schema.function_execution_transactions
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, transactionID).Scan(
		&tx.ID, &tx.FunctionID, &tx.UserOrganizationID, &tx.DeveloperOrganizationID,
		&tx.UserAccountID, &tx.DeveloperWalletID, &tx.Amount, &tx.PlatformFee, &tx.NetAmount,
		&tx.Description, &tx.Status, &tx.ExecutedAt, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

func (r *stripeConnectRepository) scanTransactions(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
}) ([]*models.FunctionExecutionTransaction, error) {
	var transactions []*models.FunctionExecutionTransaction

	for rows.Next() {
		tx := &models.FunctionExecutionTransaction{}
		err := rows.Scan(
			&tx.ID, &tx.FunctionID, &tx.UserOrganizationID, &tx.DeveloperOrganizationID,
			&tx.UserAccountID, &tx.DeveloperWalletID, &tx.Amount, &tx.PlatformFee, &tx.NetAmount,
			&tx.Description, &tx.Status, &tx.ExecutedAt, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// ================================
// ACCOUNT OPERATIONS
// ================================

func (r *stripeConnectRepository) GetAccountByOrgID(ctx context.Context, orgID string) (*Account, error) {
	account := &Account{}

	query := `
		SELECT id, organization_id, account_balance
		FROM tenant_schema.accounts
		WHERE organization_id = $1
	`

	err := r.db.QueryRow(ctx, query, orgID).Scan(&account.ID, &account.OrganizationID, &account.AccountBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

func (r *stripeConnectRepository) DeductUserBalance(ctx context.Context, accountID string, amount float64) error {
	query := `
		UPDATE tenant_schema.accounts
		SET account_balance = account_balance - $1, updated_at = NOW()
		WHERE id = $2 AND account_balance >= $1
	`

	result, err := r.db.Exec(ctx, query, amount, accountID)
	if err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}
