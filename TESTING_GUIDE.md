# 🧪 Complete Testing Guide

## ✅ Fixed! Now Run This:

The UUID error has been fixed. Run the setup again:

```bash
cd /home/himanshu/Rival/strpe-connect
./setup-complete.sh
```

This time it will create proper test organizations with these IDs:

```
👤 User Organization (has $500 to spend)
   ID: 11111111-1111-1111-1111-111111111111

👨‍💻 Developer Organization (will receive payments)
   ID: 22222222-2222-2222-2222-222222222222
```

---

## 🚀 Complete Testing Flow

### Step 1: Run Setup Script

```bash
./setup-complete.sh
```

Expected output:
```
✅ Database connection successful
✅ Schema applied successfully
✅ Test data added
```

---

### Step 2: Install Stripe CLI

**Mac:**
```bash
brew install stripe/stripe-cli/stripe
```

**Linux:**
```bash
curl -s https://packages.stripe.com/api/v1/keys/gpg | sudo apt-key add -
echo "deb https://packages.stripe.com/stripe-cli-debian-local stable main" | sudo tee /etc/apt/sources.list.d/stripe.list
sudo apt update
sudo apt install stripe
```

**Verify:**
```bash
stripe --version
```

---

### Step 3: Get Webhook Secret

**Terminal 1:**
```bash
# Login (opens browser)
stripe login

# Forward webhooks
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

**You'll see:**
```
> Ready! Your webhook signing secret is whsec_xxxxxxxxxxxxx
```

**Copy that `whsec_...` and update `.env`:**
```bash
STRIPE_WEBHOOK_SECRET=whsec_paste_it_here
```

**Keep Terminal 1 running!**

---

### Step 4: Start Backend

**Terminal 2:**
```bash
cd /home/himanshu/Rival/strpe-connect
go run main.go
```

Expected output:
```
✅ Database connected successfully
🚀 Server starting on http://localhost:8080
```

---

### Step 5: Start Frontend

**Terminal 3:**
```bash
cd /home/himanshu/Rival/strpe-connect/frontend
npm install  # First time only
npm run dev
```

Expected output:
```
VITE v5.0.8  ready in 500 ms
➜  Local:   http://localhost:3000/
```

---

### Step 6: Test in Browser

Open: **http://localhost:3000**

---

## 🧪 Test Scenarios

### Test 1: Developer Onboarding

1. **Click**: "👨‍💻 Developer Dashboard" tab
2. **Organization ID**: Already filled with `22222222-2222-2222-2222-222222222222`
3. **Click**: "Start Onboarding" button
4. **Fill Stripe Form**:
   - Business name: "Test Developer Inc"
   - Business type: Individual
   - Phone: Any test number
   - **Bank Account**: Use Stripe test numbers:
     - Routing: `110000000`
     - Account: `000123456789`
5. **Complete verification**
6. **Return to app**
7. **Refresh page**

**Expected Result:**
```
✅ Onboarding Complete
✅ Payouts Enabled: Yes
✅ Charges Enabled: Yes
Balance: $0.00
```

**Terminal 1 should show:**
```
account.updated [evt_xxxxx]
```

---

### Test 2: Function Execution Payment

1. **Click**: "👤 User Dashboard" tab
2. **Change Organization ID to**: `11111111-1111-1111-1111-111111111111`
3. **Function ID**: Keep as `func-test-123`
4. **Amount**: `5.00`
5. **Click**: "Execute Function & Pay $5.00"

**Expected Result:**
```json
{
  "transaction_id": "txn-xxx",
  "amount": 5.00,
  "platform_fee": 0.00,
  "net_amount": 5.00,
  "user_balance": 495.00,      ← Decreased from $500
  "developer_balance": 5.00,    ← Increased from $0
  "message": "Payment processed successfully"
}
```

**Verify in database:**
```bash
psql -U postgres -d rival -c "
SELECT
    o.name,
    COALESCE(a.account_balance, 0) as user_balance,
    COALESCE(dw.balance, 0) as dev_balance
FROM tenant_schema.organizations o
LEFT JOIN tenant_schema.accounts a ON o.id = a.organization_id
LEFT JOIN tenant_schema.developer_wallets dw ON o.id = dw.organization_id
WHERE o.id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222'
);"
```

---

### Test 3: Multiple Function Executions (Build Up Balance)

Execute 10 functions to reach $50:

1. Stay on "👤 User Dashboard"
2. Click "Execute Function & Pay $5.00" **10 times**
3. Watch balances update:
   - User: $500 → $450
   - Developer: $0 → $50

---

### Test 4: Developer Withdrawal

1. **Switch to**: "👨‍💻 Developer Dashboard"
2. **Organization ID**: `22222222-2222-2222-2222-222222222222`
3. **Scroll to**: "Request Withdrawal" section
4. **Amount**: `50.00`
5. **Click**: "Withdraw $50.00"

**Expected Result:**
```json
{
  "withdrawal_id": "wd-xxx",
  "amount": 50.00,
  "status": "processing",
  "message": "Withdrawal request created successfully",
  "estimated_arrival": "2-3 business days"
}
```

**Terminal 1 should show:**
```
payout.created [evt_xxxxx]
payout.paid [evt_xxxxx]  ← In test mode, happens instantly
```

**Developer balance should update:**
```
Current Balance: $0.00  ← Decreased from $50
Total Earned: $50.00
Total Withdrawn: $50.00
```

---

## ✅ Success Checklist

After testing, you should have:

- [x] Developer completed onboarding
- [x] Developer has Stripe Connect account
- [x] User balance decreased by function executions
- [x] Developer balance increased
- [x] Transaction history shows all payments
- [x] Withdrawal processed successfully
- [x] Webhooks received in Terminal 1
- [x] All balances updated correctly

---

## 🐛 Troubleshooting

### Error: "wallet not found"
**Solution**: Complete developer onboarding first (Test 1)

### Error: "insufficient balance"
**Solution**:
```sql
-- Add more balance to user
UPDATE tenant_schema.accounts
SET account_balance = 1000.00
WHERE organization_id = '11111111-1111-1111-1111-111111111111';
```

### Error: "minimum withdrawal amount is $50"
**Solution**: Execute more functions until developer balance >= $50

### Error: "webhook signature verification failed"
**Solution**:
1. Check `.env` has correct `whsec_...` from Stripe CLI
2. Restart backend: `go run main.go`

### Error: "Database connection failed"
**Solution**: Check PostgreSQL is running and credentials in `.env`

---

## 📊 Database Queries for Verification

### Check All Balances:
```sql
SELECT
    o.id,
    o.name,
    COALESCE(a.account_balance, 0) as user_balance,
    COALESCE(dw.balance, 0) as dev_wallet,
    COALESCE(dw.total_earned, 0) as total_earned,
    COALESCE(dw.total_withdrawn, 0) as total_withdrawn
FROM tenant_schema.organizations o
LEFT JOIN tenant_schema.accounts a ON o.id = a.organization_id
LEFT JOIN tenant_schema.developer_wallets dw ON o.id = dw.organization_id
WHERE o.id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222'
);
```

### View All Transactions:
```sql
SELECT
    id,
    user_organization_id,
    developer_organization_id,
    amount,
    net_amount,
    status,
    executed_at
FROM tenant_schema.function_execution_transactions
ORDER BY executed_at DESC
LIMIT 10;
```

### View Withdrawal History:
```sql
SELECT
    id,
    organization_id,
    amount,
    status,
    requested_at,
    completed_at
FROM tenant_schema.withdrawal_requests
ORDER BY requested_at DESC;
```

---

## 🎯 Quick Reference

### Organization IDs (Easy to remember!)
```
User Org (buyer):  11111111-1111-1111-1111-111111111111
Dev Org (seller):  22222222-2222-2222-2222-222222222222
```

### Terminal Commands:
```bash
# Terminal 1: Webhooks
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect

# Terminal 2: Backend
go run main.go

# Terminal 3: Frontend
cd frontend && npm run dev
```

### URLs:
```
Frontend: http://localhost:3000
Backend:  http://localhost:8080
Health:   http://localhost:8080/health
```

---

## 🎉 You Did It!

Once all tests pass, you've successfully:
- ✅ Set up Stripe Connect Express
- ✅ Completed developer onboarding
- ✅ Processed payments (user → developer)
- ✅ Handled withdrawals ($50+ minimum)
- ✅ Received webhook notifications
- ✅ Maintained accurate balance tracking

**Next**: Review `INTEGRATION_GUIDE.md` to merge this into your main codebase!
