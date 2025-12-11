#!/bin/bash

# End-to-End Integration Test Script

BASE_URL="http://localhost:8081"
USER_ORG="11111111-1111-1111-1111-111111111111"
DEV_ORG="22222222-2222-2222-2222-222222222222"

echo "🧪 Stripe Connect End-to-End Integration Test"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo "Test 1: Health Check"
echo "--------------------"
HEALTH=$(curl -s $BASE_URL/health)
echo "$HEALTH" | jq '.'
if echo "$HEALTH" | jq -e '.status == "healthy"' > /dev/null; then
    echo -e "${GREEN}✅ Health check passed${NC}"
else
    echo -e "${RED}❌ Health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: Create Developer Connect Account
echo "Test 2: Create Stripe Connect Account (Developer Onboarding)"
echo "-------------------------------------------------------------"
ONBOARD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/connect/onboard" \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: $DEV_ORG" \
  -d '{
    "refresh_url": "http://localhost:3000/connect/refresh",
    "return_url": "http://localhost:3000/connect/complete"
  }')

echo "$ONBOARD_RESPONSE" | jq '.'

if echo "$ONBOARD_RESPONSE" | jq -e '.onboarding_url' > /dev/null; then
    echo -e "${GREEN}✅ Onboarding link created${NC}"
    ONBOARDING_URL=$(echo "$ONBOARD_RESPONSE" | jq -r '.onboarding_url')
    echo -e "${YELLOW}📝 Onboarding URL:${NC}"
    echo "$ONBOARDING_URL"
    echo ""
    echo -e "${YELLOW}⚠️  Manual Step Required:${NC}"
    echo "1. Open this URL in your browser"
    echo "2. Complete Stripe Connect onboarding"
    echo "3. Press Enter to continue..."
    read -p ""
else
    echo -e "${YELLOW}⚠️  Onboarding link creation response:${NC}"
    echo "$ONBOARD_RESPONSE" | jq '.'
fi
echo ""

# Test 3: Check Developer Account Status
echo "Test 3: Check Developer Account Status"
echo "---------------------------------------"
STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/status" \
  -H "X-Organization-ID: $DEV_ORG")

echo "$STATUS_RESPONSE" | jq '.'

if echo "$STATUS_RESPONSE" | jq -e '.account_id' > /dev/null; then
    echo -e "${GREEN}✅ Developer account exists${NC}"

    ONBOARDING_COMPLETED=$(echo "$STATUS_RESPONSE" | jq -r '.onboarding_completed')
    PAYOUTS_ENABLED=$(echo "$STATUS_RESPONSE" | jq -r '.payouts_enabled')

    if [ "$ONBOARDING_COMPLETED" = "true" ]; then
        echo -e "${GREEN}✅ Onboarding completed${NC}"
    else
        echo -e "${YELLOW}⚠️  Onboarding not completed yet${NC}"
    fi

    if [ "$PAYOUTS_ENABLED" = "true" ]; then
        echo -e "${GREEN}✅ Payouts enabled${NC}"
    else
        echo -e "${YELLOW}⚠️  Payouts not enabled yet${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Developer account not fully set up${NC}"
fi
echo ""

# Test 4: Check Developer Wallet Balance
echo "Test 4: Check Developer Wallet Balance"
echo "---------------------------------------"
WALLET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/wallet/balance" \
  -H "X-Organization-ID: $DEV_ORG")

echo "$WALLET_RESPONSE" | jq '.'

if echo "$WALLET_RESPONSE" | jq -e '.balance' > /dev/null; then
    BALANCE=$(echo "$WALLET_RESPONSE" | jq -r '.balance')
    echo -e "${GREEN}✅ Wallet balance: \$$BALANCE${NC}"
else
    echo -e "${YELLOW}⚠️  Wallet not found or error${NC}"
fi
echo ""

# Test 5: Check User Account Balance (via database)
echo "Test 5: Check User Account Balance"
echo "-----------------------------------"
USER_BALANCE=$(PGPASSWORD=postgres psql -h localhost -U postgres -d cortexone_dev_db -t -c "
    SELECT COALESCE(account_balance, 0)
    FROM tenant_schema.accounts
    WHERE organization_id = '$USER_ORG';
" 2>/dev/null)

if [ ! -z "$USER_BALANCE" ]; then
    USER_BALANCE=$(echo $USER_BALANCE | xargs)
    echo -e "${GREEN}✅ User balance: \$$USER_BALANCE${NC}"
else
    echo -e "${RED}❌ Failed to get user balance${NC}"
fi
echo ""

# Test 6: Execute Function Payment
echo "Test 6: Execute Function Payment (User → Developer)"
echo "-----------------------------------------------------"
echo -e "${YELLOW}Testing payment of \$5.00...${NC}"

PAYMENT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/connect/payments/execute" \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: $USER_ORG" \
  -d "{
    \"function_id\": \"func-test-123\",
    \"amount\": 5.00,
    \"developer_organization_id\": \"$DEV_ORG\"
  }")

echo "$PAYMENT_RESPONSE" | jq '.'

if echo "$PAYMENT_RESPONSE" | jq -e '.transaction_id' > /dev/null; then
    echo -e "${GREEN}✅ Payment successful!${NC}"

    USER_NEW_BALANCE=$(echo "$PAYMENT_RESPONSE" | jq -r '.user_balance')
    DEV_NEW_BALANCE=$(echo "$PAYMENT_RESPONSE" | jq -r '.developer_balance')

    echo -e "${GREEN}User balance: \$$USER_BALANCE → \$$USER_NEW_BALANCE${NC}"
    echo -e "${GREEN}Developer balance: \$0.00 → \$$DEV_NEW_BALANCE${NC}"
else
    echo -e "${RED}❌ Payment failed${NC}"
    echo "$PAYMENT_RESPONSE"
fi
echo ""

# Test 7: Execute Multiple Payments (to build up balance for withdrawal)
echo "Test 7: Execute 10 More Payments (Build up to \$50)"
echo "-----------------------------------------------------"

for i in {1..10}; do
    echo -n "Payment $i... "
    RESULT=$(curl -s -X POST "$BASE_URL/api/connect/payments/execute" \
      -H "Content-Type: application/json" \
      -H "X-Organization-ID: $USER_ORG" \
      -d "{
        \"function_id\": \"func-test-123\",
        \"amount\": 5.00,
        \"developer_organization_id\": \"$DEV_ORG\"
      }")

    if echo "$RESULT" | jq -e '.transaction_id' > /dev/null; then
        echo -e "${GREEN}✅${NC}"
    else
        echo -e "${RED}❌${NC}"
    fi
    sleep 0.5
done
echo ""

# Test 8: Check Updated Balances
echo "Test 8: Check Updated Balances"
echo "-------------------------------"
WALLET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/wallet/balance" \
  -H "X-Organization-ID: $DEV_ORG")

DEV_BALANCE=$(echo "$WALLET_RESPONSE" | jq -r '.balance')
echo -e "${GREEN}Developer balance: \$$DEV_BALANCE${NC}"

if (( $(echo "$DEV_BALANCE >= 50" | bc -l) )); then
    echo -e "${GREEN}✅ Developer has enough to withdraw (\$50 minimum)${NC}"
else
    echo -e "${YELLOW}⚠️  Developer needs \$$(echo "50 - $DEV_BALANCE" | bc) more to withdraw${NC}"
fi
echo ""

# Test 9: Request Withdrawal
echo "Test 9: Request Withdrawal (\$50)"
echo "----------------------------------"

if (( $(echo "$DEV_BALANCE >= 50" | bc -l) )); then
    WITHDRAWAL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/connect/withdrawals/request" \
      -H "Content-Type: application/json" \
      -H "X-Organization-ID: $DEV_ORG" \
      -d '{
        "amount": 50.00
      }')

    echo "$WITHDRAWAL_RESPONSE" | jq '.'

    if echo "$WITHDRAWAL_RESPONSE" | jq -e '.withdrawal_id' > /dev/null; then
        echo -e "${GREEN}✅ Withdrawal request created${NC}"
        echo -e "${YELLOW}💰 Money will be transferred to bank account in 2-3 business days${NC}"
    else
        echo -e "${RED}❌ Withdrawal failed${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Skipping withdrawal test - insufficient balance${NC}"
fi
echo ""

# Test 10: Check Transaction History
echo "Test 10: Check Transaction History"
echo "-----------------------------------"
TX_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/wallet/transactions?limit=5" \
  -H "X-Organization-ID: $DEV_ORG")

echo "$TX_RESPONSE" | jq '.'

TX_COUNT=$(echo "$TX_RESPONSE" | jq '.transactions | length')
if [ "$TX_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Found $TX_COUNT transaction(s)${NC}"
else
    echo -e "${YELLOW}⚠️  No transactions found${NC}"
fi
echo ""

# Test 11: Check Withdrawal History
echo "Test 11: Check Withdrawal History"
echo "----------------------------------"
WD_RESPONSE=$(curl -s -X GET "$BASE_URL/api/connect/withdrawals/history?limit=5" \
  -H "X-Organization-ID: $DEV_ORG")

echo "$WD_RESPONSE" | jq '.'

WD_COUNT=$(echo "$WD_RESPONSE" | jq '.withdrawals | length')
if [ "$WD_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Found $WD_COUNT withdrawal(s)${NC}"
else
    echo -e "${YELLOW}⚠️  No withdrawals found${NC}"
fi
echo ""

# Final Summary
echo "=============================================="
echo "🎯 End-to-End Integration Test Summary"
echo "=============================================="
echo ""
echo -e "${GREEN}✅ Services Running:${NC}"
echo "   - Backend API: http://localhost:8081"
echo "   - Frontend UI: http://localhost:3000"
echo ""
echo -e "${GREEN}✅ Test Organizations:${NC}"
echo "   - User Org: $USER_ORG"
echo "   - Dev Org:  $DEV_ORG"
echo ""
echo -e "${GREEN}✅ Features Tested:${NC}"
echo "   - Health check"
echo "   - Developer onboarding (Stripe Connect)"
echo "   - Account status checking"
echo "   - Wallet balance tracking"
echo "   - Function execution payments"
echo "   - Withdrawal requests"
echo "   - Transaction history"
echo ""
echo -e "${YELLOW}📝 Next Steps:${NC}"
echo "1. Open http://localhost:3000 in your browser"
echo "2. Test the UI with the organization IDs above"
echo "3. Complete developer onboarding if not done"
echo "4. Watch webhook events in your Stripe CLI terminal"
echo ""
echo "=============================================="
