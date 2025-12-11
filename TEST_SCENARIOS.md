# Test Scenarios

Complete test cases to verify your Stripe Connect implementation.

## Test Data Setup

### Create Test Organizations and Accounts

```sql
-- Developer Organization
INSERT INTO tenant_schema.organizations (id, name, bio)
VALUES
  ('dev-org-1', 'Acme Functions Inc', 'We build amazing functions'),
  ('dev-org-2', 'Beta Developers', 'Quality serverless functions');

-- User Organizations
INSERT INTO tenant_schema.organizations (id, name)
VALUES
  ('user-org-1', 'Client Corp'),
  ('user-org-2', 'Enterprise Solutions');

-- User Accounts with Balance
INSERT INTO tenant_schema.accounts (id, organization_id, account_balance)
VALUES
  (uuid_generate_v4(), 'user-org-1', 500.00),
  (uuid_generate_v4(), 'user-org-2', 1000.00);

-- Developer Functions (if you have functions table)
INSERT INTO tenant_schema.functions (function_id, organization_id, function_name, price_per_api_request)
VALUES
  (uuid_generate_v4(), 'dev-org-1', 'image-processor', 2.50),
  (uuid_generate_v4(), 'dev-org-1', 'data-analyzer', 5.00),
  (uuid_generate_v4(), 'dev-org-2', 'pdf-generator', 3.00);
```

## Scenario 1: Developer Onboarding

### Test Case 1.1: First-Time Onboarding

**Steps:**
1. Make API call to create Connect account
2. Open the onboarding URL in browser
3. Complete Stripe Connect Express onboarding
4. Return to app

**Expected Result:**
- Stripe Connect account created
- Developer wallet created in database
- Onboarding URL generated
- After completion: `onboarding_completed` = true

**Verify:**
```sql
SELECT * FROM tenant_schema.developer_wallets
WHERE organization_id = 'dev-org-1';
```

**Expected Fields:**
- `stripe_connect_account_id`: acct_xxxxx
- `onboarding_completed`: true
- `payouts_enabled`: true
- `charges_enabled`: true

### Test Case 1.2: Refresh Onboarding Link

**Steps:**
1. Call onboarding endpoint again
2. Should return new onboarding link

**Expected Result:**
- New account link generated
- Can continue onboarding from where left off

### Test Case 1.3: Incomplete Onboarding

**Steps:**
1. Start onboarding but don't complete
2. Check account status

**Expected Result:**
- `onboarding_completed`: false
- `payouts_enabled`: false
- Cannot request withdrawals

## Scenario 2: Function Execution Payments

### Test Case 2.1: Successful Payment

**Prerequisites:**
- Developer completed onboarding
- User has sufficient balance

**Steps:**
1. Execute function with amount $5.00
2. Check user balance
3. Check developer wallet

**Expected Result:**
- User balance: $500.00 → $495.00
- Developer balance: $0.00 → $5.00 (assuming 0% platform fee)
- Transaction record created
- Status: "completed"

**Verify:**
```sql
-- Check transaction
SELECT * FROM tenant_schema.function_execution_transactions
WHERE user_organization_id = 'user-org-1'
ORDER BY executed_at DESC LIMIT 1;

-- Check balances
SELECT account_balance FROM tenant_schema.accounts
WHERE organization_id = 'user-org-1';

SELECT balance FROM tenant_schema.developer_wallets
WHERE organization_id = 'dev-org-1';
```

### Test Case 2.2: Insufficient User Balance

**Steps:**
1. Attempt execution with amount > user balance
2. Try to execute function for $600.00 (user only has $495.00)

**Expected Result:**
- Payment fails with error: "insufficient balance"
- No transaction created
- Balances unchanged

### Test Case 2.3: Multiple Executions

**Steps:**
1. Execute function 5 times with $2.50 each
2. Check total earned

**Expected Result:**
- User balance reduced by $12.50
- Developer earned $12.50
- 5 transaction records created

**Verify:**
```sql
SELECT COUNT(*), SUM(amount), SUM(net_amount)
FROM tenant_schema.function_execution_transactions
WHERE developer_organization_id = 'dev-org-1';
```

### Test Case 2.4: Platform Fee Calculation

**Prerequisites:**
- Set platform fee to 10% in code

**Steps:**
1. Execute function for $10.00
2. Check transaction details

**Expected Result:**
- Amount: $10.00
- Platform fee: $1.00
- Net amount to developer: $9.00
- Developer wallet increased by $9.00

## Scenario 3: Withdrawals

### Test Case 3.1: Successful Withdrawal

**Prerequisites:**
- Developer has $100.00 balance
- Onboarding completed
- Payouts enabled

**Steps:**
1. Request withdrawal of $50.00
2. Wait for processing
3. Check withdrawal status

**Expected Result:**
- Withdrawal status: "processing" → "completed"
- Developer balance: $100.00 → $50.00
- Stripe payout created
- Money transferred to bank account

**Verify:**
```sql
SELECT * FROM tenant_schema.withdrawal_requests
WHERE organization_id = 'dev-org-1'
ORDER BY requested_at DESC LIMIT 1;

SELECT balance FROM tenant_schema.developer_wallets
WHERE organization_id = 'dev-org-1';
```

### Test Case 3.2: Below Minimum Withdrawal

**Steps:**
1. Attempt to withdraw $25.00

**Expected Result:**
- Error: "minimum withdrawal amount is $50.00"
- No withdrawal request created
- Balance unchanged

### Test Case 3.3: Withdraw More Than Balance

**Steps:**
1. Developer has $40.00
2. Attempt to withdraw $50.00

**Expected Result:**
- Error: "insufficient balance"
- No withdrawal created

### Test Case 3.4: Pending Withdrawal Blocks Another

**Steps:**
1. Developer has $100.00
2. Request withdrawal of $60.00 (status: pending)
3. Attempt another withdrawal of $50.00

**Expected Result:**
- Available balance: $100 - $60 (pending) = $40
- Second withdrawal fails: insufficient available balance

### Test Case 3.5: Onboarding Not Complete

**Prerequisites:**
- Developer hasn't completed onboarding
- Has balance from test data

**Steps:**
1. Attempt withdrawal

**Expected Result:**
- Error: "please complete Stripe Connect onboarding"

## Scenario 4: Wallet & Transaction History

### Test Case 4.1: View Wallet Balance

**Steps:**
1. Get wallet balance

**Expected Result:**
```json
{
  "balance": 100.00,
  "total_earned": 150.00,
  "total_withdrawn": 50.00,
  "pending_withdrawals": 0.00,
  "can_withdraw": true,
  "minimum_withdrawal": 50.00
}
```

### Test Case 4.2: Transaction History

**Steps:**
1. Get transaction history with pagination

**Expected Result:**
- List of transactions ordered by date (newest first)
- Shows all function executions
- Includes user org, amount, fees, net amount

### Test Case 4.3: Withdrawal History

**Steps:**
1. Get withdrawal history

**Expected Result:**
- List of all withdrawal requests
- Shows status, dates, amounts
- Includes failure reasons if any

## Scenario 5: Webhooks

### Test Case 5.1: Account Updated Webhook

**Steps:**
1. Complete onboarding
2. Stripe sends `account.updated` webhook

**Expected Result:**
- Webhook received and verified
- Developer wallet updated with new status
- `onboarding_completed` = true
- `payouts_enabled` = true

**Check Logs:**
```
✅ Updated account status for dev-org-1: onboarding=true, payouts=true, charges=true
```

### Test Case 5.2: Payout Paid Webhook

**Steps:**
1. Request withdrawal
2. Stripe processes payout
3. Receives `payout.paid` webhook

**Expected Result:**
- Webhook logged
- Withdrawal status updated to "completed"

### Test Case 5.3: Payout Failed Webhook

**Steps:**
1. Simulate failed payout (test mode)
2. Receives `payout.failed` webhook

**Expected Result:**
- Webhook logged
- Failure handled gracefully
- Developer notified (if notifications implemented)

## Scenario 6: Edge Cases

### Test Case 6.1: Concurrent Withdrawals

**Steps:**
1. Developer has $100.00
2. Make 3 simultaneous withdrawal requests of $40 each

**Expected Result:**
- Only one succeeds
- Others fail with insufficient balance
- Database constraints prevent over-withdrawal

### Test Case 6.2: Deleted Organization

**Steps:**
1. Delete organization
2. Check cascade deletion

**Expected Result:**
- Developer wallet deleted (CASCADE)
- Transactions remain (or handle as per business logic)
- Withdrawals remain for audit

### Test Case 6.3: Very Large Amount

**Steps:**
1. Execute function for $100,000.00

**Expected Result:**
- Handled correctly
- No integer overflow
- Decimal precision maintained

### Test Case 6.4: Zero Amount

**Steps:**
1. Attempt function execution with $0.00

**Expected Result:**
- Error: "invalid amount"

### Test Case 6.5: Negative Amount

**Steps:**
1. Attempt withdrawal of -$50.00

**Expected Result:**
- Validation error
- Request rejected

## Scenario 7: Multi-Developer Environment

### Test Case 7.1: Multiple Developers, One User

**Setup:**
- User executes functions from different developers

**Steps:**
1. User executes dev-org-1 function: $5.00
2. User executes dev-org-2 function: $3.00
3. User executes dev-org-1 function again: $2.00

**Expected Result:**
- User balance reduced by $10.00
- Dev-org-1 wallet: $7.00
- Dev-org-2 wallet: $3.00
- Each has correct transaction history

### Test Case 7.2: One Developer, Multiple Users

**Steps:**
1. User-org-1 executes dev-org-1 function: $5.00
2. User-org-2 executes dev-org-1 function: $10.00

**Expected Result:**
- Dev-org-1 wallet: $15.00
- Transaction history shows both users
- Both user balances reduced correctly

## Scenario 8: Performance Tests

### Test Case 8.1: High Volume Transactions

**Steps:**
1. Execute 1000 function calls in quick succession

**Expected Result:**
- All transactions processed correctly
- No race conditions
- Database maintains consistency
- Response times acceptable (<500ms per transaction)

### Test Case 8.2: Concurrent User Requests

**Steps:**
1. 100 users execute functions simultaneously

**Expected Result:**
- All processed without conflicts
- Account balances correct
- No deadlocks

## Verification Queries

### Overall System Health

```sql
-- Total platform revenue
SELECT
  COUNT(*) as total_transactions,
  SUM(amount) as total_processed,
  SUM(platform_fee) as total_platform_revenue,
  SUM(net_amount) as total_paid_to_developers
FROM tenant_schema.function_execution_transactions;

-- Developer earnings summary
SELECT
  o.name,
  dw.balance as current_balance,
  dw.total_earned,
  dw.total_withdrawn,
  dw.onboarding_completed
FROM tenant_schema.developer_wallets dw
JOIN tenant_schema.organizations o ON dw.organization_id = o.id
ORDER BY dw.total_earned DESC;

-- Withdrawal statistics
SELECT
  status,
  COUNT(*) as count,
  SUM(amount) as total_amount
FROM tenant_schema.withdrawal_requests
GROUP BY status;

-- User spending
SELECT
  a.organization_id,
  o.name,
  a.account_balance,
  COALESCE(SUM(t.amount), 0) as total_spent
FROM tenant_schema.accounts a
JOIN tenant_schema.organizations o ON a.organization_id = o.id
LEFT JOIN tenant_schema.function_execution_transactions t
  ON t.user_account_id = a.id
GROUP BY a.organization_id, o.name, a.account_balance;
```

## Checklist

Before considering testing complete:

- [ ] All onboarding scenarios pass
- [ ] Payment flow works correctly
- [ ] Withdrawal minimum enforced
- [ ] Webhook handling verified
- [ ] Edge cases handled gracefully
- [ ] Database constraints working
- [ ] Performance acceptable under load
- [ ] No race conditions or deadlocks
- [ ] Proper error messages shown
- [ ] Audit trail complete (all transactions logged)
- [ ] Platform fees calculated correctly
- [ ] Multi-tenant isolation verified

## Test Automation

Consider automating these tests:

```bash
# Run all test scenarios
go test ./internal/services/... -v
go test ./internal/repository/... -v

# Integration tests
go test ./test/integration/... -v

# Load tests
k6 run loadtest.js
```

## Monitoring & Alerts

Set up monitoring for:
- Failed transactions
- Failed withdrawals
- Low user balances
- High pending withdrawal amounts
- Webhook failures
- Slow API responses

That's it! If all these scenarios pass, your Stripe Connect integration is production-ready.
