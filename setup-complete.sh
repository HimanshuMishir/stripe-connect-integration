#!/bin/bash

echo "🚀 Complete Stripe Connect Setup"
echo "================================="
echo ""

# Load environment
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    exit 1
fi

source .env

# Step 1: Check Stripe keys
echo "Step 1: Checking Stripe Configuration"
echo "--------------------------------------"
if [[ $STRIPE_SECRET_KEY == sk_live_* ]]; then
    echo "⚠️  WARNING: Using LIVE Stripe keys!"
    echo "   Please switch to TEST keys for development"
    echo "   Get them from: https://dashboard.stripe.com/test/apikeys"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
elif [[ $STRIPE_SECRET_KEY == sk_test_* ]]; then
    echo "✅ Using TEST keys (correct)"
else
    echo "❌ Invalid Stripe key format"
    exit 1
fi
echo ""

# Step 2: Test Database Connection
echo "Step 2: Testing Database Connection"
echo "------------------------------------"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    echo "   Check credentials in .env"
    exit 1
fi
echo ""

# Step 3: Apply Database Schema
echo "Step 3: Applying Database Schema"
echo "---------------------------------"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f database/stripe_connect_schema.sql > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Schema applied successfully"
else
    echo "⚠️  Schema might already exist (this is okay)"
fi
echo ""

# Step 4: Add Test Data
echo "Step 4: Adding Test Data"
echo "------------------------"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << 'SQL'
-- Fixed UUIDs for test organizations (easy to remember and reference)
-- User Org:  11111111-1111-1111-1111-111111111111
-- Dev Org:   22222222-2222-2222-2222-222222222222

-- Test User Organization (will execute functions)
INSERT INTO tenant_schema.organizations (id, name, bio)
VALUES ('11111111-1111-1111-1111-111111111111', 'Test User Organization', 'For testing payments')
ON CONFLICT (id) DO NOTHING;

-- User Account with $500 balance
INSERT INTO tenant_schema.accounts (id, organization_id, account_balance)
SELECT uuid_generate_v4(), '11111111-1111-1111-1111-111111111111', 500.00
WHERE NOT EXISTS (
    SELECT 1 FROM tenant_schema.accounts WHERE organization_id = '11111111-1111-1111-1111-111111111111'
);

-- Test Developer Organization (will receive payments)
INSERT INTO tenant_schema.organizations (id, name, bio)
VALUES ('22222222-2222-2222-2222-222222222222', 'Test Developer Organization', 'For testing earnings')
ON CONFLICT (id) DO NOTHING;

-- Verify
SELECT
    o.name,
    COALESCE(a.account_balance, 0) as balance
FROM tenant_schema.organizations o
LEFT JOIN tenant_schema.accounts a ON o.id = a.organization_id
WHERE o.id IN ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222');
SQL
echo "✅ Test data added"
echo ""

# Step 5: Check for Stripe CLI
echo "Step 5: Checking Stripe CLI"
echo "----------------------------"
if command -v stripe &> /dev/null; then
    echo "✅ Stripe CLI is installed"
    echo ""
    echo "To get your webhook secret, run this in a new terminal:"
    echo "  stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect"
    echo ""
    echo "Then copy the 'whsec_...' secret to your .env file"
else
    echo "⚠️  Stripe CLI not installed"
    echo ""
    echo "Install it with:"
    echo "  Mac: brew install stripe/stripe-cli/stripe"
    echo "  Linux: See https://stripe.com/docs/stripe-cli"
fi
echo ""

# Step 6: Summary
echo "================================="
echo "📋 Setup Complete!"
echo "================================="
echo ""
echo "Test Organizations Created:"
echo "  User Org ID: 11111111-1111-1111-1111-111111111111 (Balance: \$500)"
echo "  Dev Org ID:  22222222-2222-2222-2222-222222222222 (Balance: \$0)"
echo ""
echo "Next Steps:"
echo ""
echo "1. Get webhook secret:"
echo "   Terminal 1: stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect"
echo "   Copy the whsec_... secret to .env"
echo ""
echo "2. Start the backend:"
echo "   Terminal 2: go run main.go"
echo ""
echo "3. Start the frontend:"
echo "   Terminal 3: cd frontend && npm install && npm run dev"
echo ""
echo "4. Open browser: http://localhost:3000"
echo ""
echo "5. Test the flow:"
echo "   - Developer Dashboard: Use org ID: 22222222-2222-2222-2222-222222222222"
echo "   - User Dashboard: Use org ID: 11111111-1111-1111-1111-111111111111"
echo "   - Developer Dashboard: Request withdrawal when balance >= \$50"
echo ""
