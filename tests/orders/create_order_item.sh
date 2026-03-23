#!/bin/bash

curl -X POST http://localhost:8081/v1/api/orders/item \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJVc2VySWQiOiIzIiwiaXNzIjoiYm9va3N0b3JlLWF1dGgiLCJzdWIiOiIzIiwiZXhwIjoxNzczMzg1MzU1LCJpYXQiOjE3NzMzODE3NTV9.II-7PACjIcbS1RllH36er8t2NEPjin2XvOvMxBM3Xa0" \
    -d '{"book_id": 1}'
