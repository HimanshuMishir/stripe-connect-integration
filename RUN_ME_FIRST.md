# 🚀 Quick Start - Run This First!

## Complete Testing in 5 Steps

### Step 1: Update .env with TEST Keys ⚠️

**IMPORTANT**: You're currently using LIVE keys. Switch to TEST keys!

1. Go to: https://dashboard.stripe.com/test/apikeys
2. Make sure "Test mode" toggle is ON (top-right)
3. Copy your keys:
   - Secret key: `sk_test_...`
   - Publishable key: `pk_test_...`
4. Update `.env`:
   ```bash
   STRIPE_SECRET_KEY=sk_test_YOUR_KEY_HERE
   STRIPE_PUBLISHABLE_KEY=pk_test_YOUR_KEY_HERE
   ```

### Step 2: Setup Database

```bash
cd /home/himanshu/Rival/strpe-connect
./setup-complete.sh
```

This will:
- Apply database schema
- Create test organizations
- Add test user with $500 balance

### Step 3: Get Webhook Secret

**Terminal 1** - Start webhook forwarding:
```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

You'll see:
```
> Ready! Your webhook signing secret is whsec_xxxxx...
```

**Copy that secret** and update `.env`:
```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxx...
```

**Keep this terminal running!**

### Step 4: Start Backend

**Terminal 2** - Start the Go server:
```bash
cd /home/himanshu/Rival/strpe-connect
go mod download  # First time only
go run main.go
```

You should see:
```
✅ Database connected successfully
🚀 Server starting on http://localhost:8080
```

### Step 5: Start Frontend

**Terminal 3** - Start React app:
```bash
cd /home/himanshu/Rival/strpe-connect/frontend
npm install  # First time only
npm run dev
```

Visit: **http://localhost:3000**

---

## 🧪 Testing the Complete Flow

### Test 1: Developer Onboarding

1. Open http://localhost:3000
2. Click "👨‍💻 Developer Dashboard"
3. Enter Organization ID: `dev-org-test`
4. Click "Start Onboarding"
5. Complete Stripe Connect form:
   - Use test business information
   - Bank account: Use Stripe test account numbers
   - Complete verification

### Test 2: Function Execution Payment

1. Click "👤 User Dashboard"
2. Enter Organization ID: `user-org-test`
3. Enter amount: `5.00`
4. Click "Execute Function & Pay"
5. Check results:
   - User balance decreased: $500 → $495
   - Developer balance increased: $0 → $5

### Test 3: Developer Withdrawal

1. Switch to "👨‍💻 Developer Dashboard"
2. Organization ID: `dev-org-test`
3. Execute more functions until balance >= $50
4. Click "Request Withdrawal"
5. Enter amount: `50.00`
6. Watch webhook terminal for payout events

---

## ✅ Expected Results

After testing, you should see:

**Database:**
```sql
-- Check balances
SELECT
    o.name,
    COALESCE(a.account_balance, 0) as user_balance,
    COALESCE(dw.balance, 0) as dev_balance
FROM tenant_schema.organizations o
LEFT JOIN tenant_schema.accounts a ON o.id = a.organization_id
LEFT JOIN tenant_schema.developer_wallets dw ON o.id = dw.organization_id
WHERE o.id IN ('user-org-test', 'dev-org-test');
```

**Terminal 1 (Webhooks):**
```
account.updated [evt_xxx]
payout.created [evt_xxx]
payout.paid [evt_xxx]
```

**Browser:**
- Developer dashboard shows earnings
- User dashboard shows payment confirmation
- Transaction history displays correctly

---

## 🐛 Troubleshooting

### "Webhook signature verification failed"
- Make sure `.env` has the correct `whsec_...` from Stripe CLI
- Restart the server after updating `.env`

### "Wallet not found"
- Complete developer onboarding first
- Check organization ID is correct

### "Insufficient balance"
- Add more balance to user account:
  ```sql
  UPDATE tenant_schema.accounts
  SET account_balance = 500.00
  WHERE organization_id = 'user-org-test';
  ```

### "Database connection failed"
- Check PostgreSQL is running
- Verify credentials in `.env`

---

## 📚 Next Steps

Once everything works:

1. **Review the code** - Check how it's implemented
2. **Read INTEGRATION_GUIDE.md** - Merge into your main codebase
3. **Run TEST_SCENARIOS.md** - Complete test cases
4. **Deploy to production** - Use LIVE keys and real webhooks

---

## 💡 Quick Commands Reference

```bash
# Start all services (3 terminals)

# Terminal 1: Webhooks
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect

# Terminal 2: Backend
go run main.go

# Terminal 3: Frontend
cd frontend && npm run dev

# Test API endpoints
./test-api.sh

# Check database
psql -U postgres -d rival
```

---

## 🎉 Success Criteria

You know it's working when:

✅ Server starts without errors
✅ Frontend loads at localhost:3000
✅ Developer can complete onboarding
✅ User can pay developer
✅ Developer receives payment in wallet
✅ Developer can withdraw (>= $50)
✅ Webhooks show in Terminal 1

**Happy testing!** 🚀
