#!/bin/bash

BOOK_ID=1

curl -X GET http://localhost:8081/v1/api/book/$BOOK_ID
