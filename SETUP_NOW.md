# 🚀 Setup Instructions - DO THIS NOW

## ⚠️ IMPORTANT: For Local Testing, DON'T Use Dashboard Yet!

You're currently on the Stripe Dashboard webhook setup screen. **Close it for now** - we'll use Stripe CLI instead for local testing.

---

## Step 1: Close Dashboard Tab

The dashboard webhook setup is for **production only**. For local testing, Stripe CLI is much easier.

---

## Step 2: Get Test Stripe Keys

1. Go to: https://dashboard.stripe.com/test/apikeys
2. Make sure **"Test mode"** toggle is ON (top-right corner)
3. Copy these two keys:
   - **Secret key**: Starts with `sk_test_...`
   - **Publishable key**: Starts with `pk_test_...`

---

## Step 3: Update .env File

Open `/home/himanshu/Rival/strpe-connect/.env` and update:

```bash
# Database Configuration (keep as is)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=rival

# ⚠️ REPLACE THESE WITH TEST KEYS!
STRIPE_SECRET_KEY=sk_test_YOUR_TEST_KEY_HERE_FROM_DASHBOARD
STRIPE_PUBLISHABLE_KEY=pk_test_YOUR_TEST_KEY_HERE_FROM_DASHBOARD

# Leave this blank for now - we'll get it in Step 5
STRIPE_WEBHOOK_SECRET=

# Server Configuration (keep as is)
PORT=8080

# Frontend URL (keep as is)
FRONTEND_URL=http://localhost:3000
```

**Save the file** after updating!

---

## Step 4: Setup Database

Run this command:

```bash
cd /home/himanshu/Rival/strpe-connect
./setup-complete.sh
```

This will:
- ✅ Apply database schema
- ✅ Create test organizations
- ✅ Add test user with $500 balance

---

## Step 5: Get Webhook Secret (Using Stripe CLI)

### Install Stripe CLI (if not already installed):

**Mac:**
```bash
brew install stripe/stripe-cli/stripe
```

**Linux (Ubuntu/Debian):**
```bash
curl -s https://packages.stripe.com/api/v1/keys/gpg | sudo apt-key add -
echo "deb https://packages.stripe.com/stripe-cli-debian-local stable main" | sudo tee /etc/apt/sources.list.d/stripe.list
sudo apt update
sudo apt install stripe
```

**Verify installation:**
```bash
stripe --version
```

### Login to Stripe:
```bash
stripe login
```

This will open your browser - click "Allow access"

### Forward Webhooks to Your Local Server:

**Open a NEW terminal** and run:
```bash
cd /home/himanshu/Rival/strpe-connect
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

You'll see:
```
> Ready! Your webhook signing secret is whsec_xxxxxxxxxxxxxxxxxxxxx
```

### Copy That Secret!

**IMPORTANT**: Copy the entire `whsec_...` string

Update your `.env` file:
```bash
STRIPE_WEBHOOK_SECRET=whsec_paste_the_secret_here
```

**Keep this terminal running!** This is forwarding webhooks to your local server.

---

## Step 6: Start Backend Server

**Open a NEW terminal** (Terminal 2):

```bash
cd /home/himanshu/Rival/strpe-connect

# Download dependencies (first time only)
go mod download

# Start server
go run main.go
```

You should see:
```
✅ Database connected successfully
🚀 Server starting on http://localhost:8080
```

---

## Step 7: Start Frontend

**Open a NEW terminal** (Terminal 3):

```bash
cd /home/himanshu/Rival/strpe-connect/frontend

# Install dependencies (first time only)
npm install

# Start dev server
npm run dev
```

You should see:
```
VITE v5.0.8  ready in 500 ms

➜  Local:   http://localhost:3000/
```

---

## Step 8: Test in Browser

1. **Open**: http://localhost:3000

2. **Test Developer Onboarding**:
   - Click "👨‍💻 Developer Dashboard"
   - Organization ID: `dev-org-test`
   - Click "Start Onboarding"
   - Complete Stripe Connect form
   - Watch Terminal 1 (webhooks) for `account.updated` event

3. **Test Function Payment**:
   - Click "👤 User Dashboard"
   - Organization ID: `user-org-test`
   - Amount: `5.00`
   - Click "Execute Function & Pay $5.00"
   - Check developer balance increases

4. **Test Withdrawal** (after balance >= $50):
   - Execute functions until dev balance >= $50
   - Go to Developer Dashboard
   - Click "Request Withdrawal"
   - Amount: `50.00`
   - Watch Terminal 1 for `payout.created` and `payout.paid` events

---

## 🎯 You Should Have 3 Terminals Running:

```
Terminal 1: Stripe CLI (webhook forwarding)
├─ stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
└─ Shows webhook events in real-time

Terminal 2: Backend Server
├─ go run main.go
└─ Port: 8080

Terminal 3: Frontend Dev Server
├─ npm run dev
└─ Port: 3000
```

---

## ✅ Success Checklist

Before testing, verify:

- [ ] ✅ Updated .env with **TEST** keys (not LIVE!)
- [ ] ✅ Updated .env with webhook secret from Stripe CLI
- [ ] ✅ Ran `./setup-complete.sh` successfully
- [ ] ✅ Terminal 1 running: `stripe listen`
- [ ] ✅ Terminal 2 running: `go run main.go`
- [ ] ✅ Terminal 3 running: `npm run dev`
- [ ] ✅ Browser open: http://localhost:3000

---

## 🐛 Troubleshooting

### "stripe: command not found"
**Fix**: Install Stripe CLI (see Step 5)

### "Database connection failed"
**Fix**: Check PostgreSQL is running and credentials in .env are correct

### "Webhook signature verification failed"
**Fix**:
1. Make sure `.env` has the correct `whsec_...` from Stripe CLI
2. Restart backend server: `go run main.go`

### "go: cannot find module"
**Fix**: Run `go mod download` first

### Frontend won't start
**Fix**: Make sure Node.js 18+ is installed, then run `npm install`

---

## 📚 When to Use Dashboard Webhooks

Use the Stripe Dashboard webhook setup (what you were doing) **ONLY** when:

- ✅ Deploying to production
- ✅ Have a public HTTPS URL
- ✅ Need permanent webhook endpoint

For local testing, **always use Stripe CLI** - it's much easier!

---

## 🎉 You're Ready!

Once all 3 terminals are running and you can access http://localhost:3000, you're ready to test the complete Stripe Connect flow!

**Next**: Follow Step 8 above to test developer onboarding, payments, and withdrawals.

---

## 💡 Quick Reference

```bash
# Get webhook secret
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect

# Start backend
go run main.go

# Start frontend
cd frontend && npm run dev

# Test database connection
psql -U postgres -d rival -c "SELECT 1"

# Check test data
psql -U postgres -d rival -c "SELECT * FROM tenant_schema.organizations WHERE id IN ('user-org-test', 'dev-org-test')"
```
