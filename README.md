# Pharmacy POS API

A RESTful API backend for a Pharmacy Point-of-Sale system built with **Go**, **Gin**, and **MongoDB**.

## Tech Stack

- **Go** 1.25
- **Gin** — HTTP web framework
- **MongoDB** — Primary database
- **Redis** — Rate limiting (optional)
- **JWT** — Authentication (`golang-jwt/v5`)
- **Logrus** — Structured logging

## Project Structure

```
api/
├── main.go                  # Entry point
├── app/
│   ├── routes.go            # Route definitions & server bootstrap
│   ├── core/
│   │   ├── config/          # App configuration
│   │   ├── constant/        # Shared constants
│   │   └── errs/            # Custom error types
│   ├── domain/
│   │   ├── model/           # Data models (Product, Patient, Sale, Inventory, Setting)
│   │   ├── repository/      # Repository interfaces
│   │   └── usecase/         # Business logic layer
│   └── features/
│       ├── api/             # HTTP handlers
│       ├── repo/            # Repository implementations
│       └── request/         # Request DTOs
├── db/
│   └── init.go              # MongoDB & Redis connection
├── middlewares/
│   ├── authentication.go    # JWT authentication
│   ├── authorization.go     # Role-based access control
│   ├── cors.go              # CORS configuration
│   ├── rate_limit.go        # Redis-based rate limiting
│   ├── recovery.go          # Panic recovery
│   └── no_route.go          # 404 handler
└── routes/                  # (reserved)
```

## Getting Started

### Prerequisites

- Go 1.25+
- MongoDB instance
- Redis instance (optional, for rate limiting)

### Environment Variables

Create a `.env` file in the project root:

```env
PORT=8000
GIN_MODE=debug              # debug | release | test
MONGO_HOST=<mongodb-uri>
MONGO_PHARM_DB_NAME=pharm_pos
REDIS_HOST=<redis-uri>      # optional
SECRET_KEY=<jwt-secret>
CLIENT_ID=<client-id>
SYSTEM=PHARM
```

### Run

```bash
go mod download
go run main.go
```

The server starts on `http://localhost:8000` by default.

### Health Check

```
GET /api/v1/ping
```

## API Endpoints

All protected routes require a valid JWT token in the `Authorization` header.

### Products (`/api/v1/products`) — Authenticated

| Method | Path                  | Description          |
|--------|-----------------------|----------------------|
| POST   | `/`                   | Create product       |
| GET    | `/`                   | List all products    |
| GET    | `/:id`                | Get product by ID    |
| GET    | `/barcode/:barcode`   | Get product by barcode |
| PUT    | `/:id`                | Update product       |
| DELETE | `/:id`                | Delete product       |

### Patients (`/api/v1/patients`) — Authenticated

| Method | Path       | Description        |
|--------|------------|--------------------|
| POST   | `/`        | Create patient     |
| GET    | `/`        | List all patients  |
| GET    | `/:id`     | Get patient by ID  |
| PUT    | `/:id`     | Update patient     |
| DELETE | `/:id`     | Delete patient     |

### Sales (`/api/v1/sales`) — Authenticated

| Method | Path                  | Description              |
|--------|-----------------------|--------------------------|
| POST   | `/`                   | Create sale              |
| GET    | `/`                   | List all sales           |
| GET    | `/:id`                | Get sale by ID           |
| POST   | `/check-interactions` | Check drug interactions  |
| POST   | `/check-allergies`    | Check patient allergies  |

### Batches / Inventory (`/api/v1/batches`) — Authenticated

| Method | Path                    | Description                 |
|--------|-------------------------|-----------------------------|
| POST   | `/`                     | Create batch                |
| GET    | `/product/:productId`   | Get batches by product      |
| GET    | `/expiring`             | Get expiring batches        |
| GET    | `/low-stock`            | Get low-stock items         |
| PUT    | `/:id`                  | Update batch                |
| DELETE | `/:id`                  | Delete batch                |

### Dashboard (`/api/v1/dashboard`) — Authenticated

| Method | Path                | Description            |
|--------|---------------------|------------------------|
| GET    | `/stats`            | Overview statistics    |
| GET    | `/sales-summary`    | Sales summary          |
| GET    | `/monthly-summary`  | Monthly summary        |
| GET    | `/gross-margin`     | Gross margin report    |
| GET    | `/abc-analysis`     | ABC inventory analysis |
| GET    | `/dead-stock`       | Dead stock report      |
| GET    | `/refill-reminders` | Refill reminders       |
| GET    | `/expiring`         | Expiring batches       |
| GET    | `/low-stock`        | Low-stock alerts       |

### Reports (`/api/v1/reports`) — SUPER / ADMIN only

| Method | Path    | Description                          |
|--------|---------|--------------------------------------|
| GET    | `/ky9`  | Drug purchase ledger (บัญชีซื้อยา)   |
| GET    | `/ky10` | Dangerous drug sales (ขายยาอันตราย)  |
| GET    | `/ky11` | Specially controlled drugs (ยาควบคุมพิเศษ) |
| GET    | `/ky12` | Psychotropic substances (วัตถุออกฤทธิ์ฯ) |
| GET    | `/ky13` | Narcotics category 3 (ยาเสพติดให้โทษ ประเภท 3) |

### Patient History — Authenticated

| Method | Path                          | Description         |
|--------|-------------------------------|---------------------|
| GET    | `/patients/:id/history`       | Patient sale history|

### Settings (`/api/v1/settings`) — SUPER / ADMIN only

| Method | Path     | Description         |
|--------|----------|---------------------|
| GET    | `/`      | List all settings   |
| GET    | `/:key`  | Get setting by key  |
| PUT    | `/:key`  | Upsert setting      |

## Middlewares

- **Authentication** — JWT token validation
- **Authorization** — Role-based access (`SUPER`, `ADMIN`)
- **CORS** — Cross-origin resource sharing
- **Rate Limiting** — Redis-backed request throttling
- **Recovery** — Graceful panic recovery
- **NoRoute** — Custom 404 response
