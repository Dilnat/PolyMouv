#!/bin/bash
set -e

# Download grpcurl if not present (assuming linux x86_64)
if ! command -v ./grpcurl &> /dev/null; then
    echo "Downloading grpcurl..."
    curl -L -O https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_x86_64.tar.gz
    tar -xvf grpcurl_1.8.9_linux_x86_64.tar.gz
    rm grpcurl_1.8.9_linux_x86_64.tar.gz
fi

echo "Listing Services..."
./grpcurl -plaintext localhost:50051 list

echo "Calling GetLatestNews (Limit 2)..."
./grpcurl -plaintext -d '{"limit": 2}' localhost:50051 mi8.MI8Service/GetLatestNews

echo "Calling GetLatestNewsInCity (Berlin)..."
./grpcurl -plaintext -d '{"city": "Berlin", "limit": 5}' localhost:50051 mi8.MI8Service/GetLatestNewsInCity
