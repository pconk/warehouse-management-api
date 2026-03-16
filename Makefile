# Variabel
APP_NAME=warehouse-api
MAIN_PATH=cmd/api/main.go

.PHONY: all build run test cover clean help

# Default action: run
all: run

# Menjalankan aplikasi
run:
	go run $(MAIN_PATH)

# Kompilasi aplikasi menjadi binary
build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

# Menjalankan semua unit test dengan ringkas
test:
	go test ./internal/handler/... -cover

# Menjalankan test dan melihat coverage di terminal
cover:
	go test ./internal/handler/... -coverprofile=cover.out
	go tool cover -func=cover.out

# Membuka laporan coverage di browser (visual)
cover-html:
	go test ./internal/handler/... -coverprofile=cover.out
	go tool cover -html=cover.out

# Download semua dependencies
tidy:
	go mod tidy

# Menghapus file binary dan coverage
clean:
	rm -f bin/$(APP_NAME)
	rm -f cover.out

# Docker commands
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f

swagger:
    swag init -g cmd/api/main.go -d ./ --parseDependency --parseInternal


# Menampilkan bantuan
help:
	@echo "Perintah yang tersedia:"
	@echo "  make run        - Menjalankan aplikasi"
	@echo "  make test       - Menjalankan semua unit test"
	@echo "  make cover      - Melihat persentase coverage test"
	@echo "  make cover-html - Melihat detail baris yang tercover di browser"
	@echo "  make build      - Membuat file binary untuk production"
	@echo "  make tidy       - Merapikan go.mod"
	@echo "  make docker-up       - Membangun image dan menjalankan MySQL + API di background (-d). Database otomatis terisi 10 data item dari folder scripts/"
	@echo "  docker-logs       - Melihat log dari aplikasi yang sedang jalan di dalam Docker (buat debugging)"
	@echo "  make docker-down       - Mematikan server dan menghapus volume database (bersih-bersih)"

