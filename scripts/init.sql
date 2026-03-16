CREATE USER IF NOT EXISTS 'root'@'%' IDENTIFIED BY 'rootpassword';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;

-- 1. Create Database
CREATE DATABASE IF NOT EXISTS warehouse_go;
USE warehouse_go;

-- 2. Table Categories
CREATE TABLE categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Table Items
CREATE TABLE items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category_id INT,
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(15, 2) NOT NULL,
    stock INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- 4. Table Users
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    api_key VARCHAR(100) UNIQUE NOT NULL,
    role ENUM('admin', 'staff') DEFAULT 'staff',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Table Stock Logs
CREATE TABLE stock_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    item_id INT,
    user_id INT,
    type ENUM('IN', 'OUT') NOT NULL,
    quantity INT NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_item FOREIGN KEY (item_id) REFERENCES items(id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 6. Seed Initial Data (Data Awal)
INSERT INTO categories (name, description) VALUES 
('Elektronik', 'Gadget dan peralatan elektronik'),
('Alat Kantor', 'Perlengkapan tulis dan mebel kantor');

INSERT INTO users (username, api_key, role) VALUES 
('admin_gudang', 'secret-admin-key', 'admin'),
('staff_gudang', 'secret-staff-key', 'staff');

INSERT INTO items (category_id, sku, name, price, stock) VALUES 
(1, 'MAC-001', 'Macbook Pro M2 14-inch', 28000000.00, 10),
(1, 'DL-002', 'Dell UltraSharp U2723QE', 8500000.00, 5),
(2, 'CH-001', 'Ergonomic Office Chair', 2500000.00, 15),
(1, 'IP-014', 'iPhone 14 Pro 256GB', 18000000.00, 8),
(1, 'KB-001', 'Keychron K2 Mechanical Keyboard', 1200000.00, 20),
(1, 'MS-001', 'Logitech MX Master 3S', 1500000.00, 12),
(2, 'DSK-002', 'Adjustable Standing Desk', 4500000.00, 4),
(1, 'TAB-001', 'iPad Air Gen 5 M1', 9500000.00, 7),
(1, 'HD-001', 'Sony WH-1000XM5', 4800000.00, 6),
(2, 'PRN-001', 'Epson EcoTank L3210', 2300000.00, 9);

INSERT INTO stock_logs (item_id, user_id, type, quantity, reason) VALUES 
(1, 1, 'IN', 10, 'Initial Stock Restock'),
(2, 1, 'IN', 5, 'Initial Stock Restock');
