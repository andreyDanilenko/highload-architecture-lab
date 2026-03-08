CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    version INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_transactions (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    request_id VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert test data
INSERT INTO products (sku, name, stock_quantity) 
VALUES ('SKU-TEST-001', 'Highload Test Item', 1000)
ON CONFLICT (sku) DO NOTHING;
