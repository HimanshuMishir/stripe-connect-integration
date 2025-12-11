# Visual Guide: Stripe Webhooks Setup

## What Are Webhooks?

Webhooks are HTTP callbacks that Stripe sends to your server when events happen (like when a developer completes onboarding or a payout succeeds).

```
┌─────────────┐                ┌─────────────┐                ┌─────────────┐
│   Stripe    │                │  Your App   │                │  Database   │
│   Server    │                │  (Backend)  │                │             │
└──────┬──────┘                └──────┬──────┘                └──────┬──────┘
       │                              │                              │
       │  1. Developer completes      │                              │
       │     onboarding on Stripe     │                              │
       │                              │                              │
       │  2. Stripe sends webhook     │                              │
       │     POST /webhooks/...       │                              │
       ├─────────────────────────────>│                              │
       │     {                        │                              │
       │       "type": "account.      │                              │
       │         updated",             │                              │
       │       "data": {...}          │                              │
       │     }                        │                              │
       │                              │                              │
       │                              │  3. Verify signature         │
       │                              │     (using webhook secret)   │
       │                              │                              │
       │                              │  4. Update developer wallet  │
       │                              ├─────────────────────────────>│
       │                              │     UPDATE developer_wallets │
       │                              │     SET onboarding_completed │
       │                              │       = true                 │
       │                              │                              │
       │  5. Return 200 OK            │                              │
       │<─────────────────────────────┤                              │
       │                              │                              │
```

## Two Ways to Get Webhooks

### Option 1: Stripe CLI (For Testing)

```
┌──────────────────────────────────────────────────────────────────┐
│  Your Computer                                                    │
│                                                                   │
│  ┌─────────────┐         ┌──────────────┐      ┌─────────────┐ │
│  │   Stripe    │  WSS    │  Stripe CLI  │ HTTP │  Your App   │ │
│  │   Server    ├────────>│  (Proxy)     ├─────>│  :8080      │ │
│  └─────────────┘         └──────────────┘      └─────────────┘ │
│                                 │                                │
│                                 │ Shows events in terminal       │
│                                 v                                │
│                          whsec_xxxxx...                          │
│                          (Your webhook secret)                   │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

**Command:**
```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
```

**Pros:**
- ✅ Instant setup (2 minutes)
- ✅ No public URL needed
- ✅ See events in real-time
- ✅ Perfect for testing

**Cons:**
- ❌ Must keep terminal open
- ❌ Secret changes each restart
- ❌ Only for local development

---

### Option 2: Stripe Dashboard (For Production)

```
┌──────────────────────────────────────────────────────────────────┐
│  Internet                                                         │
│                                                                   │
│  ┌─────────────┐         ┌──────────────┐                       │
│  │   Stripe    │  HTTPS  │  Your App    │                       │
│  │   Server    ├────────>│  (Production)│                       │
│  └─────────────┘         └──────────────┘                       │
│                           https://api.rival.io/                  │
│                           webhooks/stripe-connect                │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

**Setup:**
1. Go to https://dashboard.stripe.com/webhooks
2. Add endpoint: `https://your-domain.com/api/webhooks/stripe-connect`
3. Select events
4. Get permanent webhook secret

**Pros:**
- ✅ Permanent setup
- ✅ Works in production
- ✅ Secret doesn't change
- ✅ Automatic retry on failure

**Cons:**
- ❌ Need public URL (HTTPS)
- ❌ Harder to debug locally
- ❌ Must configure in dashboard

---

## Which Events to Select?

### Minimal Setup (3 events)

If you only want the essentials:

```
✅ account.updated     → Know when developer completes onboarding
✅ payout.paid         → Confirm withdrawal successful
✅ payout.failed       → Handle withdrawal failures
```

### Recommended Setup (11 events)

For better tracking and debugging:

```
Account Events:
✅ account.updated
✅ account.application.deauthorized
✅ account.external_account.created
✅ account.external_account.updated
✅ account.external_account.deleted

Payout Events:
✅ payout.created
✅ payout.paid
✅ payout.failed
✅ payout.canceled
✅ payout.updated
✅ payout.reconciliation_completed
```

---

## How Your Code Handles Webhooks

```go
// handlers/stripe_connect_handler.go

func (h *StripeConnectHandler) HandleWebhook(c *gin.Context) {
    // 1. Read webhook payload
    payload, _ := io.ReadAll(c.Request.Body)

    // 2. Get Stripe signature from header
    signature := c.GetHeader("Stripe-Signature")

    // 3. Verify signature using webhook secret
    event, err := webhook.ConstructEvent(
        payload,
        signature,
        h.webhookSecret,  // ← This is the whsec_... from .env
    )

    // 4. Handle different event types
    switch event.Type {
    case "account.updated":
        // Update developer onboarding status
        h.service.HandleAccountUpdated(...)

    case "payout.paid":
        // Mark withdrawal as completed

    case "payout.failed":
        // Mark withdrawal as failed
    }

    // 5. Return 200 OK to Stripe
    c.JSON(200, gin.H{"received": true})
}
```

---

## Webhook Secret Explained

### What Is It?

A secret key used to verify webhooks actually come from Stripe.

### Format

```
whsec_1a2b3c4d5e6f7g8h9i0j...
```

### How It Works

```
┌─────────────┐                              ┌─────────────┐
│   Stripe    │                              │  Your App   │
└──────┬──────┘                              └──────┬──────┘
       │                                            │
       │ 1. Create payload                         │
       │    { "type": "account.updated", ... }     │
       │                                            │
       │ 2. Sign with webhook secret               │
       │    HMAC-SHA256(payload, secret)           │
       │    = signature                            │
       │                                            │
       │ 3. Send payload + signature               │
       ├───────────────────────────────────────────>│
       │    Header: Stripe-Signature: t=...,v1=... │
       │    Body: { "type": ... }                  │
       │                                            │
       │                                            │ 4. Verify signature
       │                                            │    using same secret
       │                                            │
       │                                            │ 5. If valid:
       │                                            │    Process event
       │                                            │
       │                                            │    If invalid:
       │                                            │    Reject (attacker!)
       │                                            │
       │ 6. Return 200 OK                          │
       │<───────────────────────────────────────────┤
       │                                            │
```

**Why is this important?**

Without verification, anyone could send fake webhooks to your server and manipulate your database!

---

## Testing Webhooks

### Method 1: Stripe CLI Trigger

```bash
# Trigger test events
stripe trigger account.updated
stripe trigger payout.paid
stripe trigger payout.failed
```

### Method 2: Real Flow

```
1. Developer completes onboarding
   → account.updated webhook sent

2. Developer requests withdrawal
   → payout.created webhook sent

3. Stripe processes payout
   → payout.paid webhook sent (2-3 days later)
```

### Method 3: Dashboard Test

1. Go to https://dashboard.stripe.com/webhooks
2. Click your endpoint
3. Click "Send test webhook"
4. Select event type
5. Click send

---

## Common Issues

### Issue: "Webhook signature verification failed"

**Cause:** Wrong webhook secret in `.env`

**Fix:**
```bash
# For Stripe CLI (local):
# 1. Look at terminal running stripe listen
# 2. Copy the whsec_... it shows
# 3. Update .env
STRIPE_WEBHOOK_SECRET=whsec_from_cli_output

# For Dashboard (production):
# 1. Go to dashboard.stripe.com/webhooks
# 2. Click your endpoint
# 3. Click "Reveal" under signing secret
# 4. Copy and update .env
STRIPE_WEBHOOK_SECRET=whsec_from_dashboard

# 4. Restart server
```

### Issue: "Webhooks not received"

**Checklist:**
- [ ] Stripe CLI running? (`stripe listen`)
- [ ] Server running? (port 8080)
- [ ] Correct URL? (`localhost:8080/api/webhooks/stripe-connect`)
- [ ] Firewall not blocking?

### Issue: "Webhook received but event not handled"

**Check server logs:**
```
Received webhook event: account.updated  ← Should see this
✅ Updated account status...              ← And this
```

If you see "Unhandled webhook event", add handler in code.

---

## Security Best Practices

### ✅ DO:
- Store webhook secret in `.env` (gitignored)
- Always verify webhook signature
- Use HTTPS in production
- Return 200 quickly (don't do slow operations)
- Log all webhook events
- Retry failed webhooks

### ❌ DON'T:
- Commit webhook secrets to Git
- Process webhooks without verification
- Use HTTP in production (must be HTTPS)
- Take > 30 seconds to respond
- Trust webhook data without validation
- Ignore failed webhooks

---

## Quick Reference

### Get Webhook Secret - Stripe CLI
```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect
# Copy the whsec_... it shows
```

### Get Webhook Secret - Dashboard
1. https://dashboard.stripe.com/webhooks
2. Click endpoint
3. "Reveal" signing secret
4. Copy whsec_...

### Test Webhook
```bash
stripe trigger account.updated
```

### Check Webhook Logs
- **Stripe CLI**: Shows in terminal
- **Dashboard**: https://dashboard.stripe.com/webhooks → Event log
- **Your Logs**: Check server console

---

## Summary

✅ Webhooks notify your app of Stripe events
✅ Webhook secret verifies authenticity
✅ Stripe CLI: Easy for local testing
✅ Dashboard: Required for production
✅ Must handle: account.updated, payout.paid, payout.failed
✅ Always return 200 OK

**Now you're ready to set up webhooks!** 🎉
