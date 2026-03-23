token=$(curl -s -X POST http://localhost:8081/v1/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "hello@wasimmohammed.com",
           "password": "password123"
       }' | jq -r '.token')

echo "Token: $token"

echo "\n"

curl -X GET http://localhost:8081/v1/api/orders/ \
     -H "Authorization: Bearer $token" | jq .

echo "\n"


curl -X POST http://localhost:8080/v1/api/payments/pay/7 \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $token" \
     -d '{
           "success_url": "http://localhost:8080/success",
           "cancel_url": "http://localhost:8080/cancel"
         }'
