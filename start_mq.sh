#!/bin/bash

docker run  \
    -d \
    --name rabbitmq \
    -e RABBITMQ_DEFAULT_USER=user \
    -e RABBITMQ_DEFAULT_PASS=password \
    -p 8080:15672 -p 5672:5672 \
    rabbitmq:3-management

