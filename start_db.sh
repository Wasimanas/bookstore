#!/bin/bash


docker run -d \
  --name bookstore \
  -e POSTGRES_USER=bookstore \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=bookstore \
  -p 5432:5432 \
  -v bstore_data:/var/lib/postgresql/data \
  postgres:16


