# Stripe Connect Express - Function Marketplace

A complete implementation of Stripe Connect Express for a Function Marketplace where developers sell function executions.

## Features

### For Developers
- **Stripe Connect Onboarding**: Easy Express account setup
- **Wallet Management**: Track earnings in real-time
- **Automatic Payments**: Receive payments when users execute your functions
- **Withdrawals**: Withdraw earnings to bank account (minimum $50)
- **Transaction History**: View all function execution payments

### For Users
- **Wallet Top-up**: Add funds to account balance
- **Function Execution**: Pay developers when executing their functions
- **Transaction Tracking**: View all payment history

### Technical Features
- PostgreSQL database with complete schema
- RESTful API with Gin framework
- Stripe Connect Express integration
- Webhook support for real-time updates
- React frontend for testing

## Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│    User     │────────▶│  Your Platform   │────────▶│  Developer  │
│  (Buyer)    │         │  (Marketplace)   │         │  (Seller)   │
└─────────────┘         └──────────────────┘         └─────────────┘
     │                          │                           │
     │ 1. Top up wallet         │                           │
     ▼                          │                           │
┌─────────────┐                │                           │
│   Stripe    │                │                           │
│  (Payment)  │                │                           │
└─────────────┘                │                           │
                                │                           │
     2. Execute Function        │                           │
     ────────────────────────▶  │                           │
                                │  3. Deduct from user      │
                                │  4. Credit to developer   │
                                │  ─────────────────────▶   │
                                │                           │
                                │    5. Request withdrawal  │
                                │  ◀─────────────────────   │
                                │                           │
                                ▼                           ▼
                         ┌──────────────┐         ┌─────────────┐
                         │ User Account │         │ Dev Wallet  │
                         │  (Balance)   │         │  (Earnings) │
                         └──────────────┘         └─────────────┘
```

## Database Schema

### Tables

1. **developer_wallets** - Track developer earnings
   - Stripe Connect account ID
   - Balance, total earned, total withdrawn
   - Onboarding status

2. **withdrawal_requests** - Manage withdrawal requests
   - Amount (minimum $50)
   - Status (pending, processing, completed, failed)
   - Stripe payout ID

3. **function_execution_transactions** - Record all payments
   - User → Developer transfers
   - Platform fees
   - Transaction history

## API Endpoints

### Stripe Connect Onboarding
```http
POST   /api/connect/onboard            # Create Connect account & get onboarding link
GET    /api/connect/status             # Get account status
POST   /api/connect/refresh-onboarding # Refresh onboarding link
```

### Wallet Management
```http
GET    /api/connect/wallet/balance     # Get wallet balance
GET    /api/connect/wallet/transactions # Get transaction history
```

### Withdrawals
```http
POST   /api/connect/withdrawals/request # Request withdrawal
GET    /api/connect/withdrawals/history # Get withdrawal history
```

### Payments
```http
POST   /api/connect/payments/execute    # Process function execution payment
```

### Webhooks
```http
POST   /api/webhooks/stripe-connect    # Handle Stripe webhooks
```

## Setup Instructions

### 1. Database Setup

```bash
# Connect to your PostgreSQL database
psql -U postgres -d rival

# Run the schema
\i database/stripe_connect_schema.sql
```

### 2. Environment Configuration

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your configuration
nano .env
```

Required environment variables:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `STRIPE_SECRET_KEY` - Your Stripe secret key (get from https://dashboard.stripe.com)
- `STRIPE_WEBHOOK_SECRET` - Webhook signing secret
- `PORT` - Server port (default: 8080)

### 3. Get Stripe API Keys

1. Go to https://dashboard.stripe.com/test/apikeys
2. Copy your **Secret key** (starts with `sk_test_`)
3. Save it to `STRIPE_SECRET_KEY` in `.env`

### 4. Setup Stripe Webhooks

1. Install Stripe CLI: https://stripe.com/docs/stripe-cli
2. Login to Stripe CLI:
   ```bash
   stripe login
   ```
3. Forward webhooks to your local server:
   ```bash
   stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
   ```
4. Copy the webhook signing secret (starts with `whsec_`)
5. Save it to `STRIPE_WEBHOOK_SECRET` in `.env`

### 5. Run the Application

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

The server will start on http://localhost:8080

## Usage Flow

### Developer Onboarding

1. **Create Connect Account**
```bash
curl -X POST http://localhost:8080/api/connect/onboard \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: your-org-id" \
  -d '{
    "refresh_url": "http://localhost:3000/connect/refresh",
    "return_url": "http://localhost:3000/connect/complete"
  }'
```

Response:
```json
{
  "account_id": "acct_xxxxx",
  "onboarding_url": "https://connect.stripe.com/setup/...",
  "message": "Please complete Stripe Connect onboarding to receive payments"
}
```

2. **Check Onboarding Status**
```bash
curl -X GET http://localhost:8080/api/connect/status \
  -H "X-Organization-ID: your-org-id"
```

### User Executes Function

```bash
curl -X POST http://localhost:8080/api/connect/payments/execute \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: user-org-id" \
  -d '{
    "function_id": "func-123",
    "amount": 5.00
  }'
```

Response:
```json
{
  "transaction_id": "txn-xxx",
  "amount": 5.00,
  "platform_fee": 0.00,
  "net_amount": 5.00,
  "user_balance": 95.00,
  "developer_balance": 5.00,
  "message": "Payment processed successfully"
}
```

### Developer Withdraws Earnings

```bash
curl -X POST http://localhost:8080/api/connect/withdrawals/request \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: developer-org-id" \
  -d '{
    "amount": 50.00
  }'
```

Response:
```json
{
  "withdrawal_id": "wd-xxx",
  "amount": 50.00,
  "status": "processing",
  "message": "Withdrawal request created successfully",
  "estimated_arrival": "2-3 business days"
}
```

## Testing with React Frontend

A React testing interface is included. To run it:

```bash
cd frontend
npm install
npm run dev
```

Visit http://localhost:3000 to access the testing dashboard.

## Important Notes

### Minimum Withdrawal Amount
- Developers can only withdraw when they have **$50 or more**
- This is enforced at both database and application level

### Platform Fees
- Currently set to 0%
- Can be configured in `models/stripe_connect.go`
- Change `DefaultPlatformFeePercent` to add commission (e.g., 10.0 for 10%)

### Stripe Connect Express
- Uses **Express** accounts (easiest for developers)
- Stripe handles KYC, compliance, payouts
- Developers get paid directly to their bank account

### Webhooks
- **account.updated**: Updates onboarding status
- **payout.paid**: Confirms successful payout
- **payout.failed**: Handles payout failures

## Production Deployment

Before deploying to production:

1. **Use production Stripe keys**
   ```bash
   STRIPE_SECRET_KEY=sk_live_xxxxx
   ```

2. **Set up webhook endpoint**
   - Add webhook endpoint in Stripe Dashboard
   - Use your production URL: `https://yourapi.com/api/webhooks/stripe-connect`
   - Select events: `account.updated`, `payout.paid`, `payout.failed`

3. **Update CORS settings**
   - Update `AllowOrigins` in `main.go` to your production frontend URL

4. **Database**
   - Use managed PostgreSQL (AWS RDS, Google Cloud SQL, etc.)
   - Enable SSL connections
   - Regular backups

5. **Security**
   - Use environment variables for all secrets
   - Enable HTTPS
   - Validate all webhook signatures

## Integration with Main Codebase

To integrate this into your existing `/home/himanshu/Rival/rival-api`:

1. **Copy database schema**
   ```bash
   cat database/stripe_connect_schema.sql >> /home/himanshu/Rival/rival-api/database/alter.sql
   ```

2. **Copy Go files**
   - Move `models/stripe_connect.go` to `internal/domain/`
   - Move `repository/stripe_connect_repository.go` to `internal/repository/`
   - Move `services/stripe_connect_service.go` to `internal/services/`
   - Move `handlers/stripe_connect_handler.go` to `internal/api/handlers/`

3. **Update main.go**
   - Initialize repository, service, and handler
   - Register routes

4. **Update config**
   - Add Stripe webhook secret to config

See `INTEGRATION_GUIDE.md` for detailed steps.

## Troubleshooting

### "wallet not found"
- Developer hasn't started onboarding yet
- Call `/api/connect/onboard` first

### "insufficient balance"
- User doesn't have enough funds
- User needs to top up their wallet first

### "minimum withdrawal amount is $50"
- Developer balance is below $50
- Wait for more function executions

### Webhook signature verification failed
- Check `STRIPE_WEBHOOK_SECRET` is correct
- Make sure Stripe CLI is forwarding to correct URL

## License

MIT

## Support

For issues or questions, please open an issue on GitHub.
# stripe-connect-integration
