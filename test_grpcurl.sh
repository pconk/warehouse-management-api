#!/bin/bash

# Konfigurasi variabel
HOST="localhost:50053"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzU4NzIxNTYsImlzcyI6ImF1ZGl0LXNlcnZpY2UtdGVzdCIsInJvbGUiOiJhZG1pbiIsInVzZXJfaWQiOiIxIiwidXNlcm5hbWUiOiJhZG1pbl9ndWRhbmciLCJ3YXJlaG91c2VfaWQiOiJXSC1KS1QtMDk5In0.ISffAU9X6ijm6LyJFPR5A6NL-RQ9v_CbrPDskieWCNo"
SERVICE_METHOD="warehouse.WarehouseService/IncreaseStock"

echo "====================================================="
echo " gRPC Testing Tool for Warehouse System"
echo "====================================================="

# 1. Test Skenario Sukses
echo -e "\n[1] Testing Success Scenario..."
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{
    "items": [
      {"item_id": 1, "quantity": 2},
      {"item_id": 7, "quantity": 4},
      {"item_id": 9, "quantity": 1}
    ]
  }' \
  "$HOST" "$SERVICE_METHOD"

# 2. Test Skenario Validasi Salah (Quantity Negatif / ID 0)
echo -e "\n[2] Testing Validation Error (Quantity -5)..."
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{
    "items": [
      {"item_id": 1, "quantity": -5}
    ]
  }' \
  "$HOST" "$SERVICE_METHOD"

# 3. Test Skenario Tanpa Token (Unauthenticated)
echo -e "\n[3] Testing Unauthenticated Request..."
grpcurl -plaintext \
  -d '{"items": [{"item_id": 1, "quantity": 5}]}' \
  "$HOST" "$SERVICE_METHOD"