#!/bin/bash

curl -X POST http://localhost:8081/v1/api/book \
    -H "Content-Type: application/json" \
    -d '{
    "title": "The Go Programming Language",
    "author": "Alan Donovan",
    "year": 2015,
    "price": 45.50
}'
