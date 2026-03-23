#!/bin/bash

curl -X POST http://localhost:8081/v1/api/orders \
    -d '{"user_id": 3}'
