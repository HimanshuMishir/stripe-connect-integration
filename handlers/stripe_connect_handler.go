package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strpe-connect/models"
	"strpe-connect/services"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v83/webhook"
)

type StripeConnectHandler struct {
	service       services.StripeConnectService
	webhookSecret string
}

func NewStripeConnectHandler(service services.StripeConnectService, webhookSecret string) *StripeConnectHandler {
	return &StripeConnectHandler{
		service:       service,
		webhookSecret: webhookSecret,
	}
}

// ================================
// ONBOARDING ENDPOINTS
// ================================

// CreateConnectAccount godoc
// @Summary Create Stripe Connect account for developer
// @Description Initiates Stripe Connect Express onboarding for a developer organization
// @Tags Stripe Connect
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param request body models.CreateConnectAccountRequest true "Onboarding URLs"
// @Success 200 {object} models.CreateConnectAccountResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/connect/onboard [post]
func (h *StripeConnectHandler) CreateConnectAccount(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	var req models.CreateConnectAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateConnectAccount(c.Request.Context(), orgID, req.RefreshURL, req.ReturnURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetConnectAccountStatus godoc
// @Summary Get Connect account status
// @Description Retrieves the status of developer's Stripe Connect account and wallet
// @Tags Stripe Connect
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Success 200 {object} models.GetConnectAccountStatusResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/connect/status [get]
func (h *StripeConnectHandler) GetConnectAccountStatus(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	resp, err := h.service.GetConnectAccountStatus(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RefreshOnboardingLink godoc
// @Summary Refresh onboarding link
// @Description Generates a new onboarding link for incomplete onboarding
// @Tags Stripe Connect
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param request body models.CreateConnectAccountRequest true "Onboarding URLs"
// @Success 200 {object} models.CreateConnectAccountResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/refresh-onboarding [post]
func (h *StripeConnectHandler) RefreshOnboardingLink(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	var req models.CreateConnectAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.RefreshOnboardingLink(c.Request.Context(), orgID, req.RefreshURL, req.ReturnURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ================================
// WALLET ENDPOINTS
// ================================

// GetWalletBalance godoc
// @Summary Get wallet balance
// @Description Retrieves developer's wallet balance and earnings information
// @Tags Wallet
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Success 200 {object} models.GetWalletBalanceResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/wallet/balance [get]
func (h *StripeConnectHandler) GetWalletBalance(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	resp, err := h.service.GetWalletBalance(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTransactionHistory godoc
// @Summary Get transaction history
// @Description Retrieves developer's function execution transaction history
// @Tags Wallet
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} models.GetTransactionHistoryResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/wallet/transactions [get]
func (h *StripeConnectHandler) GetTransactionHistory(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	resp, err := h.service.GetTransactionHistory(c.Request.Context(), orgID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetConnectedDevelopers godoc
// @Summary Get list of all connected developers
// @Description Returns a list of all developers who have created wallets (onboarded or in progress)
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Results per page" default(50)
// @Success 200 {object} models.GetConnectedDevelopersResponse
// @Router /api/admin/connected-developers [get]
func (h *StripeConnectHandler) GetConnectedDevelopers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	resp, err := h.service.GetConnectedDevelopers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetConnectedDevelopersForOrg godoc
// @Summary Get list of connected developers for a specific user organization
// @Description Returns developers that this user organization has paid
// @Tags Connect
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "User Organization ID"
// @Success 200 {object} models.GetConnectedDevelopersResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/connected-developers [get]
func (h *StripeConnectHandler) GetConnectedDevelopersForOrg(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	resp, err := h.service.GetConnectedDevelopersForOrg(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ================================
// WITHDRAWAL ENDPOINTS
// ================================

// RequestWithdrawal godoc
// @Summary Request withdrawal
// @Description Creates a withdrawal request to transfer earnings to bank account
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param request body models.CreateWithdrawalRequest true "Withdrawal amount"
// @Success 200 {object} models.CreateWithdrawalResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/withdrawals/request [post]
func (h *StripeConnectHandler) RequestWithdrawal(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	var req models.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.RequestWithdrawal(c.Request.Context(), orgID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetWithdrawalHistory godoc
// @Summary Get withdrawal history
// @Description Retrieves developer's withdrawal history
// @Tags Withdrawals
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} models.GetWithdrawalHistoryResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/withdrawals/history [get]
func (h *StripeConnectHandler) GetWithdrawalHistory(c *gin.Context) {
	orgID := c.GetHeader("X-Organization-ID")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	resp, err := h.service.GetWithdrawalHistory(c.Request.Context(), orgID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ================================
// FUNCTION EXECUTION PAYMENT
// ================================

// ProcessFunctionPayment godoc
// @Summary Process function execution payment
// @Description Deducts from user balance and credits developer wallet
// @Tags Payments
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "User Organization ID"
// @Param request body models.FunctionExecutionPaymentRequest true "Payment details"
// @Success 200 {object} models.FunctionExecutionPaymentResponse
// @Failure 400 {object} map[string]string
// @Router /api/connect/payments/execute [post]
func (h *StripeConnectHandler) ProcessFunctionPayment(c *gin.Context) {
	userOrgID := c.GetHeader("X-Organization-ID")
	if userOrgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	var req models.FunctionExecutionPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In production, developer org ID would be looked up from functions table
	if req.DeveloperOrganizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "developer_organization_id is required"})
		return
	}

	resp, err := h.service.ProcessFunctionExecutionPayment(c.Request.Context(), userOrgID, req.FunctionID, req.DeveloperOrganizationID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ================================
// WEBHOOK ENDPOINT
// ================================

// HandleWebhook godoc
// @Summary Handle Stripe webhook events
// @Description Receives and processes Stripe webhook events for Connect accounts
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/webhooks/stripe-connect [post]
func (h *StripeConnectHandler) HandleWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing stripe-signature header"})
		return
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, h.webhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	log.Printf("Received webhook event: %s", event.Type)

	// Handle different event types
	switch event.Type {
	case "account.updated":
		if err := h.handleAccountUpdated(c, event.Data.Raw); err != nil {
			log.Printf("Error handling account.updated: %v", err)
		}

	case "payout.paid":
		log.Printf("Payout paid event received")
		if err := h.handlePayoutPaid(c, event.Data.Raw); err != nil {
			log.Printf("Error handling payout.paid: %v", err)
		}

	case "payout.failed":
		log.Printf("Payout failed event received")
		if err := h.handlePayoutFailed(c, event.Data.Raw); err != nil {
			log.Printf("Error handling payout.failed: %v", err)
		}

	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *StripeConnectHandler) handleAccountUpdated(c *gin.Context, rawData json.RawMessage) error {
	var account struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(rawData, &account); err != nil {
		return err
	}

	return h.service.HandleAccountUpdated(c.Request.Context(), account.ID)
}

func (h *StripeConnectHandler) handlePayoutPaid(c *gin.Context, rawData json.RawMessage) error {
	var payout struct {
		ID       string            `json:"id"`
		Metadata map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(rawData, &payout); err != nil {
		return err
	}

	withdrawalID := payout.Metadata["withdrawal_id"]
	if withdrawalID == "" {
		log.Printf("WARNING: payout.paid event missing withdrawal_id in metadata")
		return nil
	}

	return h.service.HandlePayoutPaid(c.Request.Context(), withdrawalID, payout.ID)
}

func (h *StripeConnectHandler) handlePayoutFailed(c *gin.Context, rawData json.RawMessage) error {
	var payout struct {
		ID            string            `json:"id"`
		Metadata      map[string]string `json:"metadata"`
		FailureCode   string            `json:"failure_code"`
		FailureMessage string           `json:"failure_message"`
	}

	if err := json.Unmarshal(rawData, &payout); err != nil {
		return err
	}

	withdrawalID := payout.Metadata["withdrawal_id"]
	if withdrawalID == "" {
		log.Printf("WARNING: payout.failed event missing withdrawal_id in metadata")
		return nil
	}

	failureReason := fmt.Sprintf("%s: %s", payout.FailureCode, payout.FailureMessage)
	return h.service.HandlePayoutFailed(c.Request.Context(), withdrawalID, failureReason)
}
