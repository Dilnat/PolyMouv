#!/bin/bash
set -e
# Verify Polytech Service

BASE_URL="http://localhost:8080"
echo "Starting Verification..."

# 1. Create Student
echo "Creating Student..."
curl -v -X POST $BASE_URL/student -d '{"firstname":"John", "name":"Doe", "domain":"IT"}' | grep "John"

# Id extraction is tricky without jq, assuming first created is ID 1 for now or rely on consistent output.
# For manual verification script, simple checks are fine.

echo "Listing Students..."
curl -v $BASE_URL/student

echo "Getting specific student..."
curl -v $BASE_URL/student/1

echo "Updating Student..."
curl -v -X PUT $BASE_URL/student/1 -d '{"firstname":"Johnny", "name":"Doe", "domain":"IT"}'

echo "Deleting Student..."
curl -v -X DELETE $BASE_URL/student/1

echo "Done."
