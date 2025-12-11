
# Stripe Webhook Setup Guide

## Option 1: Local Development (Recommended for Testing)

### Step 1: Install Stripe CLI

**Mac:**
```bash
brew install stripe/stripe-cli/stripe
```

**Linux:**
```bash
curl -s https://packages.stripe.com/api/v1/keys/gpg | sudo apt-key add -
echo "deb https://packages.stripe.com/stripe-cli-debian-local stable main" | sudo tee -a /etc/apt/sources.list.d/stripe.list
sudo apt update
sudo apt install stripe
```

**Windows:**
Download from: https://github.com/stripe/stripe-cli/releases/latest

### Step 2: Login to Stripe

```bash
stripe login
```

This will open your browser to authenticate. Click "Allow access".

### Step 3: Forward Webhooks to Your Local Server

```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

You'll see output like:
```
> Ready! Your webhook signing secret is whsec_1234567890abcdefghijklmnopqrstuvwxyz
```

### Step 4: Copy the Webhook Secret

Copy the `whsec_...` value and add it to your `.env` file:

```bash
STRIPE_WEBHOOK_SECRET=whsec_1234567890abcdefghijklmnopqrstuvwxyz
```

### Step 5: Keep the CLI Running

Keep this terminal window open while testing. The CLI will forward all Stripe events to your local server.

---

## Option 2: Production Deployment (Stripe Dashboard)

### Step 1: Go to Stripe Dashboard

1. Visit https://dashboard.stripe.com/webhooks
2. Click "Add endpoint" button

### Step 2: Configure Endpoint

**Endpoint URL:**
```
https://your-production-domain.com/api/webhooks/stripe-connect
```

Example: `https://api.rival.io/api/webhooks/stripe-connect`

**Description:** (Optional)
```
Stripe Connect webhook for Function Marketplace
```

### Step 3: Select Events to Listen To

Click "Select events" and choose these **IMPORTANT** events:

#### Connect Account Events
- ✅ `account.updated` - **(REQUIRED)** Updates onboarding status
- ✅ `account.application.deauthorized` - Account disconnected
- ✅ `account.external_account.created` - Bank account added
- ✅ `account.external_account.deleted` - Bank account removed
- ✅ `account.external_account.updated` - Bank account updated

#### Payout Events
- ✅ `payout.created` - Payout initiated
- ✅ `payout.paid` - **(REQUIRED)** Payout successful
- ✅ `payout.failed` - **(REQUIRED)** Payout failed
- ✅ `payout.canceled` - Payout cancelled
- ✅ `payout.updated` - Payout status changed

#### Transfer Events (Optional but Recommended)
- ✅ `transfer.created` - Transfer created
- ✅ `transfer.reversed` - Transfer reversed
- ✅ `transfer.updated` - Transfer updated

### Minimum Required Events
If you want to select fewer events, these are **absolutely required**:
1. `account.updated` - To update onboarding status
2. `payout.paid` - To confirm successful withdrawals
3. `payout.failed` - To handle failed withdrawals

### Step 4: API Version

**Select the latest API version** (usually pre-selected)

### Step 5: Save and Get Signing Secret

1. Click "Add endpoint"
2. On the next screen, click "Reveal" under "Signing secret"
3. Copy the secret (looks like: `whsec_...`)
4. Add it to your production `.env`:
   ```
   STRIPE_WEBHOOK_SECRET=whsec_your_production_secret_here
   ```

### Step 6: Test the Webhook

Stripe provides a "Send test webhook" button. Click it and send a test `account.updated` event.

Check your server logs to verify it was received.

---

## Webhook Secret Explained

### What is it?
The webhook signing secret is used to verify that webhook events actually come from Stripe and not from an attacker.

### Format
- Test mode: `whsec_test_...`
- Live mode: `whsec_...`

### Where to find it?

**Local Development:**
- Stripe CLI shows it when you run `stripe listen`
- Changes every time you restart the CLI

**Production:**
- Stripe Dashboard → Webhooks → Click your endpoint → "Signing secret"
- Stays the same unless you regenerate it

### Security
- ⚠️ Never commit webhook secrets to Git
- ⚠️ Keep them in `.env` file (which is gitignored)
- ⚠️ Use different secrets for test vs production

---

## Testing Your Webhook Setup

### Method 1: Use Stripe CLI (Easiest)

```bash
# In one terminal, forward webhooks
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect

# In another terminal, trigger test events
stripe trigger account.updated
stripe trigger payout.paid
stripe trigger payout.failed
```

### Method 2: Test from Dashboard

1. Go to https://dashboard.stripe.com/webhooks
2. Click your endpoint
3. Click "Send test webhook"
4. Select an event type
5. Click "Send test webhook"

### Method 3: Actual Flow Testing

1. Complete Connect onboarding (triggers `account.updated`)
2. Request a withdrawal (triggers `payout.created`, then `payout.paid`)

---

## Troubleshooting

### "Webhook signature verification failed"

**Cause:** Wrong webhook secret in `.env`

**Fix:**
1. Check that `STRIPE_WEBHOOK_SECRET` in `.env` matches exactly
2. For local dev: Copy from Stripe CLI output
3. For production: Copy from Dashboard
4. Restart your server after updating `.env`

### "No webhook events received"

**Cause:** Webhook not set up correctly

**Fix:**
1. Verify Stripe CLI is running: `stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect`
2. Check server is running on port 8080
3. Verify endpoint URL is correct

### "Webhook received but not processed"

**Cause:** Event type not handled

**Fix:**
Check server logs for unhandled event types. Add handlers in `handlers/stripe_connect_handler.go`

---

## Event Examples

### account.updated
Sent when:
- Developer completes onboarding
- Verification status changes
- Bank account updated

Your handler updates:
- `onboarding_completed`
- `payouts_enabled`
- `charges_enabled`

### payout.paid
Sent when:
- Money successfully transferred to developer's bank
- Usually 2-3 days after withdrawal request

Your handler updates:
- Withdrawal status to "completed"
- Logs successful payout

### payout.failed
Sent when:
- Bank account is invalid
- Insufficient funds (shouldn't happen with Connect)
- Bank rejects transfer

Your handler updates:
- Withdrawal status to "failed"
- Stores failure reason

---

## Quick Reference

### Local Development
```bash
# Terminal 1: Start server
go run main.go

# Terminal 2: Forward webhooks
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect

# Terminal 3: Trigger test events
stripe trigger account.updated
```

### Production
- Dashboard: https://dashboard.stripe.com/webhooks
- Endpoint URL: `https://your-domain.com/api/webhooks/stripe-connect`
- Required events: `account.updated`, `payout.paid`, `payout.failed`

---

## Need Help?

1. Check server logs for detailed error messages
2. Verify webhook secret matches exactly
3. Ensure server is running and accessible
4. Test with `stripe trigger` commands
5. Check Stripe Dashboard → Webhooks → Event log for delivery status
