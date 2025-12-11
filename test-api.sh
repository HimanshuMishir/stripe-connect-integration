#!/bin/bash

# API Testing Script for Stripe Connect

BASE_URL="http://localhost:8080"
ORG_ID="test-org-123"

echo "🧪 Stripe Connect API Testing"
echo "==============================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo "Test 1: Health Check"
echo "--------------------"
curl -s "$BASE_URL/health" | jq '.'
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Server is running${NC}"
else
    echo -e "${RED}❌ Server is not responding${NC}"
    exit 1
fi
echo ""

# Test 2: Create Connect Account (Onboarding)
echo "Test 2: Create Stripe Connect Account"
echo "--------------------------------------"
ONBOARD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/connect/onboard" \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: $ORG_ID" \
  -d '{
    "refresh_url": "http://localhost:3000/connect/refresh",
    "return_url": "http://localhost:3000/connect/complete"
  }')

echo "$ONBOARD_RESPONSE" | jq '.'

if echo "$ONBOARD_RESPONSE" | jq -e '.onboarding_url' > /dev/null; then
    echo -e "${GREEN}✅ Onboarding link created${NC}"
    ONBOARDING_URL=$(echo "$ONBOARD_RESPONSE" | jq -r '.onboarding_url')
    echo -e "${YELLOW}📝 Open this URL to complete onboarding:${NC}"
    echo "$ONBOARDING_URL"
else
    echo -e "${RED}❌ Failed to create onboarding link${NC}"
fi
echo ""

# Test 3: Get Connect Account Status
echo "Test 3: Get Account Status"
echo "--------------------------"
STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/status" \
  -H "X-Organization-ID: $ORG_ID")

echo "$STATUS_RESPONSE" | jq '.'

if echo "$STATUS_RESPONSE" | jq -e '.account_id' > /dev/null; then
    echo -e "${GREEN}✅ Account status retrieved${NC}"
else
    echo -e "${YELLOW}⚠️  Account not found (need to complete onboarding first)${NC}"
fi
echo ""

# Test 4: Get Wallet Balance
echo "Test 4: Get Wallet Balance"
echo "--------------------------"
BALANCE_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/wallet/balance" \
  -H "X-Organization-ID: $ORG_ID")

echo "$BALANCE_RESPONSE" | jq '.'

if echo "$BALANCE_RESPONSE" | jq -e '.balance' > /dev/null; then
    echo -e "${GREEN}✅ Wallet balance retrieved${NC}"
else
    echo -e "${YELLOW}⚠️  Wallet not found (need to complete onboarding first)${NC}"
fi
echo ""

# Test 5: Simulate Function Execution Payment (requires setup)
echo "Test 5: Function Execution Payment"
echo "-----------------------------------"
echo -e "${YELLOW}⚠️  This requires:${NC}"
echo "   1. A user organization with account balance"
echo "   2. A developer with completed onboarding"
echo ""
echo "Skipping for now. To test manually:"
echo "curl -X POST $BASE_URL/api/connect/payments/execute \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'X-Organization-ID: user-org-id' \\"
echo "  -d '{\"function_id\": \"func-123\", \"amount\": 5.00}'"
echo ""

# Test 6: Request Withdrawal (requires balance)
echo "Test 6: Request Withdrawal"
echo "--------------------------"
echo -e "${YELLOW}⚠️  This requires:${NC}"
echo "   1. Completed onboarding"
echo "   2. Balance >= \$50.00"
echo ""
echo "Skipping for now. To test manually:"
echo "curl -X POST $BASE_URL/api/connect/withdrawals/request \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'X-Organization-ID: dev-org-id' \\"
echo "  -d '{\"amount\": 50.00}'"
echo ""

echo "==============================="
echo "📊 Test Summary"
echo "==============================="
echo ""
echo "To complete testing:"
echo "1. Complete the onboarding using the URL above"
echo "2. Add test data to database (see below)"
echo "3. Test function execution payment"
echo "4. Test withdrawal"
echo ""
echo "Add test data:"
echo "psql -U postgres -d rival << EOF"
echo "-- Add user organization with balance"
echo "INSERT INTO tenant_schema.organizations (id, name)"
echo "VALUES ('user-org-1', 'Test User Org')"
echo "ON CONFLICT (id) DO NOTHING;"
echo ""
echo "INSERT INTO tenant_schema.accounts (id, organization_id, account_balance)"
echo "VALUES (uuid_generate_v4(), 'user-org-1', 100.00)"
echo "ON CONFLICT DO NOTHING;"
echo "EOF"
echo ""
