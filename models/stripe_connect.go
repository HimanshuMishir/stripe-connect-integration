package models

import (
	"time"
)

// DeveloperWallet represents a developer's earnings wallet
type DeveloperWallet struct {
	ID                      string    `json:"id" db:"id"`
	OrganizationID          string    `json:"organization_id" db:"organization_id"`
	StripeConnectAccountID  *string   `json:"stripe_connect_account_id" db:"stripe_connect_account_id"`
	Balance                 float64   `json:"balance" db:"balance"`
	TotalEarned             float64   `json:"total_earned" db:"total_earned"`
	TotalWithdrawn          float64   `json:"total_withdrawn" db:"total_withdrawn"`
	OnboardingCompleted     bool      `json:"onboarding_completed" db:"onboarding_completed"`
	OnboardingURL           *string   `json:"onboarding_url" db:"onboarding_url"`
	PayoutsEnabled          bool      `json:"payouts_enabled" db:"payouts_enabled"`
	ChargesEnabled          bool      `json:"charges_enabled" db:"charges_enabled"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// WithdrawalRequest represents a developer's withdrawal request
type WithdrawalRequest struct {
	ID                  string     `json:"id" db:"id"`
	DeveloperWalletID   string     `json:"developer_wallet_id" db:"developer_wallet_id"`
	OrganizationID      string     `json:"organization_id" db:"organization_id"`
	Amount              float64    `json:"amount" db:"amount"`
	Status              string     `json:"status" db:"status"` // pending, processing, completed, failed, rejected
	StripeTransferID    *string    `json:"stripe_transfer_id" db:"stripe_transfer_id"`
	StripePayoutID      *string    `json:"stripe_payout_id" db:"stripe_payout_id"`
	FailureReason       *string    `json:"failure_reason" db:"failure_reason"`
	RequestedAt         time.Time  `json:"requested_at" db:"requested_at"`
	CompletedAt         *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// FunctionExecutionTransaction represents a payment for function execution
type FunctionExecutionTransaction struct {
	ID                      string    `json:"id" db:"id"`
	FunctionID              string    `json:"function_id" db:"function_id"`
	UserOrganizationID      string    `json:"user_organization_id" db:"user_organization_id"`
	DeveloperOrganizationID string    `json:"developer_organization_id" db:"developer_organization_id"`
	UserAccountID           string    `json:"user_account_id" db:"user_account_id"`
	DeveloperWalletID       string    `json:"developer_wallet_id" db:"developer_wallet_id"`
	Amount                  float64   `json:"amount" db:"amount"`
	PlatformFee             float64   `json:"platform_fee" db:"platform_fee"`
	NetAmount               float64   `json:"net_amount" db:"net_amount"`
	Description             *string   `json:"description" db:"description"`
	Status                  string    `json:"status" db:"status"` // completed, failed, refunded
	ExecutedAt              time.Time `json:"executed_at" db:"executed_at"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// ================================
// REQUEST/RESPONSE DTOs
// ================================

// CreateConnectAccountRequest represents request to start Stripe Connect onboarding
type CreateConnectAccountRequest struct {
	RefreshURL string `json:"refresh_url" binding:"required"` // Where to redirect if user leaves onboarding
	ReturnURL  string `json:"return_url" binding:"required"`  // Where to redirect after onboarding
}

// CreateConnectAccountResponse represents response with onboarding link
type CreateConnectAccountResponse struct {
	AccountID     string `json:"account_id"`
	OnboardingURL string `json:"onboarding_url"`
	Message       string `json:"message"`
}

// GetConnectAccountStatusResponse represents Connect account status
type GetConnectAccountStatusResponse struct {
	AccountID           string  `json:"account_id"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
	PayoutsEnabled      bool    `json:"payouts_enabled"`
	ChargesEnabled      bool    `json:"charges_enabled"`
	Balance             float64 `json:"balance"`
	TotalEarned         float64 `json:"total_earned"`
	TotalWithdrawn      float64 `json:"total_withdrawn"`
	CanWithdraw         bool    `json:"can_withdraw"`
	MinimumWithdrawal   float64 `json:"minimum_withdrawal"`
}

// CreateWithdrawalRequest represents request to withdraw funds
type CreateWithdrawalRequest struct {
	Amount float64 `json:"amount" binding:"required,min=50"`
}

// CreateWithdrawalResponse represents response after creating withdrawal
type CreateWithdrawalResponse struct {
	WithdrawalID string  `json:"withdrawal_id"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
	Message      string  `json:"message"`
	EstimatedArrival string `json:"estimated_arrival,omitempty"` // e.g., "2-3 business days"
}

// GetWalletBalanceResponse represents developer wallet balance
type GetWalletBalanceResponse struct {
	Balance           float64 `json:"balance"`
	TotalEarned       float64 `json:"total_earned"`
	TotalWithdrawn    float64 `json:"total_withdrawn"`
	PendingWithdrawals float64 `json:"pending_withdrawals"`
	CanWithdraw       bool    `json:"can_withdraw"`
	MinimumWithdrawal float64 `json:"minimum_withdrawal"`
}

// GetConnectedDevelopersResponse represents list of all connected developers
type GetConnectedDevelopersResponse struct {
	Developers []ConnectedDeveloperSummary `json:"developers"`
	Total      int                         `json:"total"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
}

// ConnectedDeveloperSummary represents a summary of a connected developer
type ConnectedDeveloperSummary struct {
	OrganizationID      string  `json:"organization_id"`
	StripeAccountID     *string `json:"stripe_account_id"`
	Balance             float64 `json:"balance"`
	TotalEarned         float64 `json:"total_earned"`
	TotalWithdrawn      float64 `json:"total_withdrawn"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
	PayoutsEnabled      bool    `json:"payouts_enabled"`
	ChargesEnabled      bool    `json:"charges_enabled"`
	JoinedAt            string  `json:"joined_at"`
}

// GetTransactionHistoryResponse represents transaction history
type GetTransactionHistoryResponse struct {
	Transactions []TransactionSummary `json:"transactions"`
	Total        int                  `json:"total"`
	Page         int                  `json:"page"`
	Limit        int                  `json:"limit"`
}

// TransactionSummary represents a summary of a transaction
type TransactionSummary struct {
	ID                 string    `json:"id"`
	FunctionID         string    `json:"function_id"`
	FunctionName       string    `json:"function_name"`
	UserOrganization   string    `json:"user_organization"`
	Amount             float64   `json:"amount"`
	PlatformFee        float64   `json:"platform_fee"`
	NetAmount          float64   `json:"net_amount"`
	Status             string    `json:"status"`
	ExecutedAt         time.Time `json:"executed_at"`
}

// GetWithdrawalHistoryResponse represents withdrawal history
type GetWithdrawalHistoryResponse struct {
	Withdrawals []WithdrawalSummary `json:"withdrawals"`
	Total       int                 `json:"total"`
	Page        int                 `json:"page"`
	Limit       int                 `json:"limit"`
}

// WithdrawalSummary represents a summary of a withdrawal
type WithdrawalSummary struct {
	ID              string     `json:"id"`
	Amount          float64    `json:"amount"`
	Status          string     `json:"status"`
	RequestedAt     time.Time  `json:"requested_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	FailureReason   *string    `json:"failure_reason"`
}

// FunctionExecutionPaymentRequest represents payment for function execution
type FunctionExecutionPaymentRequest struct {
	FunctionID              string  `json:"function_id" binding:"required"`
	Amount                  float64 `json:"amount" binding:"required,min=0.01"`
	DeveloperOrganizationID string  `json:"developer_organization_id"` // Optional for testing, in production lookup from functions table
}

// FunctionExecutionPaymentResponse represents response after function execution payment
type FunctionExecutionPaymentResponse struct {
	TransactionID   string  `json:"transaction_id"`
	Amount          float64 `json:"amount"`
	PlatformFee     float64 `json:"platform_fee"`
	NetAmount       float64 `json:"net_amount"`
	UserBalance     float64 `json:"user_balance"`
	DeveloperBalance float64 `json:"developer_balance"`
	Message         string  `json:"message"`
}

// ConnectAccountLink represents Stripe account link
type ConnectAccountLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Constants
const (
	MinimumWithdrawalAmount = 50.00
	DefaultPlatformFeePercent = 0.00 // Future: You can add platform commission (e.g., 10%)

	// Withdrawal statuses
	WithdrawalStatusPending    = "pending"
	WithdrawalStatusProcessing = "processing"
	WithdrawalStatusCompleted  = "completed"
	WithdrawalStatusFailed     = "failed"
	WithdrawalStatusRejected   = "rejected"

	// Transaction statuses
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusRefunded  = "refunded"
)
