-- ============================================
-- RUN THIS SCRIPT TO INSERT DUMMY DATA
-- Organization: 437213bc-d871-401e-ac0e-aa8d4f321e70
-- ============================================

-- First, let's get the wallet ID
DO $$
DECLARE
    v_wallet_id UUID;
BEGIN
    -- Get the wallet ID
    SELECT id INTO v_wallet_id
    FROM tenant_schema.developer_wallets
    WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70';

    -- Update wallet balance
    UPDATE tenant_schema.developer_wallets
    SET
        balance = 125.50,
        total_earned = 300.00,
        total_withdrawn = 174.50,
        updated_at = NOW()
    WHERE id = v_wallet_id;

    -- Insert 5 withdrawal requests with different statuses

    -- Withdrawal 1: Completed $100
    INSERT INTO tenant_schema.withdrawal_requests (
    COALESCE(failure_reason, 'N/A') as failure_reason,
    requested_at,
    completed_at
FROM tenant_schema.withdrawal_requests
WHERE organization_id = '437213bc-d871-401e-ac0e-aa8d4f321e70'
ORDER BY requested_at DESC;
