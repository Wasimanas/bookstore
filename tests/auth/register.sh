#!/bin/bash

URL="http://localhost:8081/v1/api/auth/register"

curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "user",
    "last_name": "user",
    "email": "user@example.com",
    "password": "password123"
  }'

