# Quick Start Guide

Get the Stripe Connect Express marketplace running in 10 minutes!

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Node.js 18+ (for React frontend)
- Stripe account (test mode)

## Step 1: Clone & Setup (2 minutes)

```bash
# Navigate to the project
cd /home/himanshu/Rival/strpe-connect

# Copy environment file
cp .env.example .env
```

## Step 2: Configure Environment (2 minutes)

Edit `.env` file:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=rival

# Get your Stripe test key from https://dashboard.stripe.com/test/apikeys
STRIPE_SECRET_KEY=sk_test_paste_your_key_here

# Leave this blank for now - we'll get it in Step 4
STRIPE_WEBHOOK_SECRET=

# Server
PORT=8080
```

## Step 3: Setup Database (2 minutes)

```bash
# Connect to PostgreSQL
psql -U postgres -d rival

# Run the schema (from psql prompt)
\i database/stripe_connect_schema.sql

# Or if using command line:
psql -U postgres -d rival -f database/stripe_connect_schema.sql

# Verify tables were created
\dt tenant_schema.developer_wallets
\dt tenant_schema.withdrawal_requests
\dt tenant_schema.function_execution_transactions
```

## Step 4: Setup Stripe Webhooks (2 minutes)

### Install Stripe CLI

```bash
# Mac
brew install stripe/stripe-cli/stripe

# Linux
curl -s https://packages.stripe.com/api/v1/keys/gpg | sudo apt-key add -
sudo apt-get install stripe
```

### Login and Forward Webhooks

```bash
# Login to Stripe
stripe login

# Forward webhooks to local server
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

You'll see output like:
```
> Ready! Your webhook signing secret is whsec_xxxxxxxxxxxxx
```

**Copy that secret** and add it to your `.env` file:
```
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
```

## Step 5: Run the Backend (1 minute)

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

You should see:
```
✅ Database connected successfully
🚀 Server starting on http://localhost:8080
```

## Step 6: Run the Frontend (1 minute)

Open a **new terminal**:

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

You should see:
```
VITE ready in 500ms
Local: http://localhost:3000
```

## Step 7: Test the Flow (2 minutes)

### Open the Frontend

Visit http://localhost:3000 in your browser.

### Test Developer Onboarding

1. Click on "👨‍💻 Developer Dashboard" tab
2. Click "Start Onboarding" button
3. A new tab will open with Stripe Connect onboarding
4. Fill in test information:
   - Use test mode (should be pre-selected)
   - Business name: "Test Developer"
   - Any test information
5. Complete the onboarding
6. Return to your app and refresh
7. You should see "Onboarding Complete" status

### Test Function Execution Payment

First, add test data to your database:

```sql
-- Add a test user organization and account
INSERT INTO tenant_schema.organizations (id, name)
VALUES ('test-user-org', 'Test User Org');

INSERT INTO tenant_schema.accounts (id, organization_id, account_balance)
VALUES (uuid_generate_v4(), 'test-user-org', 100.00);
```

Now test the payment:

1. Click on "👤 User Dashboard" tab
2. Change the Organization ID to: `test-user-org`
3. Enter amount: `5.00`
4. Click "Execute Function & Pay $5.00"
5. You should see success message with transaction details

### Check Developer Balance

1. Switch back to "👨‍💻 Developer Dashboard" tab
2. Refresh the page
3. You should see:
   - Current Balance: $5.00
   - Total Earned: $5.00
   - Recent transaction showing the payment

### Test Withdrawal

1. In Developer Dashboard, scroll to "Request Withdrawal"
2. Enter amount: `50.00` (or more if you have enough)
3. Click "Withdraw"
4. Check the Stripe CLI terminal - you should see webhook events
5. Money will be transferred to the test bank account

## Troubleshooting

### "wallet not found"
- Make sure you completed the onboarding first
- Check that organization ID is correct

### "insufficient balance"
- Add more balance to the test account:
```sql
UPDATE tenant_schema.accounts
SET account_balance = 100.00
WHERE organization_id = 'test-user-org';
```

### "failed to connect to database"
- Verify PostgreSQL is running: `pg_isready`
- Check credentials in `.env`
- Ensure database exists: `psql -l | grep rival`

### Webhook errors
- Make sure Stripe CLI is running
- Verify webhook secret in `.env` matches CLI output
- Check that server is running on port 8080

## Next Steps

Now that everything is working:

1. **Read the full README.md** for detailed API documentation
2. **Review INTEGRATION_GUIDE.md** to merge into your main codebase
3. **Test different scenarios**:
   - Multiple function executions
   - Withdrawals below minimum ($50)
   - Onboarding multiple developers
4. **Explore the code** to understand how it works
5. **Customize** for your specific needs

## API Testing with cURL

### Create Connect Account
```bash
curl -X POST http://localhost:8080/api/connect/onboard \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: test-org-123" \
  -d '{
    "refresh_url": "http://localhost:3000/refresh",
    "return_url": "http://localhost:3000/complete"
  }'
```

### Get Wallet Balance
```bash
curl -X GET http://localhost:8080/api/connect/wallet/balance \
  -H "X-Organization-ID: test-org-123"
```

### Request Withdrawal
```bash
curl -X POST http://localhost:8080/api/connect/withdrawals/request \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: test-org-123" \
  -d '{"amount": 50.00}'
```

### Execute Function Payment
```bash
curl -X POST http://localhost:8080/api/connect/payments/execute \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: test-user-org" \
  -d '{
    "function_id": "func-test-123",
    "amount": 5.00
  }'
```

## Database Queries for Testing

### Check developer wallet
```sql
SELECT * FROM tenant_schema.developer_wallets;
```

### View all transactions
```sql
SELECT * FROM tenant_schema.function_execution_transactions
ORDER BY executed_at DESC;
```

### View withdrawal history
```sql
SELECT * FROM tenant_schema.withdrawal_requests
ORDER BY requested_at DESC;
```

### Add test balance to user
```sql
UPDATE tenant_schema.accounts
SET account_balance = account_balance + 100.00
WHERE organization_id = 'test-user-org';
```

## Help & Support

- Check server logs for error details
- Verify all environment variables are set
- Ensure Stripe CLI is running and forwarding webhooks
- Review the full README.md for comprehensive documentation
- Check INTEGRATION_GUIDE.md for merging into your main codebase

Happy testing!
