# 🚀 START HERE - Quick Setup

## Fixed! Ready to Test

The UUID error has been fixed. Follow these 5 simple steps:

---

## Step 1: Run Setup (1 minute)

```bash
cd /home/himanshu/Rival/strpe-connect
./setup-complete.sh
```

This creates:
- User Org: `11111111-1111-1111-1111-111111111111` (with $500 balance)
- Dev Org:  `22222222-2222-2222-2222-222222222222` (will earn money)

---

## Step 2: Get Webhook Secret (2 minutes)

**Terminal 1:**
```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

Copy the `whsec_...` it shows and update `.env`:
```
STRIPE_WEBHOOK_SECRET=whsec_paste_here
```

**Keep Terminal 1 running!**

---

## Step 3: Start Backend (30 seconds)

**Terminal 2:**
```bash
go run main.go
```

---

## Step 4: Start Frontend (30 seconds)

**Terminal 3:**
```bash
cd frontend && npm run dev
```

---

## Step 5: Test in Browser

Open: **http://localhost:3000**

### Developer Onboarding:
- Organization ID: `22222222-2222-2222-2222-222222222222`
- Click "Start Onboarding"
- Complete Stripe form

### User Payment:
- Organization ID: `11111111-1111-1111-1111-111111111111`
- Amount: `5.00`
- Click "Execute Function & Pay"

### Withdrawal:
- Do 10 payments (to reach $50)
- Request withdrawal of $50

---

## ✅ Success!

You should see:
- Developer earns money from user payments
- Developer can withdraw when balance >= $50
- All webhooks received in Terminal 1

**Full guide**: Read `TESTING_GUIDE.md` for detailed testing scenarios.
