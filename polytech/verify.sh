#!/bin/bash
set -e
# Verify Polytech Service

BASE_URL="http://localhost:8080"
echo "Starting Verification..."

# 1. Create Student
echo "Creating Student (IT)..."
RESPONSE=$(curl -s -X POST $BASE_URL/student -d '{"firstname":"John", "name":"Doe", "domain":"IT"}')
echo "Response: $RESPONSE"

# Extract ID using grep/sed (simple approximation since jq might not be available)
STUDENT_ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "Created Student ID: $STUDENT_ID"

if [ -z "$STUDENT_ID" ]; then
    echo "Failed to create student"
    exit 1
fi

echo "Listing Students..."
curl -v $BASE_URL/student

# --- Integration Tests ---

# 2. Create Offer (IT) - Valid
echo "Creating Offer (IT) in Erasmumu..."
OFFER_RESP_IT=$(curl -s -X POST http://localhost:8081/offer -d '{
    "title": "IT Internship",
    "link": "http://example.com",
    "city": "Paris",
    "domain": "IT",
    "salary": 1000,
    "available": true
}')
OFFER_ID_IT=$(echo $OFFER_RESP_IT | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Created IT Offer ID: $OFFER_ID_IT"

# 3. Register (Valid)
echo "Registering Student to IT Offer..."
REG_RESP_VALID=$(curl -s -X POST $BASE_URL/internship -d "{\"studentId\":$STUDENT_ID, \"offerId\":\"$OFFER_ID_IT\"}")
echo "Registration Response: $REG_RESP_VALID"

# 4. Create Offer (Biology) - Invalid Domain
echo "Creating Offer (Biology) in Erasmumu..."
OFFER_RESP_BIO=$(curl -s -X POST http://localhost:8081/offer -d '{
    "title": "Bio Internship",
    "link": "http://example.com",
    "city": "Paris",
    "domain": "Biology",
    "salary": 1000,
    "available": true
}')
OFFER_ID_BIO=$(echo $OFFER_RESP_BIO | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Created Bio Offer ID: $OFFER_ID_BIO"

# 5. Register (Invalid)
echo "Registering Student to Bio Offer..."
REG_RESP_INVALID=$(curl -s -X POST $BASE_URL/internship -d "{\"studentId\":$STUDENT_ID, \"offerId\":\"$OFFER_ID_BIO\"}")
echo "Registration Response (Should be Rejected): $REG_RESP_INVALID"


echo "Getting specific student..."
curl -v $BASE_URL/student/$STUDENT_ID

echo "Updating Student..."
curl -v -X PUT $BASE_URL/student/$STUDENT_ID -d '{"firstname":"Johnny", "name":"Doe", "domain":"IT"}'

echo "Deleting Student..."
curl -v -X DELETE $BASE_URL/student/$STUDENT_ID

echo "Done."
