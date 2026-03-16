FROM golang:1.26-alpine

WORKDIR /app

# Copy go mod dan sum dulu agar caching layer lebih cepat
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build aplikasi
RUN go build -o main cmd/api/main.go

# Expose port
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
