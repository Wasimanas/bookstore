token=$(curl -s -X POST http://localhost:8081/v1/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "hello@wasimmohammed.com",
    "password": "password123"
  }' | jq -r '.token')

echo "Token: $token"
echo ""

curl -X POST http://localhost:8081/v1/api/internal/orders/refund \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer '"$token"'" \
  -d '{
    "order_id": 7,
    "items": [
      { "id": 8, "quantity": 1 }
    ],
    "reason": "Customer returned items"
  }'
