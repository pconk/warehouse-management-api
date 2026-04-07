#!/bin/bash

# Konfigurasi variabel
HOST="localhost:50053"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzU4NzIxNTYsImlzcyI6ImF1ZGl0LXNlcnZpY2UtdGVzdCIsInJvbGUiOiJhZG1pbiIsInVzZXJfaWQiOiIxIiwidXNlcm5hbWUiOiJhZG1pbl9ndWRhbmciLCJ3YXJlaG91c2VfaWQiOiJXSC1KS1QtMDk5In0.ISffAU9X6ijm6LyJFPR5A6NL-RQ9v_CbrPDskieWCNo"
SERVICE_METHOD="warehouse.WarehouseService/IncreaseStock"

# Tentukan jumlah thread/proses paralel
THREADS=3

echo "====================================================="
echo " Parallel gRPC Testing Tool ($THREADS threads)"
echo "====================================================="

# Fungsi untuk mengirim request
send_request() {
  local id=$1
  echo "[Thread-$id] Starting request..."
  
  # Menjalankan grpcurl
  # Output dikelompokkan agar tidak terlalu berantakan di terminal
  RESPONSE=$(grpcurl -plaintext \
    -H "authorization: Bearer $TOKEN" \
    -d '{
        "items": [
        {"item_id": 1, "quantity": 2},
        {"item_id": 7, "quantity": 4},
        {"item_id": 9, "quantity": 1}
        ]
    }' \
    "$HOST" "$SERVICE_METHOD" 2>&1)

  echo "[Thread-$id] Response: $RESPONSE"
}

# Jalankan loop untuk memicu background process
for i in $(seq 1 $THREADS); do
  send_request "$i" &
done

# Tunggu semua background process selesai
wait

echo "====================================================="
echo " All threads finished."
echo "====================================================="