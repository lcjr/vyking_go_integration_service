# Senior Backend - Third-Party Integration Task (Helpwise Integration)

This repository contains a production-ready Go service for integrating an internal MySQL 8 database with **Helpwise**, a third-party support platform used across multiple company brands.

---

## 1. Architecture & Design Diagram

```
                 +-------------------------------------------------------+
                 |               HELPWISE SUPPORT PLATFORM               |
                 +-------------------+-------------------+---------------+
                                     |                   ^
      Bearer Token Auth (Partner API)|                   | Bearer <brand-api-key>
       GET /api/v1/transactions      v                   | (Outbound Sync)
+------------------------------------+-------------------+-----------------------+
| GO INTEGRATION SERVICE (Multi-Replica Safe)                                   |
|                                                                                |
|  +----------------------------+             +-------------------------------+  |
|  |     Inbound HTTP Handler   |             |     Outbound Sync Worker      |  |
|  |  - Auth Middleware (Tokens)|             |  - Cron Scheduler / Ticker    |  |
|  |  - Query Validator         |             |  - Distributed Lock / Redis   |  |
|  |  - Status Mapper           |             |  - Rate-Limited Client        |  |
|  +--------------+-------------+             +---------------+---------------+  |
|                 |                                           |                  |
|                 +-------------------+-----------------------+                  |
|                                     |                                          |
|                         +-----------v-----------+                              |
|                         |  Domain / Service Layer|                              |
|                         +-----------+-----------+                              |
|                                     |                                          |
+-------------------------------------|------------------------------------------+
                                      v
                 +--------------------+-------------------+
                 |        MYSQL 8 DATABASE (Read/Write)   |
                 +----------------------------------------+
                 | Tables Read:                           |
                 |  - customers, transactions, brands     |
                 |  - currencies, transaction_statuses    |
                 | Tables Written:                        |
                 |  - sync_queue (SYQ_dProcessed)         |
                 |  - partner_contacts (PCO_cContactId)   |
                 +----------------------------------------+
```

---

## 2. Questions, Assumptions, & Architectural Decisions (NOTES)

### A. Questions to Ask Before Production Deployment
1. **Partner API Authentication**: How are partner tokens generated, rotated, and scoped to brands? Is there an authorization/IAM service, or a dedicated table mapping tokens to brand IDs?
2. **Inactive Customers**: How should customers with `CUS_bActive = 0` be handled? Should they be set as `"active": false` in Helpwise attributes or skipped entirely?
3. **Permanent Failure Handling**: How should the `sync_queue` handle permanent 4xx validation errors (e.g., `422 Unprocessable Entity`) to avoid blocking future items?
4. **Currency Minor Units Edge Case**: For currencies where `CUR_iMinorUnits` is `NULL` (e.g., USD in reference data), what is the fallback rule? (Defaulted to 2 decimal places).

### B. Key Assumptions
1. **Customer Matching**: Customer `CUS_iID` is used as `external_id` in Helpwise calls. On success, `id` returned by Helpwise is stored in `partner_contacts`.
2. **Status Mapping**: MySQL transaction status IDs map to Helpwise's vocabulary (`pending`, `completed`, `failed`, `cancelled`).
3. **Multi-Replica Execution**: High-volume queue processing relies on MySQL `FOR UPDATE SKIP LOCKED` transactions to prevent row locking contention across replicas.
4. **Rate Limiting**: Helpwise's strict 60 req/min (burst 10) per brand API key limit requires a shared sliding window / token bucket implemented in Redis.

### C. Technical Decisions & Trade-Offs
* **MySQL `FOR UPDATE SKIP LOCKED` over Message Queues**: Avoided introducing Kafka/RabbitMQ to keep infrastructure lean while ensuring multi-replica safety natively in MySQL 8.
* **Redis Rate Limiting**: Added Redis to enforce global rate limits across all scaling service replicas.
* **Clean Architecture**: Domain, Repository, Service, and Handler layers are cleanly decoupled to allow high unit test coverage and easy mocking.

---

## 3. Project Structure

```
.
├── Dockerfile
├── docker-compose.yml
├── init.sql
├── go.mod
├── README.md
├── cmd/
│   └── service/
│       └── main.go
└── internal/
    ├── domain/
    │   └── models.go
    ├── repository/
    │   └── mysql.go
    ├── service/
    │   ├── sync.go
    │   ├── transaction.go
    │   └── transaction_test.go
    ├── handler/
    │   └── http.go
    └── cache/
        └── redis.go
```

---

## 4. How to Build & Run

### Prerequisites
* Docker & Docker Compose
* Go 1.22+ (for local testing)

### Running with Docker Compose
To launch MySQL 8, Redis, and the Integration Service concurrently:
```bash
docker-compose up --build -d
```

### Running Unit Tests
Execute the unit test suite:
```bash
go test -v ./internal/service/...
```
### Running Postman Tests 
See POSTMAN_README.md

### AI Disclaimer
Gemini was used to create both .md files and Postman collection