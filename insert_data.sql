-- Simple insert for org: 437213bc-d871-401e-ac0e-aa8d4f321e70

-- Update wallet balance
UPDATE tenant_schema.developer_wallets
SET balance = 125.50, total_earned = 300.00, total_withdrawn = 174.50
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

-- Get wallet ID
\set wallet_id '(SELECT id FROM tenant_schema.developer_wallets WHERE organization_id = ''437213bc-d871-401e-ac0e-aa8d4f321e70'')'

INSERT INTO tenant_schema.withdrawal_requests (id, developer_wallet_id, organization_id, amount, status, stripe_payout_id, requested_at, completed_at, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    :wallet_id,
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    100.00,
    'completed',
    'po_test_completed_001',
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '10 days' + INTERVAL '2 hours',
    NOW() - INTERVAL '10 days',
    NOW()
);
INSERT INTO tenant_schema.withdrawal_requests (id, developer_wallet_id, organization_id, amount, status, stripe_payout_id, requested_at, completed_at, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    :wallet_id,
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    74.50,
    'completed',
    'po_test_completed_002',
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '5 days' + INTERVAL '3 hours',
    NOW() - INTERVAL '5 days',
    NOW()
);

INSERT INTO tenant_schema.withdrawal_requests (id, developer_wallet_id, organization_id, amount, status, requested_at, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    :wallet_id,
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    50.00,
    'pending',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '1 hour',
    NOW()
);

SELECT * FROM tenant_schema.withdrawal_requests WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70' ORDER BY requested_at DESC;
