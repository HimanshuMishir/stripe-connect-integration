# Insert Dummy Data for Testing

## Quick Start - Run This Command

```bash
psql -U your_username -d your_database -f /home/himanshu/Rival/strpe-connect/RUN_THIS.sql
```

Or if you're already in psql:

```sql
\i /home/himanshu/Rival/strpe-connect/RUN_THIS.sql
```

## What This Does

For organization ID: `437213bc-d871-401e-ac0e-aa8d4f321e70`

✅ **Updates Developer Wallet**:
- Balance: $125.50
- Total Earned: $300.00
- Total Withdrawn: $174.50

✅ **Inserts 5 Withdrawal Requests**:
1. **Completed** - $100.00 (10 days ago)
2. **Completed** - $74.50 (5 days ago)
3. **Failed** - $50.00 (3 days ago) - shows failure reason
4. **Pending** - $50.00 (2 hours ago) - waiting to process
5. **Processing** - $75.50 (30 minutes ago) - currently processing

## After Running

### View in Frontend

1. Open: http://localhost:5173
2. Go to **Developer Dashboard** tab
3. Set Org ID to: `437213bc-d871-401e-ac0e-aa8d4f321e70`
4. You should see:
   - Wallet Balance: $125.50
   - Can Withdraw: Yes ✓
   - Transaction History (if you added transactions)
   - **Withdrawal History with 5 entries**

### View in Database

```sql
-- Check wallet
SELECT balance, total_earned, total_withdrawn
FROM tenant_schema.developer_wallets
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

-- Check withdrawals
SELECT amount, status, requested_at, completed_at
FROM tenant_schema.withdrawal_requests
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'
ORDER BY requested_at DESC;
```

## Additional Files

### Option 1: Simple Insert (RUN_THIS.sql)
**Recommended** - Just withdrawals and wallet balance
```bash
psql -U user -d db -f RUN_THIS.sql
```

### Option 2: Quick Insert (QUICK_INSERT.sql)
Includes transactions + withdrawals
```bash
psql -U user -d db -f QUICK_INSERT.sql
```

### Option 3: Complete Insert (INSERT_DUMMY_DATA.sql)
Full dataset with user orgs, accounts, transactions, withdrawals
```bash
psql -U user -d db -f INSERT_DUMMY_DATA.sql
```

## Expected Results

After running `RUN_THIS.sql`, you'll see:

```
✅ Wallet Updated
current_balance: 125.50
total_earned: 300.00
total_withdrawn: 174.50
can_withdraw: Yes ✓

✅ Withdrawals Inserted
status         | count | total_amount
---------------|-------|-------------
processing     |     1 |        75.50
pending        |     1 |        50.00
completed      |     2 |       174.50
failed         |     1 |        50.00
```

## Withdrawal Statuses Explained

- **completed** ✅ - Payout successful, money transferred to bank
- **pending** ⏳ - Waiting to be processed
- **processing** 🔄 - Currently being sent to Stripe
- **failed** ❌ - Payout failed, amount credited back to wallet

## Math Check

```
Total Earned:           $300.00
Total Withdrawn:        $174.50  (completed: $100 + $74.50)
Pending Withdrawals:    $125.50  (pending: $50 + processing: $75.50)
Current Balance:        $125.50
```

Balance Math: `$300 - $174.50 = $125.50` ✓

## Troubleshooting

### Error: "relation does not exist"
Make sure you've run the database schema:
```bash
psql -U user -d db -f /home/himanshu/Rival/rival-api/database/alter.sql
```

### Error: "violates foreign key constraint"
The developer wallet doesn't exist. First create it:
```sql
INSERT INTO tenant_schema.developer_wallets (
    id, organization_id, stripe_connect_account_id,
    balance, total_earned, total_withdrawn,
    onboarding_completed, payouts_enabled, charges_enabled,
    created_at, updated_at
) VALUES (
    gen_random_uuid(),
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    'acct_1Sd7YnE7kX8AZKi1',
    0, 0, 0, true, true, true, NOW(), NOW()
);
```

### Withdrawals Not Showing in Frontend
1. Check JWT token is valid
2. Verify organization ID matches
3. Check browser console for errors
4. Refresh the page

## Add More Data

### Add More Withdrawals

```sql
INSERT INTO tenant_schema.withdrawal_requests (
    id, developer_wallet_id, organization_id, amount, status,
    requested_at, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    (SELECT id FROM tenant_schema.developer_wallets WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'),
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    60.00,
    'completed',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 day',
    NOW();
```

### Add Transaction History

```sql
-- First create a user org
INSERT INTO tenant_schema.organizations (id, name, created_at, updated_at)
VALUES ('test-user-org', 'Test User Org', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Create user account
INSERT INTO tenant_schema.accounts (id, organization_id, account_balance, created_at, updated_at)
VALUES (gen_random_uuid(), 'test-user-org', 500.00, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Add transaction
INSERT INTO tenant_schema.function_execution_transactions (
    id, function_id, user_organization_id, developer_organization_id,
    user_account_id, developer_wallet_id, amount, platform_fee, net_amount,
    description, status, executed_at, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    'func-test-123',
    'test-user-org',
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'test-user-org' LIMIT 1),
    (SELECT id FROM tenant_schema.developer_wallets WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'),
    50.00, 0.00, 50.00,
    'Test transaction',
    'completed',
    NOW(), NOW(), NOW();
```

## Clean Up (Reset Data)

To remove all dummy data:

```sql
-- Delete withdrawals
DELETE FROM tenant_schema.withdrawal_requests
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

-- Reset wallet balance
UPDATE tenant_schema.developer_wallets
SET balance = 0, total_earned = 0, total_withdrawn = 0
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';
```

## Success!

You should now see a fully populated Withdrawal History in the frontend dashboard! 🎉
