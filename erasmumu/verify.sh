#!/bin/bash
set -e

BASE_URL="http://localhost:8081"
echo "Starting Erasmumu Verification..."

# 1. Create Offer
echo "Creating Offer..."
RESPONSE=$(curl -s -X POST $BASE_URL/offer -d '{
    "title": "Software Engineer Intern",
    "link": "http://example.com",
    "city": "Berlin",
    "domain": "IT",
    "salary": 1200,
    "startDate": "2023-09-01",
    "endDate": "2024-02-28",
    "available": true
}')
echo "Response: $RESPONSE"

ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Created Offer ID: $ID"

if [ -z "$ID" ]; then
    echo "Failed to create offer"
    exit 1
fi

# 2. Get Offer
echo "Getting offer..."
curl -v $BASE_URL/offer/$ID

# 3. Get Offers by City
echo "Getting offers in Berlin..."
curl -v "$BASE_URL/offer?city=Berlin"

# 4. Update Offer
echo "Updating offer..."
curl -v -X PUT $BASE_URL/offer/$ID -d '{
    "title": "Senior Software Engineer Intern",
    "link": "http://example.com",
    "city": "Berlin",
    "domain": "IT",
    "salary": 1500,
    "startDate": "2023-09-01",
    "endDate": "2024-02-28",
    "available": true
}'

# 5. Delete Offer
echo "Deleting offer..."
curl -v -X DELETE $BASE_URL/offer/$ID

echo "Done."
