# Testing Checklist - IMPORTANT!

## ⚠️ CRITICAL: Use Test Keys for Development!

You're currently using **LIVE** Stripe keys in your .env file. This is DANGEROUS for testing!

### What's the Difference?

**Test Keys** (for development):
- Start with `sk_test_...` and `pk_test_...`
- Free to use, no real money
- Can test everything safely
- Stripe provides test credit cards

**Live Keys** (for production only):
- Start with `sk_live_...` and `pk_live_...`
- Real money transactions
- Real bank accounts
- Should NEVER be used for testing!

### How to Get Test Keys

1. Go to: https://dashboard.stripe.com/test/apikeys
2. Click the "Test mode" toggle in top-right (should be ON)
3. Copy your test keys:
   - **Secret key**: `sk_test_...`
   - **Publishable key**: `pk_test_...`

### Update Your .env

Replace with TEST keys:
```bash
STRIPE_SECRET_KEY=sk_test_YOUR_TEST_KEY_HERE
STRIPE_PUBLISHABLE_KEY=pk_test_YOUR_TEST_KEY_HERE
```

## Current Status of Your .env

✅ Database configured
✅ Server port configured
❌ Using LIVE Stripe keys (should use TEST)
❌ Webhook secret not configured yet

## Next Steps

1. ✅ Get test keys from Stripe Dashboard
2. ✅ Update .env with test keys
3. ✅ Get webhook secret (instructions below)
4. ✅ Test the application
