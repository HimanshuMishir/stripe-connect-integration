# Integration Guide: Merging Stripe Connect into Rival API

This guide explains how to integrate the Stripe Connect Express implementation into your existing `/home/himanshu/Rival/rival-api` codebase.

## Overview

The Stripe Connect integration consists of:
- Database schema for wallets, withdrawals, and transactions
- Repository layer for data access
- Service layer for business logic
- API handlers and routes
- Webhook handling

## Step-by-Step Integration

### 1. Database Schema

#### Apply the SQL Schema

```bash
# Navigate to your main project
cd /home/himanshu/Rival/rival-api

# Apply the Stripe Connect schema
psql -U your_db_user -d rival -f /home/himanshu/Rival/strpe-connect/database/stripe_connect_schema.sql
```

**Alternative:** Append to your existing migration file:

```bash
cat /home/himanshu/Rival/strpe-connect/database/stripe_connect_schema.sql >> /home/himanshu/Rival/rival-api/database/alter.sql
```

### 2. Domain Models

Add Stripe Connect models to your domain package:

```bash
# Copy the models
cat /home/himanshu/Rival/strpe-connect/models/stripe_connect.go >> /home/himanshu/Rival/rival-api/internal/domain/stripe_connect.go
```

**Or manually integrate:**

Open `/home/himanshu/Rival/rival-api/internal/domain/models.go` and add:
- `DeveloperWallet`
- `WithdrawalRequest`
- `FunctionExecutionTransaction`
- All request/response DTOs
- Constants

### 3. Repository Layer

#### Create Stripe Connect Repository

```bash
# Create new file
cp /home/himanshu/Rival/strpe-connect/repository/stripe_connect_repository.go \
   /home/himanshu/Rival/rival-api/internal/repository/stripe_connect.go
```

#### Update the repository to match your project structure:

**File:** `internal/repository/stripe_connect.go`

```go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rival-api/internal/domain" // Update import path
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ... rest of the file
```

#### Update Interface Definitions

Add the `StripeConnectRepository` interface to `internal/domain/interfaces.go`:

```go
type StripeConnectRepository interface {
	CreateDeveloperWallet(ctx context.Context, organizationID string) (*DeveloperWallet, error)
	GetDeveloperWalletByOrgID(ctx context.Context, organizationID string) (*DeveloperWallet, error)
	UpdateStripeConnectAccountID(ctx context.Context, walletID, stripeAccountID string) error
	// ... add all other methods
}
```

### 4. Service Layer

#### Create Stripe Connect Service

```bash
cp /home/himanshu/Rival/strpe-connect/services/stripe_connect_service.go \
   /home/himanshu/Rival/rival-api/internal/services/stripe_connect.go
```

#### Update imports:

**File:** `internal/services/stripe_connect.go`

```go
package services

import (
	"context"
	"fmt"
	"log"
	"rival-api/internal/domain" // Update import
	"rival-api/internal/repository" // Update import
	"time"

	"github.com/stripe/stripe-go/v83"
	// ... other imports
)
```

#### Add Service Interface

Add to `internal/domain/interfaces.go`:

```go
type StripeConnectService interface {
	CreateConnectAccount(ctx context.Context, orgID, refreshURL, returnURL string) (*CreateConnectAccountResponse, error)
	GetConnectAccountStatus(ctx context.Context, orgID string) (*GetConnectAccountStatusResponse, error)
	// ... add all other methods
}
```

#### IMPORTANT: Fix Developer Organization Lookup

In the service file, locate the `ProcessFunctionExecutionPayment` method and update it to get the actual developer organization from your functions table:

```go
func (s *stripeConnectService) ProcessFunctionExecutionPayment(ctx context.Context, userOrgID, functionID string, amount float64) (*domain.FunctionExecutionPaymentResponse, error) {
	// ... existing code ...

	// TODO: Get developer org ID from function
	// Query the functions table to get the organization_id of the function owner

	// Add this query to your repository
	functionInfo, err := s.functionRepo.GetFunctionOwner(ctx, functionID)
	if err != nil {
		return nil, fmt.Errorf("function not found: %w", err)
	}

	developerOrgID := functionInfo.OrganizationID

	// ... rest of the method
}
```

You'll need to add a method to your function repository to get the owner:

```go
func (r *functionRepository) GetFunctionOwner(ctx context.Context, functionID string) (*FunctionOwnerInfo, error) {
	var info FunctionOwnerInfo
	query := `SELECT organization_id, function_name FROM tenant_schema.functions WHERE function_id = $1`
	err := r.db.QueryRow(ctx, query, functionID).Scan(&info.OrganizationID, &info.FunctionName)
	return &info, err
}
```

### 5. API Handlers

#### Create Handler File

```bash
cp /home/himanshu/Rival/strpe-connect/handlers/stripe_connect_handler.go \
   /home/himanshu/Rival/rival-api/internal/api/handlers/stripe_connect.go
```

#### Update imports and Application struct:

**File:** `internal/api/handlers/stripe_connect.go`

```go
package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"rival-api/internal/domain" // Update imports
	"rival-api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v83/webhook"
)
```

#### Update Application Struct

**File:** `internal/api/handlers/api.go`

Add the Stripe Connect service and webhook secret:

```go
type Application struct {
	// ... existing fields
	StripeConnectService   domain.StripeConnectService
	StripeWebhookSecret    string
	// ... existing fields
}
```

#### Update main.go

**File:** `cmd/api/main.go`

Initialize the new service and pass it to the application:

```go
// Add after paymentService initialization
stripeConnectRepo := repository.NewStripeConnectRepository(db)
stripeConnectService := services.NewStripeConnectService(
	stripeConnectRepo,
	cfg.Stripe.SecretKey,
)

// Update application initialization
application := handlers.NewApplication(
	// ... existing services
	stripeConnectService,
	cfg.Stripe.WebhookSecret,
	// ... existing fields
)
```

### 6. Routes

#### Create Routes File

**File:** `internal/api/routes/stripe_connect.go`

```go
package routes

import (
	"rival-api/internal/api/handlers"
	"rival-api/internal/config"

	"github.com/gin-gonic/gin"
)

func RegisterStripeConnectRoutes(r *gin.Engine, app *handlers.Application, cfg *config.Config) {
	connectHandler := handlers.NewStripeConnectHandler(
		app.StripeConnectService,
		app.StripeWebhookSecret,
	)

	api := r.Group("/api")
	{
		connect := api.Group("/connect")
		{
			// Onboarding
			connect.POST("/onboard", connectHandler.CreateConnectAccount)
			connect.GET("/status", connectHandler.GetConnectAccountStatus)
			connect.POST("/refresh-onboarding", connectHandler.RefreshOnboardingLink)

			// Wallet
			wallet := connect.Group("/wallet")
			{
				wallet.GET("/balance", connectHandler.GetWalletBalance)
				wallet.GET("/transactions", connectHandler.GetTransactionHistory)
			}

			// Withdrawals
			withdrawals := connect.Group("/withdrawals")
			{
				withdrawals.POST("/request", connectHandler.RequestWithdrawal)
				withdrawals.GET("/history", connectHandler.GetWithdrawalHistory)
			}

			// Payments
			payments := connect.Group("/payments")
			{
				payments.POST("/execute", connectHandler.ProcessFunctionPayment)
			}
		}

		// Webhooks
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/stripe-connect", connectHandler.HandleWebhook)
		}
	}
}
```

#### Register Routes in main.go

**File:** `cmd/api/main.go`

```go
// Add after other route registrations
routes.RegisterStripeConnectRoutes(r, application, cfg)
```

### 7. Configuration

#### Update Config Struct

**File:** `internal/config/config.go`

The `StripeConfig` already exists in your config, but ensure it has the webhook secret:

```go
type StripeConfig struct {
	SecretKey     string `mapstructure:"secret_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
}
```

Make sure the environment binding exists:

```go
v.BindEnv("stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET")
v.BindEnv("stripe.secret_key", "STRIPE_SECRET_KEY")
```

### 8. Update Dependencies

#### Add Required Packages

**File:** `go.mod` (in your main project)

The Stripe package version might need updating:

```bash
cd /home/himanshu/Rival/rival-api
go get github.com/stripe/stripe-go/v83
go mod tidy
```

### 9. Environment Variables

Update your `.env` file:

```bash
# Stripe Connect
STRIPE_SECRET_KEY=sk_test_your_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
```

## Integration with Function Execution

To integrate payments into your existing function execution flow, update your function execution handler:

**File:** `internal/api/handlers/function.go` (or wherever you handle function execution)

```go
func (app *Application) ExecuteFunction(c *gin.Context) {
	// ... existing execution logic ...

	// Get function price
	functionPrice := functionDetails.PricePerAPIRequest

	// Process payment
	orgID := c.GetHeader("X-Organization-ID")
	paymentResp, err := app.StripeConnectService.ProcessFunctionExecutionPayment(
		c.Request.Context(),
		orgID,
		functionID,
		functionPrice,
	)
	if err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error": "Payment failed: " + err.Error(),
		})
		return
	}

	// Continue with function execution
	// ... existing code ...

	c.JSON(http.StatusOK, gin.H{
		"result": executionResult,
		"payment": paymentResp,
	})
}
```

## Testing the Integration

### 1. Start the Server

```bash
cd /home/himanshu/Rival/rival-api
go run cmd/api/main.go
```

### 2. Test Developer Onboarding

```bash
curl -X POST http://localhost:8080/api/connect/onboard \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: test-org-123" \
  -d '{
    "refresh_url": "http://localhost:3000/connect/refresh",
    "return_url": "http://localhost:3000/connect/complete"
  }'
```

### 3. Test Payment Flow

```bash
curl -X POST http://localhost:8080/api/connect/payments/execute \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: user-org-id" \
  -d '{
    "function_id": "your-function-id",
    "amount": 5.00
  }'
```

### 4. Setup Webhook Forwarding

```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

## Common Issues and Solutions

### Issue: "wallet not found"
**Solution:** The organization doesn't have a developer wallet yet. Call the onboarding endpoint first.

### Issue: "insufficient balance"
**Solution:** The user's account balance is too low. They need to top up using your existing payment flow.

### Issue: Developer org ID is "placeholder"
**Solution:** Update the service to query the actual function owner from the database (see step 4).

### Issue: Webhook signature verification failed
**Solution:** Make sure `STRIPE_WEBHOOK_SECRET` matches the secret from Stripe CLI or Stripe Dashboard.

## Production Checklist

Before deploying to production:

- [ ] Replace test Stripe keys with live keys
- [ ] Set up production webhook endpoint in Stripe Dashboard
- [ ] Configure proper database backups
- [ ] Add monitoring and alerting for failed withdrawals
- [ ] Review and adjust platform fee percentage
- [ ] Add rate limiting on withdrawal requests
- [ ] Test the complete flow in staging environment
- [ ] Document the process for your team

## Next Steps

1. **Test thoroughly** - Use the React frontend to test all features
2. **Add monitoring** - Track wallet balances, withdrawal success rates
3. **Implement notifications** - Email developers when withdrawals complete
4. **Add admin dashboard** - View all transactions and withdrawals
5. **Optimize queries** - Add database indexes for high-traffic queries

## Support

If you encounter any issues during integration:
1. Check the logs for detailed error messages
2. Verify all environment variables are set correctly
3. Ensure database schema was applied successfully
4. Test each component individually before testing the full flow
