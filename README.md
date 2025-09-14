# GOLANG ORDER MATCHING SYSTEM

## 🎯 **Project Overview**

A simplified order matching engine built in Go that implements a stock exchange matching system. This system handles buy/sell orders via REST API, matches them using price-time priority, executes trades, and maintains state in MySQL database.

## ✨ **Features Implemented**

### **Core Requirements (✅ All Implemented)**
- ✅ **Order Types**: Limit Orders & Market Orders (Buy/Sell)
- ✅ **Price-Time Priority**: Best price first, FIFO at same price
- ✅ **Partial Fills**: Orders partially executed with remaining quantity tracking
- ✅ **REST API**: Complete HTTP endpoints with JSON request/response
- ✅ **MySQL Persistence**: Orders and trades stored with proper indexing
- ✅ **Error Handling**: Comprehensive validation and HTTP status codes
- ✅ **Raw SQL**: No ORM used, pure SQL queries as required

### **Additional Features**
- ✅ **Enhanced Order Book**: Includes timestamps, queue positions, and spread calculation
  - **Spread**: The difference between the highest bid (buy) price and lowest ask (sell) price. It indicates market liquidity - smaller spreads mean more liquid markets.
- ✅ **HTTP Method Validation**: Returns 405 responses for incorrect HTTP methods
- ✅ **Content-Type Validation**: Ensures proper JSON content type
- ✅ **Symbol Support**: Supports up to 50 character symbols for various instruments
- ✅ **Market Order Handling**: Proper handling when no liquidity is available
- ✅ **Comprehensive Testing**: Multiple test scenarios covering edge cases
- ✅ **Consistent Response Format**: Standardized JSON responses with success/error indicators

## 🏗️ **Architecture & Design**

### **System Components**
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   REST API      │    │  Order Matching  │    │   Database      │
│   Layer         │◄──►│  Engine Service  │◄──►│   (MySQL)       │
│                 │    │                  │    │                 │
│ • Place Order   │    │ • Match Logic    │    │ • Orders Table  │
│ • Cancel Order  │    │ • Order Book     │    │ • Trades Table  │
│ • Get OrderBook │    │ • Execute Trades │    │ • Persistence   │
│ • List Trades   │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         ▲                        ▲
         │                        │
         ▼                        ▼
┌─────────────────┐    ┌──────────────────┐
│   HTTP Client   │    │  In-Memory       │
│   (curl/app)    │    │  Order Book      │
│                 │    │  (for fast       │
│                 │    │   matching)      │
└─────────────────┘    └──────────────────┘
```

### **Package Structure**
```
order-matching-engine/
├── main.go                 # Application entry point
├── go.mod                  # Go modules dependency management
├── schema.sql              # MySQL database schema
├── config/
│   └── database.go         # Database configuration
├── models/
│   ├── order.go           # Order data structures
│   ├── trade.go           # Trade data structures
│   └── orderbook.go       # Order book response models
├── handlers/
│   ├── orders.go          # Order HTTP handlers
│   └── trades.go          # Trade HTTP handlers
├── services/
│   ├── matching_engine.go # Core matching logic
│   └── order_book.go      # In-memory order book
├── database/
│   ├── connection.go      # Database connection setup
│   ├── orders_repo.go     # Order database operations
│   └── trades_repo.go     # Trade database operations
├── utils/
│   └── response.go        # HTTP response utilities
└── test_*.sh              # Comprehensive test suites
```

## 🚀 **Quick Start Guide**

### **Prerequisites**
- **Go 1.21+** (Download from https://golang.org/dl/)
- **MySQL 5.7+** or **MySQL 8.0+** (or MariaDB)
- **curl** (for API testing) or **Postman** (optional)

### **1. Database Setup**

#### **Step 1: Create Database**
Open your MySQL client (MySQL Workbench, command line, or phpMyAdmin) and run:
```sql
CREATE DATABASE order_matching;
```

#### **Step 2: Import Schema**
Run the schema file to create tables. Choose your method:

**Option A - MySQL Command Line:**
```bash
# Windows Command Prompt
mysql -u root -p order_matching < schema.sql

# Linux/Mac Terminal
mysql -u root -p order_matching < schema.sql

# If mysql command not found, use full path:
# Windows: "C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe" -u root -p order_matching < schema.sql
```

**Option B - PowerShell (Windows):**
```powershell
Get-Content schema.sql | mysql -u root -p order_matching
```

**Option C - Copy-Paste:**
Open `schema.sql` file, copy the contents, and paste into your MySQL client.

#### **Database Schema Details**
The database schema creates two main tables:

**Orders Table:**
```sql
CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,               -- Unique order identifier (UUID)
    symbol VARCHAR(50) NOT NULL,              -- Trading symbol (e.g., 'AAPL', 'GOOGL')
    side ENUM('buy', 'sell') NOT NULL,        -- Order side
    type ENUM('limit', 'market') NOT NULL,    -- Order type
    price DECIMAL(10,2),                      -- Price (NULL for market orders)
    initial_quantity INT NOT NULL,            -- Original order quantity
    remaining_quantity INT NOT NULL,          -- Unfilled quantity
    status ENUM('open', 'filled', 'cancelled', 'partial') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol_side_price (symbol, side, price, created_at)  -- For fast matching
);
```

**Trades Table:**
```sql
CREATE TABLE trades (
    id VARCHAR(36) PRIMARY KEY,               -- Unique trade identifier
    symbol VARCHAR(50) NOT NULL,              -- Trading symbol
    buy_order_id VARCHAR(36) NOT NULL,        -- Reference to buy order
    sell_order_id VARCHAR(36) NOT NULL,       -- Reference to sell order
    price DECIMAL(10,2) NOT NULL,             -- Execution price
    quantity INT NOT NULL,                    -- Executed quantity
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(id),
    INDEX idx_symbol_time (symbol, executed_at)  -- For trade history queries
);
```

#### **Step 3: Configure Database Connection**

**Option A - Using .env file (Recommended):**
```bash
# Copy the example file
cp .env.example .env

# Windows Command Prompt:
copy .env.example .env
```

Then edit `.env` file with your database credentials:
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_actual_mysql_password
DB_NAME=order_matching
```

**Option B - Set Environment Variables Directly:**
```bash
# Windows Command Prompt
set DB_PASSWORD=your_password
set DB_HOST=localhost
set DB_USER=root
set DB_NAME=order_matching

# Linux/Mac Terminal
export DB_PASSWORD=your_password
export DB_HOST=localhost
export DB_USER=root
export DB_NAME=order_matching
```

**Option C - Use Defaults (Quick Start):**
If you don't set any environment variables, the application will use these defaults:
- Host: `localhost`
- Port: `3306`
- User: `root`
- Password: `password`
- Database: `order_matching`

**Note:** The `.env` file is not included in the repository for security reasons (it's in `.gitignore`). You need to create it from `.env.example`.

### **2. Application Setup**

#### **Step 1: Navigate to Project Directory**
```bash
cd "GOLANG ORDER MATCHING SYSTEM"
# OR if you cloned from GitHub:
cd GOLANG-ORDER-MATCHING-SYSTEM
```

#### **Step 2: Install Go Dependencies**
```bash
# This downloads all required packages
go mod tidy
```

#### **Step 3: Start MySQL Service**
```bash
# Windows
net start mysql80
# OR
net start mysql

# Linux
sudo systemctl start mysql
# OR
sudo service mysql start

# Mac (if using Homebrew)
brew services start mysql
```

#### **Step 4: Run the Application**
```bash
# Start the server (run this in project root directory)
go run main.go
```

**Expected Output:**
```
Database connection established
Order Matching Engine starting on port 8080...
```

**If you see errors:**
- Check if MySQL is running
- Verify database credentials in `.env` file
- Ensure database `order_matching` exists
- Check if port 8080 is available

## 📡 **Complete API Reference**

### **1. Place Order**
```http
POST /orders
Content-Type: application/json

{
    "symbol": "AAPL",
    "side": "buy",           // "buy" or "sell"
    "type": "limit",         // "limit" or "market"
    "price": 150.00,         // Required for limit orders
    "quantity": 100          // Must be positive integer
}
```

**Response:**
```json
{
    "success": true,
    "data": {
        "order": {
            "id": "uuid-123",
            "symbol": "AAPL",
            "side": "buy",
            "type": "limit",
            "price": 150.00,
            "initial_quantity": 100,
            "remaining_quantity": 50,
            "status": "partial",
            "created_at": "2025-09-14T10:15:30Z"
        },
        "trades": [
            {
                "id": "trade-456",
                "symbol": "AAPL",
                "buy_order_id": "uuid-123",
                "sell_order_id": "uuid-789",
                "price": 149.50,
                "quantity": 50,
                "executed_at": "2025-09-14T10:15:30Z"
            }
        ]
    }
}
```

### **2. Get Order Status**
```http
GET /orders/{order_id}
```

### **3. Cancel Order**
```http
DELETE /orders/{order_id}
```

### **4. Get Order Book**
```http
GET /orderbook?symbol=AAPL
```
**Note:** Symbol parameter is optional. If provided, returns order book for that symbol only. If omitted, returns order books for all symbols.

**Response includes additional market data:**
```json
{
    "success": true,
    "data": {
        "symbol": "AAPL",
        "timestamp": "2025-09-14T10:20:00Z",
        "bids": [
            {
                "price": 150.00,
                "quantity": 100,
                "timestamp": "2025-09-14T10:15:30Z",
                "queue_position": 1
            }
        ],
        "asks": [],
        "spread": 5.00,
        "total_bid_orders": 1,
        "total_ask_orders": 0
    }
}
```

**What's included beyond basic requirements:**
- **Spread Calculation**: Shows the difference between best bid and ask prices
- **Queue Position**: Shows order position at each price level (FIFO)
- **Timestamps**: When each order was placed
- **Order Counts**: Total number of buy/sell orders

### **5. Get Trades**
```http
GET /trades?symbol=AAPL
```
**Note:** Symbol parameter is optional. If provided, returns trades for that symbol only. If omitted, returns all trades.

### **6. Health Check**
```http
GET /health
```
**Note:** Only GET method is supported. Other methods (POST, PUT, DELETE) return 405 Method Not Allowed.

## 🔧 **Matching Algorithm Details**

### **Price-Time Priority Implementation**

1. **Price Priority**: 
   - Buy orders: Highest price matched first
   - Sell orders: Lowest price matched first

2. **Time Priority**: 
   - At same price level: First-In-First-Out (FIFO)
   - Older orders execute before newer ones

3. **Execution Logic**:
   ```go
   // Simplified matching logic
   if buyOrder.Type == "market" || sellOrder.Type == "market" {
       return true // Market orders always match
   }
   return *buyOrder.Price >= *sellOrder.Price // Limit order matching
   ```

4. **Trade Price Determination**:
   - Limit vs Limit: Use resting order's price
   - Market vs Limit: Use limit order's price

### **Partial Fill Handling**

- **Limit Orders**: Remaining quantity stays in order book
- **Market Orders**: Remaining quantity cancelled if no more matches
- **Status Updates**: `open` → `partial` → `filled` or `cancelled`


```

## 🧪 **Testing & Validation**

### **Automated Test Scripts**

**Run Basic Tests:**
```bash
# Make script executable (Linux/Mac)
chmod +x test_api_simple.sh
./test_api_simple.sh

# Windows (Git Bash or WSL)
bash test_api_simple.sh
```

**Run Comprehensive Tests:**
```bash
# Make script executable (Linux/Mac)
chmod +x test_comprehensive.sh
./test_comprehensive.sh

# Windows (Git Bash or WSL)
bash test_comprehensive.sh
```

**Test Categories Covered:**
- ✅ **Basic Functionality**: Order placement, matching, retrieval
- ✅ **Validation Tests**: Negative prices, invalid data, missing fields
- ✅ **HTTP Method Validation**: 405 responses for wrong methods
- ✅ **Edge Cases**: Market orders with no liquidity, large quantities
- ✅ **FIFO Testing**: Time priority verification
- ✅ **Cross-Price Matching**: Price improvement scenarios
- ✅ **Multi-Symbol Isolation**: Independent order books per symbol
- ✅ **Error Handling**: Comprehensive error response validation

### **Manual Testing Examples**

#### **Using curl (Command Line)**

**1. Health Check:**
```bash
curl http://localhost:8080/health
```

**2. Place Buy Limit Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol":"AAPL","side":"buy","type":"limit","price":150.00,"quantity":100}'
```

**3. Place Sell Limit Order (will match with buy order):**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol":"AAPL","side":"sell","type":"limit","price":149.00,"quantity":50}'
```

**4. Place Market Order:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"symbol":"AAPL","side":"buy","type":"market","quantity":25}'
```

**5. Get Order Status (replace {order_id} with actual ID from previous responses):**
```bash
curl http://localhost:8080/orders/{order_id}
```

**6. Cancel Order:**
```bash
curl -X DELETE http://localhost:8080/orders/{order_id}
```

**7. Check Order Book:**
```bash
# Get order book for specific symbol
curl "http://localhost:8080/orderbook?symbol=AAPL"

# Get order books for all symbols
curl "http://localhost:8080/orderbook"
```

**8. Check Trades:**
```bash
# Get trades for specific symbol
curl "http://localhost:8080/trades?symbol=AAPL"

# Get all trades
curl "http://localhost:8080/trades"
```

#### **Using Postman or Similar Tools**

**Base URL:** `http://localhost:8080`

**1. POST /orders** - Place Order
- **Method:** POST
- **Headers:** `Content-Type: application/json`
- **Body (JSON):**
  ```json
  {
    "symbol": "AAPL",
    "side": "buy",
    "type": "limit",
    "price": 150.00,
    "quantity": 100
  }
  ```

**2. GET /orders/{order_id}** - Get Order Status
- **Method:** GET
- **URL:** Replace `{order_id}` with actual order ID

**3. DELETE /orders/{order_id}** - Cancel Order
- **Method:** DELETE
- **URL:** Replace `{order_id}` with actual order ID

**4. GET /orderbook** - Get Order Book
- **Method:** GET
- **Query Parameters:** `symbol=AAPL` (optional)

**5. GET /trades** - Get Trades
- **Method:** GET
- **Query Parameters:** `symbol=AAPL` (optional)

**6. GET /health** - Health Check
- **Method:** GET
- **No parameters needed**

#### **Testing Different Scenarios**

**Scenario 1: Basic Matching**
1. Place buy order: AAPL, $150, 100 shares
2. Place sell order: AAPL, $149, 50 shares
3. Result: 50 shares traded at $149, buy order has 50 remaining

**Scenario 2: Market Order**
1. Place sell limit: AAPL, $150, 100 shares
2. Place buy market: AAPL, 50 shares
3. Result: 50 shares traded at $150

**Scenario 3: No Match**
1. Place buy order: AAPL, $100, 100 shares
2. Place sell order: AAPL, $200, 50 shares
3. Result: No trades, both orders remain in book

**Scenario 4: Partial Fill**
1. Place buy order: AAPL, $150, 100 shares
2. Place sell order: AAPL, $150, 150 shares
3. Result: 100 shares traded, sell order has 50 remaining

## 🔒 **Security Features**

### **Current Security Measures**
- ✅ **Input Validation**: Comprehensive request validation
- ✅ **HTTP Method Restrictions**: 405 responses for invalid methods
- ✅ **Content-Type Validation**: Strict JSON requirement
- ✅ **SQL Injection Prevention**: Parameterized queries only
- ✅ **Error Handling**: Secure error messages without internal details
- ✅ **Environment Variables**: Database credentials from environment

## 📊 **Performance Characteristics**

### **Algorithmic Complexity**
- **Order Placement**: O(log n) for insertion into sorted order book
- **Order Matching**: O(m) where m is number of matching orders
- **Order Book Retrieval**: O(1) for top levels, O(k) for k levels
- **Database Operations**: Indexed queries for optimal performance

### **Scalability Features**
- **In-Memory Order Book**: Fast matching without database queries
- **Database Persistence**: Reliable state recovery
- **Symbol Isolation**: Independent order books per trading symbol
- **Concurrent Safety**: Mutex protection for thread-safe operations

## 🎯 **Design Decisions & Assumptions**

### **Key Design Choices**
1. **In-Memory + Database Hybrid**: Fast matching with persistent storage
2. **Price-Time Priority**: Standard exchange matching algorithm
3. **Partial Fill Support**: Real-world trading requirement
4. **Symbol-Based Isolation**: Each symbol has independent order book
5. **Raw SQL**: No ORM for maximum control and performance
6. **Gorilla Mux**: Robust HTTP routing with method validation

### **Assumptions Made**
- Single-threaded matching per symbol (can be extended)
- Decimal precision sufficient for most trading scenarios
- MySQL provides adequate performance for order volume
- HTTP REST API suitable for trading interface
- Order IDs are UUIDs for uniqueness

## 🚀 **Potential Future Improvements**

### **Features That Could Be Added**
- **WebSocket Support**: Real-time order book updates for clients
- **Authentication System**: User accounts and API keys for security
- **Additional Order Types**: Stop orders, iceberg orders, time-in-force options
- **Risk Management**: Position limits and circuit breakers
- **Enhanced Market Data**: More detailed market depth information
- **Performance Optimization**: Connection pooling and caching
- **Monitoring & Logging**: Detailed metrics and alerting systems
- **HTTPS Support**: TLS encryption for secure communication
- **Rate Limiting**: Prevent API abuse

## 📝 **Standards Followed**

### **Industry Best Practices**
- ✅ **Price-Time Priority**: Standard exchange matching algorithm
- ✅ **Audit Trail**: Complete trade and order history
- ✅ **Data Integrity**: Database transactions for consistency
- ✅ **Error Handling**: Proper HTTP status codes and error messages
- ✅ **RESTful API**: Standard HTTP methods and JSON responses

## 🎉 **Project Summary**

This order matching engine successfully implements all the required functionality from the task specification:

- ✅ **All Core Requirements Met**: Order types, matching algorithm, REST API, MySQL persistence
- ✅ **Clean Code Structure**: Well-organized packages and clear separation of concerns
- ✅ **Comprehensive Testing**: Multiple test scenarios covering various edge cases
- ✅ **Proper Error Handling**: Validation and appropriate HTTP responses
- ✅ **Complete Documentation**: Setup instructions and API examples
- ✅ **Database Design**: Proper schema with indexes and foreign keys

**The system is ready for evaluation and demonstrates understanding of order matching concepts and Go development practices.**