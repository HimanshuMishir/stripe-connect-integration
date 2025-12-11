#!/bin/bash

# Stripe Connect Test Setup Script
# This script helps you verify your Stripe configuration

echo "🎯 Stripe Connect Test Setup"
echo "=============================="
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please copy .env.example to .env first:"
    echo "cp .env.example .env"
    exit 1
fi

# Check Stripe keys
echo "📋 Checking Stripe configuration..."
source .env

if [[ $STRIPE_SECRET_KEY == sk_live_* ]]; then
    echo "⚠️  WARNING: You're using LIVE Stripe keys!"
    echo "   For testing, please use TEST keys (sk_test_...)"
    echo ""
    echo "   Get test keys from: https://dashboard.stripe.com/test/apikeys"
    echo ""
fi

if [[ $STRIPE_SECRET_KEY == sk_test_* ]]; then
    echo "✅ Using TEST Stripe keys (correct for development)"
fi

if [[ $STRIPE_WEBHOOK_SECRET == "whsec_your_webhook_secret_here" ]]; then
    echo "❌ Webhook secret not configured yet"
    echo ""
    echo "To get your webhook secret:"
    echo ""
    echo "Option 1: Using Stripe CLI (Recommended for local development)"
    echo "  1. Install Stripe CLI: brew install stripe/stripe-cli/stripe"
    echo "  2. Login: stripe login"
    echo "  3. Run: stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect"
    echo "  4. Copy the 'whsec_...' secret it shows"
    echo "  5. Update .env with that secret"
    echo ""
    echo "Option 2: Using Stripe Dashboard (For production)"
    echo "  1. Go to: https://dashboard.stripe.com/webhooks"
    echo "  2. Click 'Add endpoint'"
    echo "  3. URL: https://your-domain.com/api/webhooks/stripe-connect"
    echo "  4. Select events: account.updated, payout.paid, payout.failed"
    echo "  5. Copy the signing secret"
    echo ""
else
    echo "✅ Webhook secret configured"
fi

# Check database
echo ""
echo "📋 Checking database configuration..."
echo "   Host: $DB_HOST"
echo "   Port: $DB_PORT"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"

# Test database connection
echo ""
echo "Testing database connection..."
if command -v psql &> /dev/null; then
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1" &> /dev/null
    if [ $? -eq 0 ]; then
        echo "✅ Database connection successful"

        # Check if tables exist
        TABLE_CHECK=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'tenant_schema' AND table_name = 'developer_wallets';")
        if [ "$TABLE_CHECK" -eq 1 ]; then
            echo "✅ Stripe Connect tables exist"
        else
            echo "❌ Stripe Connect tables not found!"
            echo "   Run: psql -U $DB_USER -d $DB_NAME -f database/stripe_connect_schema.sql"
        fi
    else
        echo "❌ Database connection failed"
        echo "   Check your database credentials in .env"
    fi
else
    echo "⚠️  psql command not found, skipping database test"
fi

echo ""
echo "=============================="
echo "📝 Next Steps:"
echo ""

if [[ $STRIPE_SECRET_KEY == sk_live_* ]]; then
    echo "1. ⚠️  IMPORTANT: Replace LIVE keys with TEST keys in .env"
fi

if [[ $STRIPE_WEBHOOK_SECRET == "whsec_your_webhook_secret_here" ]]; then
    echo "2. Set up webhook secret (see instructions above)"
fi

echo "3. Start the server: go run main.go"
echo "4. In another terminal, forward webhooks: stripe listen --forward-to localhost:8080/api/webhooks/stripe-connect"
echo "5. Open frontend: cd frontend && npm run dev"
echo "6. Visit: http://localhost:3000"
echo ""
echo "=============================="
