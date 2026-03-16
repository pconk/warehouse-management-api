[![Go CI](https://github.com/pconk/warehouse-management-api/actions/workflows/ci.yml/badge.svg)](https://github.com/pconk/warehouse-management-api/actions/workflows/ci.yml)

------------------------------
Warehouse Management API 📦
API sederhana namun production-ready untuk mengelola stok gudang, kategori barang, dan riwayat mutasi stok. Dibangun menggunakan Go (Golang) dengan fokus pada performa, keamanan, dan integritas data.
🚀 Fitur Utama

* CRUD Items & Categories: Manajemen data master barang dan kategori.
* Stock Transaction Control: Update stok menggunakan Database Transaction (TX) untuk menjamin konsistensi data.
* Audit Trail (Stock Logs): Mencatat setiap perubahan stok lengkap dengan siapa pelakunya (User ID), alasan, dan jumlahnya.
* Secure Authentication: Proteksi endpoint menggunakan API Key Strategy.
* Role-Based Access Control (RBAC): Membedakan hak akses antara admin dan staff menggunakan Middleware.
* Soft Delete: Menghapus data tanpa benar-benar menghilangkannya dari database (menjaga integritas riwayat).
* Pagination & Filtering: Navigasi data yang efisien dengan parameter page, limit, dan pencarian dinamis.
* Graceful Shutdown: Memastikan server berhenti dengan aman tanpa memutus transaksi yang sedang berjalan.

🛠️ Tech Stack

* Language: Go (Golang)
* Router: go-chi/chi (Lightweight & Standard Library compatible)
* Database: MySQL / MariaDB
* Validation: [go-playground/validator](https://github.com/go-playground/validator)
* Logging: slog (Structured Logging)

📁 Struktur Project
Mengikuti standar Clean Architecture sederhana:
```text
.
├── cmd/api/main.go          # Entry point aplikasi
├── internal/
│   ├── config/              # Load .env & konfigurasi
│   ├── entity/              # Struct model database
│   ├── handler/             # HTTP Handlers (Controller)
│   ├── middleware/          # Auth, RBAC, Logger, Recovery
│   ├── repository/          # Data Access Layer (Implementation of Interfaces)
│   └── helper/              # Response wrapper, Pagination helper
└── pkg/
    └── database/            # Database connection driver
```

## 🧪 Testing Strategy

Project ini mengimplementasikan **Unit Testing** dengan fokus pada isolation testing di layer Handler (Controller).

### Key Features:
- **Mocking Pattern**: Menggunakan `testify/mock` untuk mengisolasi logic handler dari dependensi repository/database.

- **Automated CI**: Setiap *push* ke GitHub akan memicu GitHub Actions untuk menjalankan seluruh test suite secara otomatis.

### How to Run Tests:
- go test -v -cover ./internal/handler/...

### 🔑 Testing Credentials
Gunakan Header `X-API-KEY` pada Swagger/Postman:
- **Admin**: `secret-admin-key` (Full Access)
- **Staff**: `secret-staff-key` (Read Only/Limited)

🚥 Endpoint API (Ringkasan)

### Public

| Method | Endpoint            | Description       |
| ------ | ------------------- | ----------------- |
| GET    | /health             | Health check      |
| GET    | /swagger/index.html | API documentation |

---

### Authenticated Routes

| Method | Endpoint          | Description      |
| ------ | ----------------- | ---------------- |
| GET    | /categories       | Daftar kategori  |
| GET    | /items            | List barang      |
| GET    | /items/{id}       | Barang dengan ID |
| POST   | /items/create     | Tambah barang    |
| GET    | /items/export     | Export ke CSV    |
| GET    | /items/stock-logs | Riwayat stok     |


| Method | Endpoint            | Description       |
| ------ | ------------------- | ----------------- |
| PUT    | /items/{id}         | Update barang     |
| DELETE | /items/{id}         | Soft delete barang|
| POST   | /items/update-stock | Update stok       |

---

⚙️ Cara Menjalankan

   1. Clone repository ini.
   2. Buat database warehouse_go dan eksekusi file .sql yang disediakan.
   3. Buat file .env (contoh ada di .env.example).
   4. Jalankan perintah:
   
   go mod tidy
   go run cmd/api/main.go

# 📊 Database Design

Database utama:

| Table      | Description            |
| ---------- | ---------------------- |
| items      | Data barang            |
| categories | Kategori barang        |
| stock_logs | Riwayat perubahan stok |

Audit trail memastikan setiap perubahan stok tercatat.

## 🐳 Quick Start with Docker
Jika Anda memiliki Docker terinstal, cukup jalankan:
```bash
make docker-up



