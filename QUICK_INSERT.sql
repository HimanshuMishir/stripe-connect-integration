-- Quick Insert Script for Organization: 437213bc-d871-401e-ac0e-aa8d4f321e70
-- Run this to get withdrawal list and transactions visible immediately

-- 1. Update wallet with balance
UPDATE tenant_schema.developer_wallets
SET
    balance = 125.50,
    total_earned = 300.00,
    total_withdrawn = 174.50,
    updated_at = NOW()
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

-- 2. Insert user organizations
INSERT INTO tenant_schema.organizations (id, name, created_at, updated_at) VALUES
    ('user-org-001', 'Acme Corporation', NOW() - INTERVAL '20 days', NOW()),
    ('user-org-002', 'TechStart Inc', NOW() - INTERVAL '15 days', NOW())
ON CONFLICT (id) DO NOTHING;

-- 3. Insert user accounts
INSERT INTO tenant_schema.accounts (id, organization_id, account_balance, created_at, updated_at) VALUES
    (gen_random_uuid(), 'user-org-001', 500.00, NOW() - INTERVAL '20 days', NOW()),
    (gen_random_uuid(), 'user-org-002', 750.00, NOW() - INTERVAL '15 days', NOW())
ON CONFLICT DO NOTHING;

-- 4. Insert transaction history (5 transactions = $300 total)
INSERT INTO tenant_schema.function_execution_transactions (
    id, function_id, user_organization_id, developer_organization_id,
    user_account_id, developer_wallet_id, amount, platform_fee, net_amount,
    description, status, executed_at, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    'func-test-' || gs,
    CASE WHEN gs % 2 = 0 THEN 'user-org-001' ELSE 'user-org-002' END,
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'user-org-001' LIMIT 1),
    (SELECT id FROM tenant_schema.developer_wallets WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'),
    CASE
        WHEN gs = 1 THEN 50.00
        WHEN gs = 2 THEN 75.00
        WHEN gs = 3 THEN 25.00
        WHEN gs = 4 THEN 100.00
        ELSE 50.00
    END,
    0.00,
    CASE
        WHEN gs = 1 THEN 50.00
        WHEN gs = 2 THEN 75.00
        WHEN gs = 3 THEN 25.00
        WHEN gs = 4 THEN 100.00
        ELSE 50.00
    END,
    'Function execution payment #' || gs,
    'completed',
    NOW() - INTERVAL '20 days' + (gs * INTERVAL '3 days'),
    NOW() - INTERVAL '20 days' + (gs * INTERVAL '3 days'),
    NOW()
FROM generate_series(1, 5) gs;

-- 5. Insert withdrawal history (5 withdrawals with different statuses)
INSERT INTO tenant_schema.withdrawal_requests (
    id, developer_wallet_id, organization_id, amount, status,
    stripe_payout_id, failure_reason, requested_at, completed_at, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    (SELECT id FROM tenant_schema.developer_wallets WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'),
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    amt,
    st,
    CASE WHEN st = 'completed' THEN 'po_test_' || random()::text ELSE NULL END,
    CASE WHEN st = 'failed' THEN 'insufficient_funds: Test failure' ELSE NULL END,
    req_at,
    CASE WHEN st IN ('completed', 'failed') THEN req_at + INTERVAL '2 hours' ELSE NULL END,
    req_at,
    NOW()
FROM (VALUES
    (100.00, 'completed'::text, NOW() - INTERVAL '12 days'),
    (74.50, 'completed'::text, NOW() - INTERVAL '7 days'),
    (50.00, 'failed'::text, NOW() - INTERVAL '4 days'),
    (75.50, 'pending'::text, NOW() - INTERVAL '2 hours'),
    (50.00, 'processing'::text, NOW() - INTERVAL '30 minutes')
) AS withdrawals(amt, st, req_at);

-- Verification: Show current state
SELECT
    'Current Balance: $' || balance as info,
    'Total Earned: $' || total_earned,
    'Total Withdrawn: $' || total_withdrawn,
    'Can Withdraw: ' || CASE WHEN balance >= 50 THEN 'Yes' ELSE 'No' END
FROM tenant_schema.developer_wallets
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

SELECT
    COUNT(*) || ' transactions, Total: $' || SUM(amount) as transaction_summary
FROM tenant_schema.function_execution_transactions
WHERE developer_organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

SELECT
    status,
    COUNT(*) as count,
    '$' || SUM(amount) as total
FROM tenant_schema.withdrawal_requests
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'
GROUP BY status;
