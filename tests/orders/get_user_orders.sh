#!/bin/bash

token=$(curl -s -X POST http://localhost:8081/v1/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "hello@wasimmohammed.com",
           "password": "password123"
       }' | jq -r '.token')


curl -X GET http://localhost:8081/v1/api/orders/ \
    -H "Authorization: Bearer $token" \
