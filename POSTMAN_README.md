# Vyking Go Integration Service — Postman Test Suite Guide (`v-service-test.json`)

This guide explains how to import, configure, and run the attached Postman collection (`v-service-test.json`) for testing the **Go-based Helpwise Integration & Partner Transaction API**. It also documents the HTTP methods, headers, query parameters, and test scenarios covered by the suite.

---

## 1. Prerequisites

Before running any requests in the collection, ensure your environment is prepared:

1. **Start the Docker Stack**:
   Make sure the Go application, MySQL 8 database, and Redis containers are running:
   ```bash
   docker-compose up --build -d
   ```
2. **Seed the Test Data**:
   Load your MySQL database with the test data batch so that customer lookups (such as `jane@example.com` under Brand 1) succeed.

---

## 2. How to Import & Run `v-service-test.json`

1. Open **Postman**.
2. Click **Import** in the upper-left corner of the workspace.
3. Drag and drop `v-service-test.json` into the import window (or click **Files** and select it from your computer).
4. Once imported, **"Vyking Integration Service Test Suite"** will appear in your **Collections** sidebar.
5. *(Optional)* To run the entire test suite at once, hover over the collection name, click the three dots (`...`), and select **Run collection**.

### Collection Variables
The collection defines a built-in variable so you do not have to hardcode server addresses:

| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `{{baseUrl}}` | `http://localhost:8080` | The root host and port where the Go partner HTTP API is listening. Update this in the collection's **Variables** tab if running on a different port or remote host. |

---

## 3. API Endpoint, HTTP Method & Parameters Explained

All requests in this collection target the read-only Partner Transaction endpoint:
`{{baseUrl}}/api/v1/transactions`

### HTTP Method
* **`GET`**: Used across every request in the collection because this endpoint is strictly read-only. It retrieves customer transaction history without modifying state in MySQL or Helpwise.

### Request Headers
| Header Key | Format / Value | Required | Purpose & Behavior |
| :--- | :--- | :--- | :--- |
| `Authorization` | `Bearer <partner-token>` | **Yes** | Authenticates the partner integration (e.g., Helpwise) and maps the request to a specific internal brand ID (`brand1-partner-token` maps to Brand `1`). Missing headers return `401 Unauthorized`; invalid tokens return `403 Forbidden`. |

### Query Parameters
| Parameter | Type | Required | Example | Purpose & Behavior |
| :--- | :--- | :--- | :--- | :--- |
| `email` | `string` | **Yes** | `jane@example.com` | Identifies the customer whose transactions are being queried. Scoped strictly to the brand associated with the provided Bearer token. Omitting this parameter triggers a `400 Bad Request` validation error. |
| `limit` | `integer` | No | `10` | Controls the maximum number of recent transactions returned in the response payload, ordered chronologically from newest to oldest. |
| `status` | `string` | No | `completed` | Filters the returned transactions by normalized Helpwise status categories (`pending`, `completed`, `failed`, `cancelled`), translating internal MySQL status IDs into readable partner statuses. |

---

## 4. Test Suite Breakdown

The `v-service-test.json` collection is organized into three logical folders covering security, input validation, and core business logic:

### Folder 1: `1. Authentication & Security Tests`
Verifies that the authentication middleware properly isolates brands and blocks unauthorized access.

1. **Missing Authorization Header**
   * **Setup**: `GET {{baseUrl}}/api/v1/transactions?email=jane@example.com` with **no** `Authorization` header.
   * **What it tests**: Ensures unauthenticated callers cannot access customer financial records.
   * **Expected Result**: `401 Unauthorized`.

2. **Invalid Partner Token**
   * **Setup**: `GET {{baseUrl}}/api/v1/transactions?email=jane@example.com` with `Authorization: Bearer invalid-token-xyz`.
   * **What it tests**: Ensures unrecognized or revoked tokens are rejected before querying the database.
   * **Expected Result**: `403 Forbidden`.

---

### Folder 2: `2. Request Parameter Validation`
Verifies that incoming requests enforce required query parameters before executing database lookups.

3. **Missing Email Parameter**
   * **Setup**: `GET {{baseUrl}}/api/v1/transactions` with a valid `Authorization: Bearer brand1-partner-token` header, but **no** `email` query parameter.
   * **What it tests**: Confirms the handler validates required query strings and returns a structured client error.
   * **Expected Result**: `400 Bad Request`.

---

### Folder 3: `3. Business Logic & Transaction Filtering`
Verifies end-to-end database querying, brand scoping, status translation, and currency minor-unit formatting.

4. **Get Recent Transactions (Brand 1 - Jane Doe)**
   * **Setup**: `GET {{baseUrl}}/api/v1/transactions?email=jane@example.com&limit=10` with `Authorization: Bearer brand1-partner-token`.
   * **What it tests**: Happy-path retrieval. Verifies that `jane@example.com` is resolved under Brand `1`, up to `10` transactions are returned, minor currency units are properly converted, and a unique `request_id` is included in the JSON response.
   * **Expected Result**: `200 OK` with a JSON payload of transactions.

5. **Filter Transactions by Status ('completed')**
   * **Setup**: `GET {{baseUrl}}/api/v1/transactions?email=jane@example.com&status=completed` with `Authorization: Bearer brand1-partner-token`.
   * **What it tests**: Verifies that passing `status=completed` filters out pending/failed/cancelled records and only returns transactions mapped to the `completed` status.
   * **Expected Result**: `200 OK` containing only completed transactions.

6. **Customer Not Found Case**
   * **Setup**: `GET {{baseUrl}}/api/v1/transactions?email=nonexistent@example.com` with `Authorization: Bearer brand1-partner-token`.
   * **What it tests**: Verifies graceful handling when a valid partner queries an email address that does not exist (or belongs to a different brand).
   * **Expected Result**: `404 Not Found` (`CUSTOMER_NOT_FOUND`).