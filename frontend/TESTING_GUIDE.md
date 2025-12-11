# Frontend Testing Guide

## Setup

1. **Start the API server**:
```bash
cd /home/himanshu/Rival/rival-api
go run cmd/api/main.go
```

2. **Start the frontend**:
```bash
cd /home/himanshu/Rival/strpe-connect/frontend
npm install  # if not already done
npm run dev
```

3. **Open in browser**: http://localhost:5173

---

## Getting a JWT Token

You need a valid JWT token from the Rival API to use the dashboard. Here are several ways to get one:

### Option 1: Using an Existing User (Recommended)

If you have an existing user account in the Rival API:

```bash
# Replace with your actual login endpoint and credentials
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-email@example.com",
    "password": "your-password"
  }'
```

The response will contain a JWT token. Copy the token value.

### Option 2: Create a Test User (If needed)

If you don't have a user account yet:

```bash
# Create a new user
curl -X POST "http://localhost:8080/api/v1/users/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "name": "Test User",
    "organization_name": "Test Org"
  }'

# Then login to get the token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!"
  }'
```

### Option 3: Use Descope Test Token (If using Descope)

If your API uses Descope for authentication, you can get a token from the Descope console:

1. Go to https://app.descope.com
2. Login to your project
3. Go to "Authentication" → "Testing"
4. Generate a test token for a user
5. Copy the token

---

## Using the Dashboard

### 1. Enter JWT Token

Once you have the frontend open:

1. Paste your JWT token in the "JWT Token" field at the top
2. It will be saved in localStorage automatically
3. The token will persist across page refreshes

### 2. Set Organization ID

Enter the organization ID you want to test with:

- For **Developer Dashboard**: Use your developer organization ID
- For **User Dashboard**: Use your user organization ID

You can find organization IDs in your database:

```sql
SELECT id, name FROM tenant_schema.organizations;
```

---

## Testing Flow

### A. Developer Onboarding

1. Switch to **Developer Dashboard** tab
2. Click **"Start Onboarding"**
3. Complete Stripe Connect onboarding in the new tab
4. Return to the dashboard and click **"Refresh Status"**
5. Verify that "Onboarding Complete" shows ✅

### B. Function Execution Payment

1. Make sure the developer has completed onboarding (Step A)
2. Switch to **User Dashboard** tab
3. Enter:
   - **Function ID**: Any test ID (e.g., `func-test-123`)
   - **Function Version**: e.g., `v1.0.0`
   - **Developer Org ID**: The developer's organization ID
   - **Amount**: e.g., `10.00`
4. Click **"Execute Function & Pay"**
5. Verify transaction details are displayed

**Note**: Your user account must have sufficient balance. If not, add balance first:

```bash
# Add $100 to user account
curl -X POST "http://localhost:8080/api/v1/payments/charge" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Organization-ID: YOUR_USER_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100.00}'
```

### C. Developer Withdrawal

1. Switch back to **Developer Dashboard**
2. Accumulate at least $50 in balance (execute multiple payments)
3. Enter withdrawal amount (minimum $50)
4. Click **"Withdraw $XX.XX"**
5. Check **Withdrawal History** section for status

---

## Troubleshooting

### Error: "JWT token is required"
- Make sure you entered a valid JWT token in the form at the top
- Check that the token hasn't expired
- Verify the token format starts with `ey...`

### Error: "401 Unauthorized"
- Your JWT token may have expired - get a new one
- Verify the token is valid for the API you're calling
- Check that the API server is running

### Error: "wallet not found"
- The developer needs to complete onboarding first
- Click "Start Onboarding" in Developer Dashboard

### Error: "insufficient balance"
- User account needs more funds
- Add balance using the payment charge endpoint (see above)

### Error: "function not found"
- The function_id doesn't exist in the database
- For testing, any ID works as long as `developer_organization_id` is provided
- In production, the API will look up the developer from the functions table

---

## Database Verification

You can verify the data in your database:

```sql
-- Check developer wallet
SELECT * FROM tenant_schema.developer_wallets
WHERE organization_id = 'your-org-id';

-- Check transactions
SELECT * FROM tenant_schema.function_execution_transactions
ORDER BY executed_at DESC LIMIT 10;

-- Check withdrawals
SELECT * FROM tenant_schema.withdrawal_requests
ORDER BY requested_at DESC LIMIT 10;

-- Check user balance
SELECT organization_id, account_balance
FROM tenant_schema.accounts
WHERE organization_id = 'your-user-org-id';
```

---

## Quick Test Script

Here's a complete test script you can use:

```bash
# 1. Get JWT token (replace with your login)
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}' \
  | jq -r '.token')

echo "JWT Token: $TOKEN"

# 2. Set variables
USER_ORG_ID="your-user-org-id"
DEV_ORG_ID="your-dev-org-id"

# 3. Start developer onboarding
curl -X POST "http://localhost:8080/api/v1/connect/onboard" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $DEV_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_url": "http://localhost:5173/refresh",
    "return_url": "http://localhost:5173/complete"
  }'

# 4. Add balance to user (complete Stripe payment in browser)
curl -X POST "http://localhost:8080/api/v1/payments/charge" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $USER_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100.00}'

# 5. Execute function and pay developer
curl -X POST "http://localhost:8080/api/v1/connect/payments/execute" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $USER_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "func-test-123",
    "version": "v1.0.0",
    "amount": 10.00,
    "developer_organization_id": "'"$DEV_ORG_ID"'"
  }'

# 6. Check developer balance
curl -X GET "http://localhost:8080/api/v1/connect/wallet/balance" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $DEV_ORG_ID"

# 7. Request withdrawal (after accumulating $50+)
curl -X POST "http://localhost:8080/api/v1/connect/withdrawals/request" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $DEV_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{"amount": 50.00}'
```

---

## Notes

- **JWT Token Storage**: The token is stored in localStorage and persists across page refreshes
- **API URL**: The frontend is configured to call `http://localhost:8080/api/v1/connect`
- **CORS**: Make sure your API allows requests from `http://localhost:5173`
- **Stripe Test Mode**: Use Stripe test API keys for development

---

## What's Changed from Demo

1. ✅ **Added JWT Authentication**: All requests now include `Authorization: Bearer <token>` header
2. ✅ **Fixed URLs**: Changed from relative URLs to full `http://localhost:8080/api/v1/connect`
3. ✅ **Added Version Field**: Function execution now requires version (e.g., `v1.0.0`)
4. ✅ **Token Management**: JWT token is saved in localStorage
5. ✅ **Better Error Handling**: Shows when token is missing or invalid

---

## Success Criteria

You'll know it's working when:

- ✅ Developer can complete Stripe Connect onboarding
- ✅ Developer dashboard shows "Onboarding Complete"
- ✅ User can execute function and payment goes through
- ✅ Transaction appears in developer's transaction history
- ✅ Developer balance increases
- ✅ User balance decreases
- ✅ Developer can request withdrawal when balance >= $50
- ✅ Withdrawal shows in history with "completed" status
