#!/bin/bash

curl -X GET "http://localhost:8081/v1/api/book/search?title=go&year=2015&orderby=title&order=asc&page=1&size=10"
