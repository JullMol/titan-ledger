# Titan Ledger

A **High-Performance Distributed Ledger System** built with Go, following **Hexagonal Architecture** (Ports & Adapters) patterns used by major tech companies (Gojek, Grab, Tokopedia).

## Features

- ✅ **Double-Entry Bookkeeping** - Every transfer creates balanced DEBIT/CREDIT entries
- ✅ **Atomic Transactions** - All database operations wrapped in transactions with auto-rollback
- ✅ **Race Condition Protection** - Row-level locking with `SELECT FOR UPDATE`
- ✅ **Idempotency** - Duplicate transaction prevention via unique `reference_id`
- ✅ **Graceful Shutdown** - Proper handling of SIGTERM signals
- ✅ **Database Constraints** - CHECK constraint prevents negative balances

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         HTTP / gRPC Layer                           │
│                    (internal/adapters/handler)                      │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          PORTS (Interfaces)                         │
│                     (internal/core/ports)                           │
│   ┌────────────────┐  ┌─────────────────────┐  ┌─────────────────┐ │
│   │ WalletService  │  │ TransactionService  │  │ WalletRepository│ │
│   └────────────────┘  └─────────────────────┘  └─────────────────┘ │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       CORE DOMAIN & SERVICES                        │
│                (internal/core/domain, services)                     │
│                    Pure business logic only                         │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       ADAPTERS (Implementations)                    │
│                   (internal/adapters/repository)                    │
│                     PostgreSQL, Redis, etc.                         │
└─────────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
titan-ledger/
├── cmd/api/                    # Application entrypoint
├── config/                     # Configuration loading
├── internal/
│   ├── core/
│   │   ├── domain/             # Business entities (Wallet, Transaction)
│   │   ├── ports/              # Interfaces (Repository, Service)
│   │   └── services/           # Business logic implementation
│   └── adapters/
│       ├── handler/http/       # Fiber HTTP handlers
│       └── repository/postgres/# PostgreSQL implementations
├── migrations/                 # Database migrations
└── docker-compose.yml          # PostgreSQL for development
```

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make (optional)

### 1. Start Database
```bash
make run-db
# or
docker-compose up -d
```

### 2. Run Migrations
```bash
docker exec -it titan_postgres psql -U titan_user -d titan_ledger -f /tmp/migrations/000001_init_schema.up.sql
```

### 3. Run Application
```bash
go run cmd/api/main.go
```

### 4. Test API

**Create Wallet:**
```bash
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123"}'
```

**Get Balance:**
```bash
curl http://localhost:8080/api/v1/wallets/{wallet_id}
```

**Transfer:**
```bash
curl -X POST http://localhost:8080/api/v1/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "from_wallet_id": "wallet-uuid-1",
    "to_wallet_id": "wallet-uuid-2",
    "amount": 50000,
    "reference_id": "txn-001"
  }'
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/wallets` | Create new wallet |
| GET | `/api/v1/wallets/:id` | Get wallet balance |
| POST | `/api/v1/transfer` | Transfer between wallets |

## Configuration

Environment variables (`.env` file):

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | 8080 |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5435 |
| `DB_USER` | Database user | titan_user |
| `DB_PASSWORD` | Database password | titan_password |
| `DB_NAME` | Database name | titan_ledger |

## Technology Stack

- **Framework**: [Fiber](https://gofiber.io/) - Express-inspired web framework
- **Database**: PostgreSQL 15 with [pgx](https://github.com/jackc/pgx) driver
- **Config**: [Viper](https://github.com/spf13/viper) - Configuration management
- **Architecture**: Hexagonal Architecture (Ports & Adapters)

## Interactive API Playground

Visit the homepage (`/`) to access an interactive API playground where you can:
- Create wallets
- Check balances
- Make transfers

No external tools needed - everything works in your browser!

## Deployment

### Deploy to Render.com (Free)

1. Push code to GitHub
2. Go to [render.com](https://render.com) and connect your repository
3. Click "New Blueprint" and select your repo
4. Render will auto-detect `render.yaml` and create:
   - PostgreSQL database (free tier)
   - Go web service

5. After deployment, run migrations:
   - Go to your database settings in Render
   - Connect via shell and run the SQL files from `/migrations`

Your app will be live at `https://your-app.onrender.com`

## Design Decisions

### Why Hexagonal Architecture?
Separates business logic from infrastructure. Database can be swapped from PostgreSQL to MongoDB without changing core business logic.

### Why Row-Level Locking (FOR UPDATE)?
Prevents race conditions when multiple concurrent transfers access the same wallet.

### Why Idempotency via reference_id?
If network fails after transaction is processed, client retrying with same `reference_id` will get duplicate error instead of double-charging.

## License

MIT
