#!/bin/bash

echo "Setting up Stripe Connect database..."

# Apply schema
psql -U postgres -d rival -f database/stripe_connect_schema.sql

if [ $? -eq 0 ]; then
    echo "✅ Database schema applied successfully"
    
    # Add test data
    psql -U postgres -d rival << SQL
-- Create test user organization with balance
INSERT INTO tenant_schema.organizations (id, name, bio)
VALUES ('user-org-test', 'Test User Organization', 'For testing payments')
ON CONFLICT (id) DO NOTHING;

INSERT INTO tenant_schema.accounts (id, organization_id, account_balance)
VALUES (uuid_generate_v4(), 'user-org-test', 500.00)
ON CONFLICT DO NOTHING;

-- Create test developer organization
INSERT INTO tenant_schema.organizations (id, name, bio)
VALUES ('dev-org-test', 'Test Developer Organization', 'For testing earnings')
ON CONFLICT (id) DO NOTHING;

SELECT 'Test data added successfully' as status;
SQL
    
    echo "✅ Test data added"
else
    echo "❌ Failed to apply schema"
    exit 1
fi
