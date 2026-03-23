#!/bin/bash

BOOK_ID=1

curl -X PATCH http://localhost:8081/v1/api/book/$BOOK_ID \
-H "Content-Type: application/json" \
-d '{
  "title": "The Go Programming Language - Updated",
  "price": 50.00
}'
