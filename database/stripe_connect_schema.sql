-- ================================
-- STRIPE CONNECT EXPRESS INTEGRATION
-- For Function Marketplace - Developer Payments
-- ================================

-- ================================
-- DEVELOPER WALLETS - Track developer earnings
-- ================================
CREATE TABLE IF NOT EXISTS tenant_schema.developer_wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL UNIQUE,
    stripe_connect_account_id VARCHAR(255) UNIQUE, -- Stripe Connect Express account ID
    balance DECIMAL(12,2) DEFAULT 0.00 NOT NULL,
    total_earned DECIMAL(12,2) DEFAULT 0.00 NOT NULL,
    total_withdrawn DECIMAL(12,2) DEFAULT 0.00 NOT NULL,
    onboarding_completed BOOLEAN DEFAULT FALSE,
    onboarding_url VARCHAR(500),
    payouts_enabled BOOLEAN DEFAULT FALSE,
    charges_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_developer_wallet_organization
        FOREIGN KEY (organization_id)
        REFERENCES tenant_schema.organizations (id) ON DELETE CASCADE
);

CREATE INDEX idx_developer_wallets_org_id ON tenant_schema.developer_wallets(organization_id);
CREATE INDEX idx_developer_wallets_stripe_account ON tenant_schema.developer_wallets(stripe_connect_account_id);

-- ================================
-- WITHDRAWAL REQUESTS - Track developer withdrawal requests
-- ================================
CREATE TABLE IF NOT EXISTS tenant_schema.withdrawal_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    developer_wallet_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    amount DECIMAL(12,2) NOT NULL CHECK (amount >= 50.00), -- Minimum $50
    status VARCHAR(50) DEFAULT 'pending' NOT NULL, -- pending, processing, completed, failed, rejected
    stripe_transfer_id VARCHAR(255),
    stripe_payout_id VARCHAR(255),
    failure_reason TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_withdrawal_developer_wallet
        FOREIGN KEY (developer_wallet_id)
        REFERENCES tenant_schema.developer_wallets (id) ON DELETE CASCADE,

    CONSTRAINT fk_withdrawal_organization
        FOREIGN KEY (organization_id)
        REFERENCES tenant_schema.organizations (id) ON DELETE CASCADE
);

CREATE INDEX idx_withdrawals_wallet_id ON tenant_schema.withdrawal_requests(developer_wallet_id);
CREATE INDEX idx_withdrawals_org_id ON tenant_schema.withdrawal_requests(organization_id);
CREATE INDEX idx_withdrawals_status ON tenant_schema.withdrawal_requests(status);

-- ================================
-- FUNCTION EXECUTION TRANSACTIONS - Track payments for function executions
-- ================================
CREATE TABLE IF NOT EXISTS tenant_schema.function_execution_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    function_id VARCHAR(255) NOT NULL,
    user_organization_id UUID NOT NULL, -- User who executed the function
    developer_organization_id UUID NOT NULL, -- Developer who owns the function
    user_account_id UUID NOT NULL, -- User's account (source)
    developer_wallet_id UUID NOT NULL, -- Developer's wallet (destination)
    amount DECIMAL(12,2) NOT NULL,
    platform_fee DECIMAL(12,2) DEFAULT 0.00, -- Future: platform commission
    net_amount DECIMAL(12,2) NOT NULL, -- amount - platform_fee
    description TEXT,
    status VARCHAR(50) DEFAULT 'completed' NOT NULL, -- completed, failed, refunded
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_transaction_user_account
        FOREIGN KEY (user_account_id)
        REFERENCES tenant_schema.accounts (id) ON DELETE CASCADE,

    CONSTRAINT fk_transaction_developer_wallet
        FOREIGN KEY (developer_wallet_id)
        REFERENCES tenant_schema.developer_wallets (id) ON DELETE CASCADE,

    CONSTRAINT fk_transaction_user_org
        FOREIGN KEY (user_organization_id)
        REFERENCES tenant_schema.organizations (id) ON DELETE CASCADE,

    CONSTRAINT fk_transaction_developer_org
        FOREIGN KEY (developer_organization_id)
        REFERENCES tenant_schema.organizations (id) ON DELETE CASCADE
);

CREATE INDEX idx_transactions_function_id ON tenant_schema.function_execution_transactions(function_id);
CREATE INDEX idx_transactions_user_org ON tenant_schema.function_execution_transactions(user_organization_id);
CREATE INDEX idx_transactions_dev_org ON tenant_schema.function_execution_transactions(developer_organization_id);
CREATE INDEX idx_transactions_executed_at ON tenant_schema.function_execution_transactions(executed_at DESC);

-- ================================
-- PLATFORM REVENUE - Track platform commission (optional, for future use)
-- ================================
CREATE TABLE IF NOT EXISTS tenant_schema.platform_revenue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    source VARCHAR(100) DEFAULT 'function_execution',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_revenue_transaction
        FOREIGN KEY (transaction_id)
        REFERENCES tenant_schema.function_execution_transactions (id) ON DELETE CASCADE
);

CREATE INDEX idx_platform_revenue_transaction ON tenant_schema.platform_revenue(transaction_id);
CREATE INDEX idx_platform_revenue_created_at ON tenant_schema.platform_revenue(created_at DESC);

-- ================================
-- ADD STRIPE CONNECT INFO TO ORGANIZATIONS (Optional enhancement)
-- ================================
-- Add column to track if organization has completed Stripe Connect onboarding
ALTER TABLE tenant_schema.organizations
ADD COLUMN IF NOT EXISTS is_developer BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS stripe_connect_onboarded BOOLEAN DEFAULT FALSE;

-- ================================
-- VIEWS FOR EASY QUERYING
-- ================================

-- Developer Earnings Summary View
CREATE OR REPLACE VIEW tenant_schema.v_developer_earnings AS
SELECT
    dw.id as wallet_id,
    dw.organization_id,
    o.name as organization_name,
    dw.stripe_connect_account_id,
    dw.balance as current_balance,
    dw.total_earned,
    dw.total_withdrawn,
    dw.onboarding_completed,
    dw.payouts_enabled,
    COUNT(DISTINCT fet.id) as total_transactions,
    COUNT(DISTINCT fet.function_id) as unique_functions_sold,
    dw.created_at,
    dw.updated_at
FROM
    tenant_schema.developer_wallets dw
    INNER JOIN tenant_schema.organizations o ON dw.organization_id = o.id
    LEFT JOIN tenant_schema.function_execution_transactions fet ON fet.developer_wallet_id = dw.id
GROUP BY
    dw.id, dw.organization_id, o.name;

-- Withdrawal History View
CREATE OR REPLACE VIEW tenant_schema.v_withdrawal_history AS
SELECT
    wr.id as withdrawal_id,
    wr.organization_id,
    o.name as organization_name,
    wr.amount,
    wr.status,
    wr.stripe_transfer_id,
    wr.stripe_payout_id,
    wr.failure_reason,
    wr.requested_at,
    wr.completed_at,
    dw.balance as current_wallet_balance
FROM
    tenant_schema.withdrawal_requests wr
    INNER JOIN tenant_schema.organizations o ON wr.organization_id = o.id
    INNER JOIN tenant_schema.developer_wallets dw ON wr.developer_wallet_id = dw.id
ORDER BY
    wr.requested_at DESC;

-- Transaction History View
CREATE OR REPLACE VIEW tenant_schema.v_transaction_history AS
SELECT
    fet.id as transaction_id,
    fet.function_id,
    f.function_name,
    fet.user_organization_id,
    uo.name as user_organization_name,
    fet.developer_organization_id,
    do.name as developer_organization_name,
    fet.amount,
    fet.platform_fee,
    fet.net_amount,
    fet.status,
    fet.executed_at
FROM
    tenant_schema.function_execution_transactions fet
    LEFT JOIN tenant_schema.organizations uo ON fet.user_organization_id = uo.id
    LEFT JOIN tenant_schema.organizations do ON fet.developer_organization_id = do.id
    LEFT JOIN tenant_schema.functions f ON fet.function_id = f.function_id::uuid
ORDER BY
    fet.executed_at DESC;

-- ================================
-- TRIGGERS FOR AUTOMATIC UPDATES
-- ================================

-- Update developer wallet updated_at timestamp
CREATE OR REPLACE FUNCTION tenant_schema.update_developer_wallet_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_developer_wallet_timestamp
    BEFORE UPDATE ON tenant_schema.developer_wallets
    FOR EACH ROW
    EXECUTE FUNCTION tenant_schema.update_developer_wallet_timestamp();

-- Update withdrawal request updated_at timestamp
CREATE OR REPLACE FUNCTION tenant_schema.update_withdrawal_request_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_withdrawal_request_timestamp
    BEFORE UPDATE ON tenant_schema.withdrawal_requests
    FOR EACH ROW
    EXECUTE FUNCTION tenant_schema.update_withdrawal_request_timestamp();

-- ================================
-- SAMPLE DATA FOR TESTING (Optional - Comment out for production)
-- ================================

-- Note: Run this only in development environment
-- COMMENT OUT BEFORE DEPLOYING TO PRODUCTION

/*
-- Create a test developer wallet
INSERT INTO tenant_schema.developer_wallets (organization_id, balance, total_earned, onboarding_completed, payouts_enabled)
VALUES (
    (SELECT id FROM tenant_schema.organizations LIMIT 1),
    0.00,
    0.00,
    FALSE,
    FALSE
);
*/

-- ================================
-- CLEANUP SCRIPT (Use with caution!)
-- ================================

/*
-- DANGER: This will delete all Stripe Connect data
-- Only use this for development/testing

DROP VIEW IF EXISTS tenant_schema.v_transaction_history;
DROP VIEW IF EXISTS tenant_schema.v_withdrawal_history;
DROP VIEW IF EXISTS tenant_schema.v_developer_earnings;

DROP TABLE IF EXISTS tenant_schema.platform_revenue CASCADE;
DROP TABLE IF EXISTS tenant_schema.function_execution_transactions CASCADE;
DROP TABLE IF EXISTS tenant_schema.withdrawal_requests CASCADE;
DROP TABLE IF EXISTS tenant_schema.developer_wallets CASCADE;

ALTER TABLE tenant_schema.organizations
DROP COLUMN IF EXISTS is_developer,
DROP COLUMN IF EXISTS stripe_connect_onboarded;
*/
