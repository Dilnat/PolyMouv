#!/bin/bash
set -e
# Verify Polytech Service

BASE_URL="http://localhost:8080"
echo "Starting Verification..."

# 1. Create Student
echo "Creating Student..."
RESPONSE=$(curl -s -X POST $BASE_URL/student -d '{"firstname":"John", "name":"Doe", "domain":"IT"}')
echo "Response: $RESPONSE"

# Extract ID using grep/sed (simple approximation since jq might not be available)
ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "Created Student ID: $ID"

if [ -z "$ID" ]; then
    echo "Failed to create student"
    exit 1
fi

echo "Listing Students..."
curl -v $BASE_URL/student

echo "Getting specific student..."
curl -v $BASE_URL/student/$ID

echo "Updating Student..."
curl -v -X PUT $BASE_URL/student/$ID -d '{"firstname":"Johnny", "name":"Doe", "domain":"IT"}'

echo "Deleting Student..."
curl -v -X DELETE $BASE_URL/student/$ID

echo "Done."
