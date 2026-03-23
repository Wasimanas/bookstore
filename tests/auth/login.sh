#!/bin/bash

URL="http://localhost:8081/v1/api/auth/login"

curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
}'
