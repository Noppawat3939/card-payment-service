## Credit Card Payment Service

A RESTful API service for processing credit card payments through a Third-Party Payment Gateway, written in Go.

---

#### 1. Overview

The Payment Gateway API enables developers and businesses to securely integrate credit card payment processing into their applications. This service acts as a **Payment Adapter** between the merchant's system and a Third-Party Payment Gateway.

```
Merchant registers and receives API Key
        │
        ▼
Client (Web / Mobile / Backend)
        │
        ▼
Credit Card Payment Service   ◄──── Webhook / Callback
        │
        ▼
Third-Party Payment Gateway
```

---

#### 2. Features

##### 2.1 Merchant Registration & API Key Management

- Register a merchant account with business information
- Issue API Key + Secret upon successful registration
- Support Key Rotation — replace keys when compromised
- Register a Webhook URL to receive payment callbacks
- Merchant status lifecycle: `pending` → `active` → `suspended`
- Only merchants with `active` status can access the Payment API

##### 2.2 Payment Transaction

- Create a charge request with amount, currency, and order reference
- Support payment via tokenized card
- Log every request and response for auditing

##### 2.3 Card Tokenization

- Convert raw card data into a secure token before forwarding to the gateway
- Never store raw card number or CVV on the server
- Reduces PCI DSS scope for the merchant

##### 2.4 Authorize & Capture

- **Authorize** — Place a hold on funds without charging the card immediately
- **Capture** — Charge the card after a successful authorization

##### 2.5 Refund

- Full refund — return the full charged amount to the cardholder
- Query refund status at any time

##### 2.6 Void / Cancel

- Cancel a transaction that has been authorized but not yet captured
- Release the authorization hold on the cardholder's account

##### 2.7 Webhook / Callback

- Deliver real-time payment events to the merchant's registered endpoint
- Supported events: `payment.success`, `payment.failed`, `refund.completed`
- Verify HMAC Signature on every incoming webhook before processing

##### 2.8 Transaction Status

The system supports the following status lifecycle:

```
pending → authorized → captured → refunded
                    ↘
                    voided
                    ↘
                    failed
```

##### 2.9 Security

- API Key + Secret to authenticate every request
- HMAC Signature verification for webhooks
- Idempotency Key — prevents duplicate charges
- Rate limiting and request validation
- Audit log for every transaction

---

#### 3. Tech Stack

| Component | Technology              |
| --------- | ----------------------- |
| Language  | Go                      |
| Framework | Gin                     |
| Database  | PostgreSQL              |
| Cache     | Redis                   |
| Logging   | zerolog                 |
| Testing   | testify                 |
| Container | Docker / Docker Compose |

#### 4. Project Structure

```
credit-card-payment-service/
├── cmd/
│   └── server/
│       └── main.go                  # App bootstrap + infra init
│
├── internal/
│   ├── config/
│   │   └── config.go                # Env config + DSN helper
│   │
│   ├── infra/
│   │   └── database/
│   │       └── database.go          # GORM PostgreSQL connection
│   │   └── redis/
│   │       └── redis.go             # Redis connection
│   │
│   ├── logger/
│   │   ├── logger.go                # Zerolog initialization
│   │   └── middleware.go            # Gin request logging middleware
│   │
│   ├── domain/
│   │   ├── merchant.go              # Merchant entity + status
│   │   ├── api_key.go               # API key entity
│   │   ├── payment.go               # Payment entity + lifecycle
│   │   └── errors.go                # Domain business errors
│   │
│   ├── handler/
│   │   ├── dto/
│   │   │   ├── merchant_dto.go      # HTTP request / response DTO
│   │   │   └── payment_dto.go
│   │   │
│   │   ├── merchant_handler.go      # Merchant endpoints
│   │   ├── payment_handler.go       # Payment endpoints
│   │   ├── webhook_handler.go       # Webhook callback endpoints
│   │   └── playground_handler.go    # Dev-only testing UI
│   │
│   ├── service/
│   │   ├── merchant_service.go      # Merchant use cases
│   │   ├── payment_service.go       # Payment business flow
│   │   └── token_service.go         # Tokenization flow
│   │
│   ├── repository/
│   │   ├── merchant_repo.go         # Merchant DB access
│   │   ├── api_key_repo.go          # API key DB access
│   │   ├── payment_repo.go          # Payment DB access
│   │   └── token_repo.go            # Token DB access
│   │
│   ├── gateway/
│   │   └── gateway_client.go        # Third-party payment adapter
│   │
│   ├── middleware/
│   │   ├── auth.go                  # API key auth
│   │   ├── idempotency.go           # Duplicate request protection
│   │   └── rate_limit.go            # Rate limiting
│   │
│   ├── response/
│   │   ├── response.go              # Success response formatter
│   │   └── error.go                 # Error response mapper
│   │
│   └── router/
│       └── router.go                # Route registration + dependency wiring
│
├── static/
│   └── playground/
│       ├── index.html
│       ├── style.css
│       └── app.js
│
├── migrations/
│   ├── 000001_create_merchants.up.sql
│   ├── 000001_create_merchants.down.sql
│   ├── 000002_create_transactions.up.sql
│   ├── 000002_create_transactions.down.sql
│   ├── 000003_create_tokens.up.sql
│   └── 000003_create_tokens.down.sql
│
├── .env.example
├── .env.local
├── docker-compose.yml
├── .air.toml
├── Makefile
└── README.md
```

---

#### 5. Sequence Diagram Flow

##### Void / Cancel Flow

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant DB as Database
    participant GW as Payment Gateway

    C->>S: POST /api/v1/payments/{transaction_id}/void
    S->>S: Validate API Key
    S->>DB: GET transaction
    DB-->>S: transaction {status, gateway_ref}

    alt Transaction status = captured
        S-->>C: 422 Unprocessable<br/>{error: cannot void a captured transaction, use refund instead}
    else Transaction status = authorized
        S->>GW: POST void {gateway_ref}
        GW-->>S: {status=voided}
        S->>DB: UPDATE transaction (status=voided)
        DB-->>S: ok
        S-->>C: 200 OK {transaction_id, status=voided}
    end
```

##### Refund Flow

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant DB as Database
    participant GW as Payment Gateway

    C->>S: POST /api/v1/payments/{transaction_id}/refund<br/>{reason}
    S->>S: Validate API Key
    S->>DB: GET transaction (assert status=captured)
    DB-->>S: transaction {gateway_ref, amount}
    S->>GW: POST refund {gateway_ref, amount}
    GW-->>S: {refund_ref, status=processing}
    S->>DB: UPDATE transaction (status=refunded)<br/>INSERT refund record
    DB-->>S: ok
    S-->>C: 200 OK {refund_id, status=processing}

    Note over S,GW: Gateway processes refund asynchronously

    GW->>S: Webhook POST /api/v1/webhooks/payment<br/>{event=refund.completed, refund_ref}
    S->>S: Verify HMAC Signature
    S->>DB: UPDATE refund (status=completed)
    DB-->>S: ok
    S-->>GW: 200 OK
```

##### Webhook / Callback Flow

```mermaid
sequenceDiagram
    participant GW as Payment Gateway
    participant S as Payment Service
    participant DB as Database
    actor M as Merchant

    GW->>S: POST /api/v1/webhooks/payment<br/>Header: X-Signature: hmac_sha256<br/>{event, transaction_id, gateway_ref, status}
    S->>S: Verify HMAC Signature

    alt Signature invalid
        S-->>GW: 401 Unauthorized
    else Signature valid
        S->>DB: GET transaction by gateway_ref
        DB-->>S: transaction
        S->>DB: UPDATE transaction status
        DB-->>S: ok
        S-->>GW: 200 OK

        Note over S,M: Forward event to merchant webhook URL

        S->>M: POST {merchant webhook_url}<br/>{event, transaction_id, status, amount}
        M-->>S: 200 OK
    end

    alt Merchant webhook fails (timeout / 5xx)
        S->>S: Schedule retry (exponential backoff)<br/>max 3 attempts
        S->>M: POST {merchant webhook_url} (retry)
        M-->>S: 200 OK
    end
```

The system is designed to integrate with any Web, Mobile, or Backend service and supports the full transaction lifecycle — from creating a payment intent through to refund and void. **Only registered merchants with an active API Key are permitted to access the Payment API.**

---

> **Note:** The `/dev/playground` route is only accessible when `APP_ENV=development`. It is automatically disabled in production.

---

#### 5. API Reference

##### Base URL

```text
http://localhost:8080/v1
```

##### 5.1 Merchant Register Flow

```mermaid
sequenceDiagram
    actor M as Merchant
    participant S as Payment Service
    participant DB as Database

    M->>S: POST /merchants/register
    S->>S: Validate request body
    S->>DB: INSERT merchant (status=pending)
    DB-->>S: merchant_id

    S->>S: Generate API Key + Secret
    S->>DB: INSERT api_key (hashed secret)
    DB-->>S: ok

    S-->>M: 201 Created
    Note over M,S: api_secret shown only once

    M->>S: PATCH /merchants/activate
    S->>DB: UPDATE status=active
    DB-->>S: ok
    S-->>M: 200 OK
```

##### 5.1.1 Register Merchant

Create a new merchant account and issue API credentials.

```http
POST /merchants/register
Content-Type: application/json
```

###### Request

```json
{
  "name": "Acme Corp",
  "email": "ops@acme.com",
  "webhook_url": "https://acme.com/webhook"
}
```

###### Success Response

**201 Created**

```json
{
  "success": true,
  "data": {
    "merchant_id": "8f7f6d4e-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "api_key": "pk_live_xxxxxxxx",
    "api_secret": "sk_live_xxxxxxxxxxxxx",
    "status": "pending"
  }
}
```

> `api_secret` is returned **only once** and must be stored securely by the merchant.

###### Error Responses

**400 Bad Request**

```json
{
  "success": false,
  "error": "invalid request body"
}
```

**406 Not Acceptable**

```json
{
  "success": false,
  "error": "merchant already exists"
}
```

---

##### 5.1.2 Activate Merchant

Activate a merchant account that is currently in `pending` status.

```http
PATCH /merchants/activate
Content-Type: application/json
```

###### Request

```json
{
  "email": "ops@acme.com"
}
```

###### Success Response

**200 OK**

```json
{
  "success": true,
  "data": {
    "name": "Acme Corp",
    "email": "ops@acme.com",
    "status": "active"
  }
}
```

###### Error Responses

**404 Not Found**

```json
{
  "success": false,
  "error": "merchant email not found"
}
```

**406 Not Acceptable**

```json
{
  "success": false,
  "error": "merchant current status not accepted"
}
```

##### 5.2 Payment Charge Flow

The payment service supports two transaction modes to cover real-world payment scenarios:
• Authorize + Capture (2-step payment)
Used when the merchant wants to hold the cardholder’s funds first (`pending → authorized`) and capture (`authorized → captured`) the payment later after confirming the order, inventory, or service.
• Direct Charge (1-step payment)
Used when the merchant wants to charge (`pending → captured`) the card immediately without placing a hold.

The following sections describe the Authorize → Capture flow, which is commonly used in booking systems, hotel reservations, and order confirmation processes.

##### 5.2.1 Authorize (Hold)

This flow places a temporary hold on the customer’s available balance without immediately transferring funds.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant GW as Payment Gateway
    participant DB as Database

    C->>S: POST /v1/payments/authorize
    S->>S: Validate API Key + request body
    S->>GW: Tokenize card data
    GW-->>S: {card_token, brand, last_four}

    S->>DB: INSERT transaction (status=pending)
    DB-->>S: transaction_id

    S->>GW: POST authorize {card_token, amount, currency}
    GW-->>S: {gateway_ref, status=authorized}

    S->>DB: UPDATE transaction<br/>status=authorized,<br/>gateway_ref
    DB-->>S: ok

    S-->>C: 200 OK<br/>{transaction_id, status=authorized}
```

##### 5.2.2 Capture

This flow completes the actual payment after a successful authorization by transferring the held amount.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant GW as Payment Gateway
    participant DB as Database

    C->>S: POST /v1/payments/{transaction_id}/capture
    S->>S: Validate API Key
    S->>DB: GET transaction by id
    DB-->>S: {status=authorized, gateway_ref}

    alt status != authorized
        S-->>C: 422 Unprocessable Entity
    else status = authorized
        S->>GW: POST capture {gateway_ref}
        GW-->>S: {status=captured}

        S->>DB: UPDATE transaction<br/>status=captured,<br/>captured_at=now()
        DB-->>S: ok

        S-->>C: 200 OK<br/>{transaction_id, status=captured}
    end
```

##### 5.2.3 Capture Directly

This flow charges the customer immediately without placing a temporary hold.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant GW as Payment Gateway
    participant DB as Database

    C->>S: POST /v1/payments/charge
    S->>S: Validate API Key + request body
    S->>GW: Tokenize card data
    GW-->>S: {card_token, brand, last_four}

    S->>DB: INSERT transaction (status=pending)
    DB-->>S: transaction_id

    S->>GW: POST charge {card_token, amount, currency}
    GW-->>S: {gateway_ref, status=success}

    S->>DB: UPDATE transaction<br/>status=captured,<br/>gateway_ref,<br/>captured_at=now()
    DB-->>S: ok

    S-->>C: 200 OK<br/>{transaction_id, status=captured}
```
