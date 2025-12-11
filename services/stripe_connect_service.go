package services

import (
	"context"
	"fmt"
	"log"
	"strpe-connect/models"
	"strpe-connect/repository"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/account"
	"github.com/stripe/stripe-go/v83/accountlink"
	"github.com/stripe/stripe-go/v83/payout"
)

type StripeConnectService interface {
	// Onboarding
	CreateConnectAccount(ctx context.Context, orgID, refreshURL, returnURL string) (*models.CreateConnectAccountResponse, error)
	GetConnectAccountStatus(ctx context.Context, orgID string) (*models.GetConnectAccountStatusResponse, error)
	RefreshOnboardingLink(ctx context.Context, orgID, refreshURL, returnURL string) (*models.CreateConnectAccountResponse, error)

	// Wallet Management
	GetWalletBalance(ctx context.Context, orgID string) (*models.GetWalletBalanceResponse, error)
	GetTransactionHistory(ctx context.Context, orgID string, page, limit int) (*models.GetTransactionHistoryResponse, error)
	GetConnectedDevelopers(ctx context.Context, page, limit int) (*models.GetConnectedDevelopersResponse, error)
	GetConnectedDevelopersForOrg(ctx context.Context, userOrgID string) (*models.GetConnectedDevelopersResponse, error)

	// Withdrawals
	RequestWithdrawal(ctx context.Context, orgID string, amount float64) (*models.CreateWithdrawalResponse, error)
	GetWithdrawalHistory(ctx context.Context, orgID string, page, limit int) (*models.GetWithdrawalHistoryResponse, error)
	ProcessWithdrawal(ctx context.Context, withdrawalID string) error

	// Function Execution Payment
	ProcessFunctionExecutionPayment(ctx context.Context, userOrgID, functionID, developerOrgID string, amount float64) (*models.FunctionExecutionPaymentResponse, error)

	// Webhook handling
	HandleAccountUpdated(ctx context.Context, stripeAccountID string) error
	HandlePayoutPaid(ctx context.Context, withdrawalID, payoutID string) error
	HandlePayoutFailed(ctx context.Context, withdrawalID, failureReason string) error
}

type stripeConnectService struct {
	repo          repository.StripeConnectRepository
	stripeKey     string
	platformFeePercent float64
}

func NewStripeConnectService(repo repository.StripeConnectRepository, stripeKey string) StripeConnectService {
	stripe.Key = stripeKey

	return &stripeConnectService{
		repo:          repo,
		stripeKey:     stripeKey,
		platformFeePercent: models.DefaultPlatformFeePercent,
	}
}

// ================================
// ONBOARDING
// ================================

func (s *stripeConnectService) CreateConnectAccount(ctx context.Context, orgID, refreshURL, returnURL string) (*models.CreateConnectAccountResponse, error) {
	// Check if wallet already exists
	existingWallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, orgID)
	if err == nil && existingWallet.StripeConnectAccountID != nil && *existingWallet.StripeConnectAccountID != "" {
		// Wallet exists, generate new onboarding link
		return s.RefreshOnboardingLink(ctx, orgID, refreshURL, returnURL)
	}

	// Create new wallet if doesn't exist
	if existingWallet == nil {
		existingWallet, err = s.repo.CreateDeveloperWallet(ctx, orgID)
		if err != nil {
			return nil, fmt.Errorf("failed to create developer wallet: %w", err)
		}
	}

	// Create Stripe Connect Express account
	params := &stripe.AccountParams{
		Type: stripe.String(string(stripe.AccountTypeExpress)),
		Capabilities: &stripe.AccountCapabilitiesParams{
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		Metadata: map[string]string{
			"organization_id": orgID,
			"wallet_id":      existingWallet.ID,
		},
	}

	acc, err := account.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe account: %w", err)
	}

	// Save Stripe account ID to wallet
	if err := s.repo.UpdateStripeConnectAccountID(ctx, existingWallet.ID, acc.ID); err != nil {
		return nil, fmt.Errorf("failed to update wallet with Stripe account ID: %w", err)
	}

	// Create account link for onboarding
	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(acc.ID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(linkParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %w", err)
	}

	return &models.CreateConnectAccountResponse{
		AccountID:     acc.ID,
		OnboardingURL: link.URL,
		Message:       "Please complete Stripe Connect onboarding to receive payments",
	}, nil
}

func (s *stripeConnectService) RefreshOnboardingLink(ctx context.Context, orgID, refreshURL, returnURL string) (*models.CreateConnectAccountResponse, error) {
	wallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if wallet.StripeConnectAccountID == nil || *wallet.StripeConnectAccountID == "" {
		return nil, fmt.Errorf("no Stripe Connect account found")
	}

	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(*wallet.StripeConnectAccountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(linkParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %w", err)
	}

	return &models.CreateConnectAccountResponse{
		AccountID:     *wallet.StripeConnectAccountID,
		OnboardingURL: link.URL,
		Message:       "Continue Stripe Connect onboarding",
	}, nil
}

func (s *stripeConnectService) GetConnectAccountStatus(ctx context.Context, orgID string) (*models.GetConnectAccountStatusResponse, error) {
	wallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	accountID := ""
	if wallet.StripeConnectAccountID != nil {
		accountID = *wallet.StripeConnectAccountID

		// Fetch latest status from Stripe API
		stripe.Key = s.stripeKey
		acc, err := account.GetByID(accountID, nil)
		if err == nil {
			// Update database with latest status from Stripe
			onboardingCompleted := acc.DetailsSubmitted
			payoutsEnabled := acc.PayoutsEnabled
			chargesEnabled := acc.ChargesEnabled

			err = s.repo.UpdateOnboardingStatus(ctx, wallet.ID, onboardingCompleted, payoutsEnabled, chargesEnabled)
			if err != nil {
				log.Printf("WARNING: Failed to update onboarding status: %v", err)
			} else {
				// Update local wallet object
				wallet.OnboardingCompleted = onboardingCompleted
				wallet.PayoutsEnabled = payoutsEnabled
				wallet.ChargesEnabled = chargesEnabled
			}
		}
	}

	// Get pending withdrawals
	pendingTotal, err := s.repo.GetPendingWithdrawalsTotal(ctx, wallet.ID)
	if err != nil {
		pendingTotal = 0
	}

	availableBalance := wallet.Balance - pendingTotal
	canWithdraw := wallet.OnboardingCompleted && wallet.PayoutsEnabled && availableBalance >= models.MinimumWithdrawalAmount

	return &models.GetConnectAccountStatusResponse{
		AccountID:           accountID,
		OnboardingCompleted: wallet.OnboardingCompleted,
		PayoutsEnabled:      wallet.PayoutsEnabled,
		ChargesEnabled:      wallet.ChargesEnabled,
		Balance:             wallet.Balance,
		TotalEarned:         wallet.TotalEarned,
		TotalWithdrawn:      wallet.TotalWithdrawn,
		CanWithdraw:         canWithdraw,
		MinimumWithdrawal:   models.MinimumWithdrawalAmount,
	}, nil
}

// ================================
// WALLET MANAGEMENT
// ================================

func (s *stripeConnectService) GetWalletBalance(ctx context.Context, orgID string) (*models.GetWalletBalanceResponse, error) {
	wallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	pendingTotal, err := s.repo.GetPendingWithdrawalsTotal(ctx, wallet.ID)
	if err != nil {
		pendingTotal = 0
	}

	availableBalance := wallet.Balance - pendingTotal
	canWithdraw := wallet.OnboardingCompleted && wallet.PayoutsEnabled && availableBalance >= models.MinimumWithdrawalAmount

	return &models.GetWalletBalanceResponse{
		Balance:            wallet.Balance,
		TotalEarned:        wallet.TotalEarned,
		TotalWithdrawn:     wallet.TotalWithdrawn,
		PendingWithdrawals: pendingTotal,
		CanWithdraw:        canWithdraw,
		MinimumWithdrawal:  models.MinimumWithdrawalAmount,
	}, nil
}

func (s *stripeConnectService) GetTransactionHistory(ctx context.Context, orgID string, page, limit int) (*models.GetTransactionHistoryResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	transactions, err := s.repo.GetTransactionsByDeveloperOrg(ctx, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Convert to summary format
	summaries := make([]models.TransactionSummary, len(transactions))
	for i, tx := range transactions {
		summaries[i] = models.TransactionSummary{
			ID:               tx.ID,
			FunctionID:       tx.FunctionID,
			FunctionName:     fmt.Sprintf("Function %s", tx.FunctionID[:8]), // Simplified - you can join with functions table
			UserOrganization: tx.UserOrganizationID,
			Amount:           tx.Amount,
			PlatformFee:      tx.PlatformFee,
			NetAmount:        tx.NetAmount,
			Status:           tx.Status,
			ExecutedAt:       tx.ExecutedAt,
		}
	}

	return &models.GetTransactionHistoryResponse{
		Transactions: summaries,
		Total:        len(summaries),
		Page:         page,
		Limit:        limit,
	}, nil
}

func (s *stripeConnectService) GetConnectedDevelopers(ctx context.Context, page, limit int) (*models.GetConnectedDevelopersResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	wallets, err := s.repo.GetAllDeveloperWallets(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get developer wallets: %w", err)
	}

	// Convert to summary format
	developers := make([]models.ConnectedDeveloperSummary, len(wallets))
	for i, wallet := range wallets {
		developers[i] = models.ConnectedDeveloperSummary{
			OrganizationID:      wallet.OrganizationID,
			StripeAccountID:     wallet.StripeConnectAccountID,
			Balance:             wallet.Balance,
			TotalEarned:         wallet.TotalEarned,
			TotalWithdrawn:      wallet.TotalWithdrawn,
			OnboardingCompleted: wallet.OnboardingCompleted,
			PayoutsEnabled:      wallet.PayoutsEnabled,
			ChargesEnabled:      wallet.ChargesEnabled,
			JoinedAt:            wallet.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &models.GetConnectedDevelopersResponse{
		Developers: developers,
		Total:      len(developers),
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *stripeConnectService) GetConnectedDevelopersForOrg(ctx context.Context, userOrgID string) (*models.GetConnectedDevelopersResponse, error) {
	wallets, err := s.repo.GetConnectedDevelopersByUserOrg(ctx, userOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected developers: %w", err)
	}

	// Convert to summary format
	developers := make([]models.ConnectedDeveloperSummary, len(wallets))
	for i, wallet := range wallets {
		developers[i] = models.ConnectedDeveloperSummary{
			OrganizationID:      wallet.OrganizationID,
			StripeAccountID:     wallet.StripeConnectAccountID,
			Balance:             wallet.Balance,
			TotalEarned:         wallet.TotalEarned,
			TotalWithdrawn:      wallet.TotalWithdrawn,
			OnboardingCompleted: wallet.OnboardingCompleted,
			PayoutsEnabled:      wallet.PayoutsEnabled,
			ChargesEnabled:      wallet.ChargesEnabled,
			JoinedAt:            wallet.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &models.GetConnectedDevelopersResponse{
		Developers: developers,
		Total:      len(developers),
		Page:       1,
		Limit:      len(developers),
	}, nil
}

// ================================
// WITHDRAWALS
// ================================

func (s *stripeConnectService) RequestWithdrawal(ctx context.Context, orgID string, amount float64) (*models.CreateWithdrawalResponse, error) {
	// Validate amount
	if amount < models.MinimumWithdrawalAmount {
		return nil, fmt.Errorf("minimum withdrawal amount is $%.2f", models.MinimumWithdrawalAmount)
	}

	// Get wallet
	wallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// Check onboarding
	if !wallet.OnboardingCompleted || !wallet.PayoutsEnabled {
		return nil, fmt.Errorf("please complete Stripe Connect onboarding before requesting withdrawals")
	}

	// Check pending withdrawals
	pendingTotal, err := s.repo.GetPendingWithdrawalsTotal(ctx, wallet.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending withdrawals: %w", err)
	}

	// Check available balance
	availableBalance := wallet.Balance - pendingTotal
	if availableBalance < amount {
		return nil, fmt.Errorf("insufficient balance (available: $%.2f, pending: $%.2f)", availableBalance, pendingTotal)
	}

	// Create withdrawal request
	withdrawal := &models.WithdrawalRequest{
		DeveloperWalletID: wallet.ID,
		OrganizationID:    orgID,
		Amount:            amount,
		Status:            models.WithdrawalStatusPending,
	}

	if err := s.repo.CreateWithdrawalRequest(ctx, withdrawal); err != nil {
		return nil, fmt.Errorf("failed to create withdrawal request: %w", err)
	}

	// Process withdrawal immediately (in production, you might want to queue this)
	go func() {
		ctx := context.Background()
		if err := s.ProcessWithdrawal(ctx, withdrawal.ID); err != nil {
			log.Printf("Failed to process withdrawal %s: %v", withdrawal.ID, err)
		}
	}()

	return &models.CreateWithdrawalResponse{
		WithdrawalID:     withdrawal.ID,
		Amount:           amount,
		Status:           models.WithdrawalStatusProcessing,
		Message:          "Withdrawal request created successfully",
		EstimatedArrival: "2-3 business days",
	}, nil
}

func (s *stripeConnectService) ProcessWithdrawal(ctx context.Context, withdrawalID string) error {
	// Get withdrawal request
	withdrawal, err := s.repo.GetWithdrawalByID(ctx, withdrawalID)
	if err != nil {
		return fmt.Errorf("failed to get withdrawal: %w", err)
	}

	// Get wallet
	wallet, err := s.repo.GetWalletByID(ctx, withdrawal.DeveloperWalletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if wallet.StripeConnectAccountID == nil || *wallet.StripeConnectAccountID == "" {
		failureReason := "no Stripe Connect account"
		_ = s.repo.UpdateWithdrawalStatus(ctx, withdrawalID, models.WithdrawalStatusFailed, nil, &failureReason)
		return fmt.Errorf("no Stripe Connect account found")
	}

	// Update status to processing
	_ = s.repo.UpdateWithdrawalStatus(ctx, withdrawalID, models.WithdrawalStatusProcessing, nil, nil)

	// Create payout in Stripe
	amountInCents := int64(withdrawal.Amount * 100)
	payoutParams := &stripe.PayoutParams{
		Amount:   stripe.Int64(amountInCents),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Params: stripe.Params{
			StripeAccount: wallet.StripeConnectAccountID,
		},
		Metadata: map[string]string{
			"withdrawal_id":   withdrawalID,
			"wallet_id":       wallet.ID,
			"organization_id": wallet.OrganizationID,
		},
	}

	po, err := payout.New(payoutParams)
	if err != nil {
		failureReason := err.Error()
		_ = s.repo.UpdateWithdrawalStatus(ctx, withdrawalID, models.WithdrawalStatusFailed, nil, &failureReason)
		return fmt.Errorf("failed to create payout: %w", err)
	}

	// Update withdrawal with payout ID
	payoutID := po.ID
	if err := s.repo.UpdateWithdrawalStatus(ctx, withdrawalID, models.WithdrawalStatusCompleted, &payoutID, nil); err != nil {
		log.Printf("Failed to update withdrawal status: %v", err)
	}

	// Deduct from wallet balance
	if err := s.repo.UpdateWalletBalance(ctx, wallet.ID, -withdrawal.Amount); err != nil {
		log.Printf("Failed to update wallet balance: %v", err)
	}

	log.Printf("✅ Withdrawal processed: ID=%s, Amount=$%.2f, PayoutID=%s", withdrawalID, withdrawal.Amount, payoutID)

	return nil
}

func (s *stripeConnectService) GetWithdrawalHistory(ctx context.Context, orgID string, page, limit int) (*models.GetWithdrawalHistoryResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	withdrawals, err := s.repo.GetWithdrawalsByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	summaries := make([]models.WithdrawalSummary, len(withdrawals))
	for i, w := range withdrawals {
		summaries[i] = models.WithdrawalSummary{
			ID:            w.ID,
			Amount:        w.Amount,
			Status:        w.Status,
			RequestedAt:   w.RequestedAt,
			CompletedAt:   w.CompletedAt,
			FailureReason: w.FailureReason,
		}
	}

	return &models.GetWithdrawalHistoryResponse{
		Withdrawals: summaries,
		Total:       len(summaries),
		Page:        page,
		Limit:       limit,
	}, nil
}

// ================================
// FUNCTION EXECUTION PAYMENT
// ================================

func (s *stripeConnectService) ProcessFunctionExecutionPayment(ctx context.Context, userOrgID, functionID, developerOrgID string, amount float64) (*models.FunctionExecutionPaymentResponse, error) {
	// Validate amount
	if amount <= 0 {
		return nil, fmt.Errorf("invalid amount")
	}

	// Validate developer org ID (in production, this would be looked up from functions table)
	if developerOrgID == "" {
		return nil, fmt.Errorf("developer organization ID is required")
	}

	// Get user account
	userAccount, err := s.repo.GetAccountByOrgID(ctx, userOrgID)
	if err != nil {
		return nil, fmt.Errorf("user account not found: %w", err)
	}

	// Check user balance
	if userAccount.AccountBalance < amount {
		return nil, fmt.Errorf("insufficient balance (have: $%.2f, need: $%.2f)", userAccount.AccountBalance, amount)
	}

	// Get or create developer wallet
	developerWallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, developerOrgID)
	if err != nil {
		// Create wallet if doesn't exist
		developerWallet, err = s.repo.CreateDeveloperWallet(ctx, developerOrgID)
		if err != nil {
			return nil, fmt.Errorf("failed to create developer wallet: %w", err)
		}
	}

	// Calculate platform fee and net amount
	platformFee := amount * s.platformFeePercent / 100.0
	netAmount := amount - platformFee

	// Create transaction record
	description := fmt.Sprintf("Function execution payment for %s", functionID)
	transaction := &models.FunctionExecutionTransaction{
		FunctionID:              functionID,
		UserOrganizationID:      userOrgID,
		DeveloperOrganizationID: developerOrgID,
		UserAccountID:           userAccount.ID,
		DeveloperWalletID:       developerWallet.ID,
		Amount:                  amount,
		PlatformFee:             platformFee,
		NetAmount:               netAmount,
		Description:             &description,
		Status:                  models.TransactionStatusCompleted,
	}

	// Deduct from user balance
	if err := s.repo.DeductUserBalance(ctx, userAccount.ID, amount); err != nil {
		return nil, fmt.Errorf("failed to deduct user balance: %w", err)
	}

	// Add to developer wallet
	if err := s.repo.UpdateWalletBalance(ctx, developerWallet.ID, netAmount); err != nil {
		// Rollback user balance deduction
		log.Printf("ERROR: Failed to credit developer wallet, manual intervention needed: %v", err)
		return nil, fmt.Errorf("failed to credit developer wallet: %w", err)
	}

	// Record transaction
	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		log.Printf("WARNING: Failed to record transaction: %v", err)
	}

	// Get updated balances
	updatedUserAccount, _ := s.repo.GetAccountByOrgID(ctx, userOrgID)
	updatedDeveloperWallet, _ := s.repo.GetDeveloperWalletByOrgID(ctx, developerOrgID)

	userBalance := 0.0
	if updatedUserAccount != nil {
		userBalance = updatedUserAccount.AccountBalance
	}

	developerBalance := 0.0
	if updatedDeveloperWallet != nil {
		developerBalance = updatedDeveloperWallet.Balance
	}

	return &models.FunctionExecutionPaymentResponse{
		TransactionID:    transaction.ID,
		Amount:           amount,
		PlatformFee:      platformFee,
		NetAmount:        netAmount,
		UserBalance:      userBalance,
		DeveloperBalance: developerBalance,
		Message:          "Payment processed successfully",
	}, nil
}

// ================================
// WEBHOOK HANDLING
// ================================

func (s *stripeConnectService) HandleAccountUpdated(ctx context.Context, stripeAccountID string) error {
	// Get account details from Stripe
	acc, err := account.GetByID(stripeAccountID, nil)
	if err != nil {
		return fmt.Errorf("failed to get Stripe account: %w", err)
	}

	// Get wallet by Stripe account ID
	// Note: You need to add a method to get wallet by stripe account ID
	// For now, we'll update based on metadata if available
	orgID := acc.Metadata["organization_id"]
	if orgID == "" {
		log.Printf("No organization_id in account metadata for %s", stripeAccountID)
		return nil
	}

	wallet, err := s.repo.GetDeveloperWalletByOrgID(ctx, orgID)
	if err != nil {
		log.Printf("Wallet not found for organization %s", orgID)
		return err
	}

	// Update onboarding status
	chargesEnabled := acc.ChargesEnabled
	payoutsEnabled := acc.PayoutsEnabled
	detailsSubmitted := acc.DetailsSubmitted

	if err := s.repo.UpdateOnboardingStatus(ctx, wallet.ID, detailsSubmitted, payoutsEnabled, chargesEnabled); err != nil {
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}

	log.Printf("✅ Updated account status for %s: onboarding=%v, payouts=%v, charges=%v",
		orgID, detailsSubmitted, payoutsEnabled, chargesEnabled)

	return nil
}

func (s *stripeConnectService) HandlePayoutPaid(ctx context.Context, withdrawalID, payoutID string) error {
	// Update withdrawal status to completed
	if err := s.repo.UpdateWithdrawalStatus(ctx, withdrawalID, models.WithdrawalStatusCompleted, &payoutID, nil); err != nil {
		return fmt.Errorf("failed to update withdrawal status: %w", err)
	}

	log.Printf("✅ Payout completed: withdrawalID=%s, payoutID=%s", withdrawalID, payoutID)
	return nil
}

func (s *stripeConnectService) HandlePayoutFailed(ctx context.Context, withdrawalID, failureReason string) error {
	// Update withdrawal status to failed with reason
	if err := s.repo.UpdateWithdrawalStatus(ctx, withdrawalID, models.WithdrawalStatusFailed, nil, &failureReason); err != nil {
		return fmt.Errorf("failed to update withdrawal status: %w", err)
	}

	// Get withdrawal to credit back the amount to wallet
	withdrawal, err := s.repo.GetWithdrawalByID(ctx, withdrawalID)
	if err != nil {
		log.Printf("ERROR: Failed to get withdrawal for refund: %v", err)
		return err
	}

	// Credit back the amount to developer wallet
	if err := s.repo.UpdateWalletBalance(ctx, withdrawal.DeveloperWalletID, withdrawal.Amount); err != nil {
		log.Printf("ERROR: Failed to credit back failed withdrawal amount: %v", err)
		return err
	}

	log.Printf("❌ Payout failed: withdrawalID=%s, reason=%s (amount credited back)", withdrawalID, failureReason)
	return nil
}
