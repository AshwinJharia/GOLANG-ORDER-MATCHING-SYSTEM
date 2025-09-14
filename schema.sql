-- Orders table
CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    side ENUM('buy', 'sell') NOT NULL,
    type ENUM('limit', 'market') NOT NULL,
    price DECIMAL(10,2), -- NULL for market orders
    initial_quantity INT NOT NULL,
    remaining_quantity INT NOT NULL,
    status ENUM('open', 'filled', 'cancelled', 'partial') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol_side_price (symbol, side, price, created_at)
);

-- Trades table
CREATE TABLE trades (
    id VARCHAR(36) PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    buy_order_id VARCHAR(36) NOT NULL,
    sell_order_id VARCHAR(36) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(id),
    INDEX idx_symbol_time (symbol, executed_at)
);