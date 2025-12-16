-- ============================================
-- INSERT DUMMY DATA FOR STRIPE CONNECT TESTING
-- Organization ID: 437213bc-d871-401e-ac0e-aa8d4f321e70
-- Stripe Account ID: acct_1Sd7YnE7kX8AZKi1
-- ============================================

-- Step 1: Ensure the organization exists
INSERT INTO tenant_schema.organizations (id, name, bio, created_at, updated_at)
VALUES (
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    'Developer Test Organization',
    'Test organization for Stripe Connect integration',
    NOW() - INTERVAL '30 days',
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    updated_at = NOW();

-- Step 2: Create or update the developer wallet with balance
INSERT INTO tenant_schema.developer_wallets (
    id,
    organization_id,
    stripe_connect_account_id,
    balance,
    total_earned,
    total_withdrawn,
    onboarding_completed,
    payouts_enabled,
    charges_enabled,
    created_at,
    updated_at
)
VALUES (
    gen_random_uuid(),
    '437213bc-d871-401e-ac0e-aa8d4f321e70',
    'acct_1Sd7YnE7kX8AZKi1',
    125.50,  -- Current balance (after earning and withdrawing)
    300.00,  -- Total earned from function executions
    174.50,  -- Total withdrawn (2 withdrawals: $100 + $74.50)
    true,
    true,
    true,
    NOW() - INTERVAL '25 days',
    NOW()
)
ON CONFLICT (organization_id) DO UPDATE SET
    stripe_connect_account_id = 'acct_1Sd7YnE7kX8AZKi1',
    balance = 125.50,
    total_earned = 300.00,
    total_withdrawn = 174.50,
    onboarding_completed = true,
    payouts_enabled = true,
    charges_enabled = true,
    updated_at = NOW();

-- Get the wallet ID for later use
DO $$
DECLARE
    wallet_id_var UUID;
BEGIN
    SELECT id INTO wallet_id_var
    FROM tenant_schema.developer_wallets
    WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

    -- Step 3: Create user organizations (who paid the developer)
    INSERT INTO tenant_schema.organizations (id, name, created_at, updated_at)
    VALUES
        ('user-org-001', 'Acme Corporation', NOW() - INTERVAL '20 days', NOW()),
        ('user-org-002', 'TechStart Inc', NOW() - INTERVAL '15 days', NOW()),
        ('user-org-003', 'Global Solutions Ltd', NOW() - INTERVAL '10 days', NOW())
    ON CONFLICT (id) DO NOTHING;

    -- Step 4: Create user accounts with balances
    INSERT INTO tenant_schema.accounts (id, organization_id, account_balance, created_at, updated_at)
    VALUES
        (gen_random_uuid(), 'user-org-001', 500.00, NOW() - INTERVAL '20 days', NOW()),
        (gen_random_uuid(), 'user-org-002', 750.00, NOW() - INTERVAL '15 days', NOW()),
        (gen_random_uuid(), 'user-org-003', 1000.00, NOW() - INTERVAL '10 days', NOW())
    ON CONFLICT DO NOTHING;

    -- Step 5: Insert function execution transactions (earnings history)
    -- Transaction 1: $50 from Acme Corporation
    INSERT INTO tenant_schema.function_execution_transactions (
        id,
        function_id,
        user_organization_id,
        developer_organization_id,
        user_account_id,
        developer_wallet_id,
        amount,
        platform_fee,
        net_amount,
        description,
        status,
        executed_at,
        created_at,
        updated_at
    )
    SELECT
        gen_random_uuid(),
        'func-image-optimizer-v1',
        'user-org-001',
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'user-org-001' LIMIT 1),
        wallet_id_var,
        50.00,
        0.00,
        50.00,
        'Function execution payment for image-optimizer (v1.0.0)',
        'completed',
        NOW() - INTERVAL '20 days' + INTERVAL '2 hours',
        NOW() - INTERVAL '20 days' + INTERVAL '2 hours',
        NOW() - INTERVAL '20 days' + INTERVAL '2 hours'
    WHERE wallet_id_var IS NOT NULL;

    -- Transaction 2: $75 from TechStart Inc
    INSERT INTO tenant_schema.function_execution_transactions (
        id,
        function_id,
        user_organization_id,
        developer_organization_id,
        user_account_id,
        developer_wallet_id,
        amount,
        platform_fee,
        net_amount,
        description,
        status,
        executed_at,
        created_at,
        updated_at
    )
    SELECT
        gen_random_uuid(),
        'func-pdf-generator-v2',
        'user-org-002',
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'user-org-002' LIMIT 1),
        wallet_id_var,
        75.00,
        0.00,
        75.00,
        'Function execution payment for pdf-generator (v2.1.0)',
        'completed',
        NOW() - INTERVAL '15 days' + INTERVAL '5 hours',
        NOW() - INTERVAL '15 days' + INTERVAL '5 hours',
        NOW() - INTERVAL '15 days' + INTERVAL '5 hours'
    WHERE wallet_id_var IS NOT NULL;

    -- Transaction 3: $25 from Global Solutions
    INSERT INTO tenant_schema.function_execution_transactions (
        id,
        function_id,
        user_organization_id,
        developer_organization_id,
        user_account_id,
        developer_wallet_id,
        amount,
        platform_fee,
        net_amount,
        description,
        status,
        executed_at,
        created_at,
        updated_at
    )
    SELECT
        gen_random_uuid(),
        'func-email-validator-v1',
        'user-org-003',
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'user-org-003' LIMIT 1),
        wallet_id_var,
        25.00,
        0.00,
        25.00,
        'Function execution payment for email-validator (v1.5.0)',
        'completed',
        NOW() - INTERVAL '10 days' + INTERVAL '3 hours',
        NOW() - INTERVAL '10 days' + INTERVAL '3 hours',
        NOW() - INTERVAL '10 days' + INTERVAL '3 hours'
    WHERE wallet_id_var IS NOT NULL;

    -- Transaction 4: $100 from Acme Corporation
    INSERT INTO tenant_schema.function_execution_transactions (
        id,
        function_id,
        user_organization_id,
        developer_organization_id,
        user_account_id,
        developer_wallet_id,
        amount,
        platform_fee,
        net_amount,
        description,
        status,
        executed_at,
        created_at,
        updated_at
    )
    SELECT
        gen_random_uuid(),
        'func-data-processor-v3',
        'user-org-001',
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'user-org-001' LIMIT 1),
        wallet_id_var,
        100.00,
        0.00,
        100.00,
        'Function execution payment for data-processor (v3.0.0)',
        'completed',
        NOW() - INTERVAL '8 days' + INTERVAL '1 hour',
        NOW() - INTERVAL '8 days' + INTERVAL '1 hour',
        NOW() - INTERVAL '8 days' + INTERVAL '1 hour'
    WHERE wallet_id_var IS NOT NULL;

    -- Transaction 5: $50 from TechStart Inc
    INSERT INTO tenant_schema.function_execution_transactions (
        id,
        function_id,
        user_organization_id,
        developer_organization_id,
        user_account_id,
        developer_wallet_id,
        amount,
        platform_fee,
        net_amount,
        description,
        status,
        executed_at,
        created_at,
        updated_at
    )
    SELECT
        gen_random_uuid(),
        'func-image-optimizer-v1',
        'user-org-002',
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        (SELECT id FROM tenant_schema.accounts WHERE organization_id = 'user-org-002' LIMIT 1),
        wallet_id_var,
        50.00,
        0.00,
        50.00,
        'Function execution payment for image-optimizer (v1.0.0)',
        'completed',
        NOW() - INTERVAL '5 days' + INTERVAL '4 hours',
        NOW() - INTERVAL '5 days' + INTERVAL '4 hours',
        NOW() - INTERVAL '5 days' + INTERVAL '4 hours'
    WHERE wallet_id_var IS NOT NULL;

    -- Step 6: Insert withdrawal requests
    -- Withdrawal 1: Completed withdrawal of $100
    INSERT INTO tenant_schema.withdrawal_requests (
        id,
        developer_wallet_id,
        organization_id,
        amount,
        status,
        stripe_transfer_id,
        stripe_payout_id,
        failure_reason,
        requested_at,
        completed_at,
        created_at,
        updated_at
    )
    VALUES (
        gen_random_uuid(),
        wallet_id_var,
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        100.00,
        'completed',
        'tr_1SdTestTransfer123',
        'po_1SdTestPayout456',
        NULL,
        NOW() - INTERVAL '12 days',
        NOW() - INTERVAL '12 days' + INTERVAL '3 hours',
        NOW() - INTERVAL '12 days',
        NOW() - INTERVAL '12 days' + INTERVAL '3 hours'
    );

    -- Withdrawal 2: Completed withdrawal of $74.50
    INSERT INTO tenant_schema.withdrawal_requests (
        id,
        developer_wallet_id,
        organization_id,
        amount,
        status,
        stripe_transfer_id,
        stripe_payout_id,
        failure_reason,
        requested_at,
        completed_at,
        created_at,
        updated_at
    )
    VALUES (
        gen_random_uuid(),
        wallet_id_var,
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        74.50,
        'completed',
        'tr_1SdTestTransfer789',
        'po_1SdTestPayout012',
        NULL,
        NOW() - INTERVAL '7 days',
        NOW() - INTERVAL '7 days' + INTERVAL '2 hours',
        NOW() - INTERVAL '7 days',
        NOW() - INTERVAL '7 days' + INTERVAL '2 hours'
    );

    -- Withdrawal 3: Failed withdrawal (for testing failed status)
    INSERT INTO tenant_schema.withdrawal_requests (
        id,
        developer_wallet_id,
        organization_id,
        amount,
        status,
        stripe_transfer_id,
        stripe_payout_id,
        failure_reason,
        requested_at,
        completed_at,
        created_at,
        updated_at
    )
    VALUES (
        gen_random_uuid(),
        wallet_id_var,
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        50.00,
        'failed',
        'tr_1SdTestTransferFailed',
        NULL,
        'insufficient_funds: Bank account has insufficient funds',
        NOW() - INTERVAL '4 days',
        NOW() - INTERVAL '4 days' + INTERVAL '1 hour',
        NOW() - INTERVAL '4 days',
        NOW() - INTERVAL '4 days' + INTERVAL '1 hour'
    );

    -- Withdrawal 4: Pending withdrawal (currently processing)
    INSERT INTO tenant_schema.withdrawal_requests (
        id,
        developer_wallet_id,
        organization_id,
        amount,
        status,
        stripe_transfer_id,
        stripe_payout_id,
        failure_reason,
        requested_at,
        completed_at,
        created_at,
        updated_at
    )
    VALUES (
        gen_random_uuid(),
        wallet_id_var,
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        75.50,
        'pending',
        NULL,
        NULL,
        NULL,
        NOW() - INTERVAL '2 hours',
        NULL,
        NOW() - INTERVAL '2 hours',
        NOW() - INTERVAL '2 hours'
    );

    -- Withdrawal 5: Processing withdrawal
    INSERT INTO tenant_schema.withdrawal_requests (
        id,
        developer_wallet_id,
        organization_id,
        amount,
        status,
        stripe_transfer_id,
        stripe_payout_id,
        failure_reason,
        requested_at,
        completed_at,
        created_at,
        updated_at
    )
    VALUES (
        gen_random_uuid(),
        wallet_id_var,
        '437213bc-d871-401e-ac0e-aa8d4f321e70',
        50.00,
        'processing',
        'tr_1SdTestTransferProcessing',
        NULL,
        NULL,
        NOW() - INTERVAL '30 minutes',
        NULL,
        NOW() - INTERVAL '30 minutes',
        NOW()
    );

END $$;

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- Check developer wallet
SELECT
    'Developer Wallet' as table_name,
    organization_id,
    stripe_connect_account_id,
    balance,
    total_earned,
    total_withdrawn,
    onboarding_completed,
    payouts_enabled
FROM tenant_schema.developer_wallets
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

-- Check transactions (earnings)
SELECT
    'Function Execution Transactions' as table_name,
    COUNT(*) as count,
    SUM(amount) as total_amount,
    SUM(net_amount) as total_net_amount
FROM tenant_schema.function_execution_transactions
WHERE developer_organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

-- Check withdrawals
SELECT
    'Withdrawal Requests' as table_name,
    status,
    COUNT(*) as count,
    SUM(amount) as total_amount
FROM tenant_schema.withdrawal_requests
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'
GROUP BY status
ORDER BY status;

-- Show recent transactions
SELECT
    id,
    function_id,
    amount,
    net_amount,
    status,
    executed_at,
    description
FROM tenant_schema.function_execution_transactions
WHERE developer_organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'
ORDER BY executed_at DESC;

-- Show all withdrawals
SELECT
    id,
    amount,
    status,
    stripe_payout_id,
    failure_reason,
    requested_at,
    completed_at
FROM tenant_schema.withdrawal_requests
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'
ORDER BY requested_at DESC;

-- Calculate expected balance
-- Expected: total_earned - total_withdrawn - pending_withdrawals = available_balance
SELECT
    dw.balance as current_balance,
    dw.total_earned,
    dw.total_withdrawn,
    COALESCE(pending.pending_amount, 0) as pending_withdrawals,
    (dw.total_earned - dw.total_withdrawn - COALESCE(pending.pending_amount, 0)) as calculated_balance
FROM tenant_schema.developer_wallets dw
LEFT JOIN (
    SELECT
        developer_wallet_id,
        SUM(amount) as pending_amount
    FROM tenant_schema.withdrawal_requests
    WHERE status IN ('pending', 'processing')
    GROUP BY developer_wallet_id
) pending ON pending.developer_wallet_id = dw.id
WHERE dw.organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';
