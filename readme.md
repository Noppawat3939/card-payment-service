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
│       └── main.go                      # App bootstrap + infra init
│
├── internal/
│   ├── config/
│   │   └── config.go                    # Env config + DSN helper + Redis address
│   │
│   ├── infra/
│   │   ├── database/
│   │   │   └── database.go              # PostgreSQL connection
│   │   └── redis/
│   │       ├── redis.go                 # Redis connection
│   │       └── locker.go                # Distributed lock (SET NX EX)
│   │
│   ├── logger/
│   │   ├── logger.go                    # Zerolog initialization
│   │   └── middleware.go                # Gin request logging middleware
│   │
│   ├── domain/
│   │   ├── merchant.go                  # Merchant entity + status
│   │   ├── api_key.go                   # API key entity
│   │   ├── transaction.go               # Transaction entity + lifecycle
│   │   ├── refund.go                    # Refund entity + status
│   │   └── errors.go                    # Domain business errors
│   │
│   ├── handler/
│   │   ├── dto/
│   │   │   ├── merchant_dto.go          # Merchant request / response DTO
│   │   │   └── payment_dto.go           # Payment request / response DTO
│   │   ├── merchant_handler.go          # Merchant endpoints
│   │   ├── payment_handler.go           # Payment endpoints
│   │   ├── webhook_handler.go           # Webhook receiver + HMAC verify
│   │   └── playground_handler.go        # Dev-only testing UI (embed.FS)
│   │
│   ├── service/
│   │   ├── merchant_service.go          # Merchant registration, key management
│   │   └── payment_service.go           # Payment business flow
│   │
│   ├── repository/
│   │   ├── merchant_repo.go             # Merchant DB access
│   │   ├── api_key_repo.go              # API key DB access
│   │   ├── transaction_repo.go          # Transaction DB access
│   │   ├── refund_repo.go               # Refund DB access
│   │   └── idempotency_repo.go          # Idempotency key DB access
│   │
│   ├── gateway/
│   │   ├── gateway.go                   # Gateway interface + request/response types
│   │   ├── mock_gateway.go              # Mock gateway (dev + test)
│   │   └── retryable_gateway.go         # Retry wrapper (exponential backoff + jitter)
│   │
│   ├── middleware/
│   │   ├── auth.go                      # API key validation + merchant status check
│   │   ├── idempotency.go               # Duplicate request protection
│   │   └── rate_limit.go                # Rate limiter
│   │
│   ├── response/
│   │   ├── response.go                  # Success response formatter
│   │   └── error.go                     # Error response mapper
│   │
│   └── router/
│       └── router.go                    # Route registration + dependency wiring
│
├── static/                              # Embedded via embed.FS (dev only)
│   └── playground/
│       ├── index.html                   # Payment testing UI
│       ├── style.css
│       └── app.js
│
├── tools/
│   └── webhook-simulator/               # Dev tool — simulate gateway webhook callbacks
│       ├── main.go                      # HTTP server (port 8081)
│       ├── event.go                     # Events from webhook
│       ├── simulator.go                 # Event builder + HMAC signer
│       └── README.md                    # How to run + available events
│
├── migrations/
│   ├── 000001_xxxxxx.up.sql
│   └── 000001_xxxxxx.down.sql
│
├── .env.example
├── .env.local
├── .air.toml
├── docker-compose.yml
├── Makefile
└── README.md
```

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
    participant R as Redis
    participant GW as Gateway Adapter
    participant PG as Payment Gateway
    participant DB as Database

    C->>S: POST /v1/payments/authorize<br/>Header: X-API-Key, Idempotency-Key
    S->>S: Validate API Key + request body
    S->>R: Check Idempotency-Key

    alt duplicate request
        R-->>S: cached response
        S-->>C: 200 OK (same response)

    else new request
        R-->>S: not found

        S->>GW: Tokenize card data
        GW->>GW: Build tokenize payload
        GW->>GW: Generate signed headers
        GW->>PG: POST /tokenize + signed headers
        PG-->>GW: {card_token, brand, last_four}
        GW-->>S: {card_token, brand, last_four}

        S->>DB: INSERT transaction<br/>status=pending
        DB-->>S: transaction_id

        S->>GW: Authorize payment {card_token, amount, currency}
        GW->>GW: Build authorize payload
        GW->>GW: Generate signed headers
        GW->>PG: POST /authorize + signed headers

        alt gateway rejected
            PG-->>GW: {status=failed, reason=card_declined}
            GW-->>S: failed response
            S->>DB: UPDATE transaction<br/>status=failed,<br/>failed_reason
            S-->>C: 402 Payment Required

        else gateway success
            PG-->>GW: {gateway_ref, status=authorized}
            GW-->>S: success response
            S->>DB: UPDATE transaction<br/>status=authorized,<br/>gateway_ref,<br/>authorized_at=now()
            S->>R: Save Idempotency-Key + response
            S-->>C: 200 OK<br/>{transaction_id, status=authorized}
        end
    end
```

##### 5.2.2 Capture

This flow completes the actual payment after a successful authorization by transferring the held amount.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant R as Redis (Lock)
    participant GW as Payment Gateway
    participant DB as Database

    C->>S: POST /v1/payments/{transaction_id}/capture<br/>Header: X-Merchant-ID
    S->>S: Validate API Key

    S->>R: Acquire lock (key=tx:{transaction_id})

    alt lock not acquired
        R-->>S: already locked
        S-->>C: 409 Conflict (duplicate request)
    else lock acquired
        R-->>S: lock success

        S->>DB: GET transaction by id + merchant_id
        DB-->>S: {status=authorized, gateway_ref}

        alt transaction not found or wrong merchant
            S-->>C: 404 Not Found
        else status != authorized
            S-->>C: 422 Unprocessable Entity
        else valid
            S->>GW: POST capture {gateway_ref}

            alt gateway rejected
                GW-->>S: {status=failed, reason=capture_failed}
                S->>DB: UPDATE transaction<br/>status=failed,<br/>failed_reason
                S-->>C: 402 Payment Required
            else gateway success
                GW-->>S: {status=captured}
                S->>DB: UPDATE transaction<br/>status=captured,<br/>captured_at=now()
                S-->>C: 200 OK<br/>{transaction_id, status=captured}
            end
        end

        S->>R: Release lock
    end
```

##### 5.2.3 Capture Directly

This flow charges the customer immediately without placing a temporary hold.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant R as Redis
    participant GW as Payment Gateway
    participant DB as Database

    C->>S: POST /v1/payments/charge<br/>Header: X-Merchant-ID, Idempotency-Key
    S->>S: Validate API Key + request body
    S->>R: Check Idempotency-Key

    alt duplicate request
        R-->>S: cached response
        S-->>C: 200 OK (same response)
    else new request
        R-->>S: not found

        S->>GW: Tokenize card data
        GW-->>S: {card_token, brand, last_four}

        S->>DB: INSERT transaction<br/>status=pending
        DB-->>S: transaction_id

        S->>GW: POST charge {card_token, amount, currency}

        alt gateway rejected
            GW-->>S: {status=failed, reason=insufficient_funds}
            S->>DB: UPDATE transaction<br/>status=failed,<br/>failed_reason
            S-->>C: 402 Payment Required
        else gateway success
            GW-->>S: {gateway_ref, status=success}
            S->>DB: UPDATE transaction<br/>status=captured,<br/>gateway_ref,<br/>captured_at=now()
            S->>R: Save Idempotency-Key + response
            S-->>C: 200 OK<br/>{transaction_id, status=captured}
        end
    end
```

##### 5.3 Void / Cancel Flow

Cancels an `authorized` transaction before capture.

The service acquires a Redis lock to prevent duplicate requests, then validates that the transaction is in authorized state. If valid, it calls the gateway to void the payment and updates the transaction status accordingly (voided or failed). Finally, the lock is released and the result is returned.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant R as Redis (Lock)
    participant DB as Database
    participant GW as Payment Gateway

    C->>S: POST /api/v1/payments/{transaction_id}/void
    S->>S: Validate API Key

    S->>R: Acquire lock (lock:tx:{transaction_id})

    alt lock not acquired
        R-->>S: already locked
        S-->>C: 409 Conflict (duplicate request)
    else lock acquired
        R-->>S: lock success
        Note over S,R: defer Release lock (runs on all paths below)

        S->>DB: GET transaction<br/>WHERE id = ? AND merchant_id = ?
        DB-->>S: result

        alt transaction not found
            S-->>C: 404 Not Found
        else status = voided
            S-->>C: 409 Conflict (already voided)
        else status = captured
            S-->>C: 422 Unprocessable (use refund instead)
        else status != authorized
            S-->>C: 422 Invalid state
        else status = authorized
            S->>GW: POST void {gateway_ref}

            alt gateway rejected
                GW-->>S: {status=failed, reason}
                S->>DB: UPDATE transaction<br/>status=failed, failed_reason<br/>WHERE status=authorized
                S-->>C: 402 Payment Required
            else success
                GW-->>S: {status=voided}
                S->>DB: UPDATE transaction<br/>status=voided<br/>WHERE status=authorized
                DB-->>S: ok
                S-->>C: 200 OK {transaction_id, status=voided}
            end
        end
    end
```

##### 5.4 Refund Flow

Refunds a `captured` transaction asynchronously.

The service validates the request and uses idempotency + Redis lock to prevent duplicate refunds. It then calls the gateway to initiate the refund and stores a refund record with processing status. The final result is updated later via webhook from the gateway.

```mermaid
sequenceDiagram
    actor C as Client
    participant S as Payment Service
    participant R as Redis (Lock + Idempotency)
    participant DB as Database
    participant GW as Payment Gateway

    C->>S: POST /api/v1/payments/{transaction_id}/refund<br/>Header: Idempotency-Key
    S->>S: Validate API Key

    S->>R: Check Idempotency-Key

    alt duplicate request
        R-->>S: cached response
        S-->>C: 200 OK (same response)
    else new request
        R-->>S: not found

        S->>R: Acquire lock (lock:tx:{transaction_id})

        alt lock not acquired
            R-->>S: already locked
            S-->>C: 409 Conflict (duplicate request)
        else lock acquired
            R-->>S: lock success
            Note over S,R: defer Release lock (runs on all paths below)

            S->>DB: GET transaction<br/>WHERE id = ? AND merchant_id = ?
            DB-->>S: result

            alt transaction not found
                S-->>C: 404 Not Found
            else status = refunded
                S-->>C: 409 Conflict (already refunded)
            else status != captured
                S-->>C: 422 Unprocessable (invalid state for refund)
            else status = captured
                S->>GW: POST refund {gateway_ref, amount}

                alt gateway rejected
                    GW-->>S: {status=failed, reason}
                    S-->>C: 402 Payment Required
                else accepted
                    GW-->>S: {refund_ref, status=processing}

                    S->>DB: UPDATE transaction<br/>status=refunded<br/>WHERE status=captured
                    S->>DB: INSERT refund record<br/>{refund_ref, status=processing}
                    DB-->>S: ok

                    S->>R: Save Idempotency-Key + response

                    S-->>C: 200 OK {refund_id, status=processing}
                end
            end
        end
    end

    Note over S,GW: async processing

    GW->>S: Webhook /payment {event=refund.completed}
    S->>S: Verify HMAC Signature
    S->>DB: UPDATE refund (status=completed)
    S-->>GW: 200 OK
```

##### 5.5 Webhook / Callback Flow

Handles asynchronous updates from the payment gateway (e.g. refund completion).

The service verifies the HMAC signature to ensure the request is trusted, then checks idempotency to avoid processing duplicate events. It updates the corresponding record (e.g. refund status) in the database and returns a success response. After that, the event is forwarded to the merchant’s webhook endpoint with retry on failure.

```mermaid
sequenceDiagram
    participant GW as Payment Gateway
    participant S as Payment Service
    participant R as Redis (Idempotency)
    participant DB as Database
    actor M as Merchant

    GW->>S: POST /api/v1/webhooks/payment<br/>Header: X-Signature<br/>{event=refund.completed, refund_ref, gateway_ref, status}

    S->>S: Verify HMAC Signature

    alt Signature invalid
        S-->>GW: 401 Unauthorized
    else Signature valid

        S->>R: Check webhook idempotency (refund_ref)

        alt duplicate webhook
            R-->>S: already processed
            S-->>GW: 200 OK
        else new webhook
            R-->>S: not found

            S->>DB: GET refund by refund_ref
            DB-->>S: result

            alt refund not found
                S->>S: log error (unknown refund_ref)
                S-->>GW: 200 OK (prevent gateway retry loop)
            else already final state (completed / failed)
                S->>R: Save webhook idempotency
                S-->>GW: 200 OK
            else valid transition (processing → completed)
                S->>DB: UPDATE refund status=completed<br/>WHERE status=processing
                DB-->>S: ok

                S->>R: Save webhook idempotency
                S-->>GW: 200 OK

                Note over S,M: async — notify merchant via registered webhook_url

                S->>M: POST {merchant webhook_url}<br/>{event, refund_id, status=completed}

                alt merchant webhook success
                    M-->>S: 200 OK
                else merchant webhook fails (timeout / 5xx)
                    S->>S: Push to retry queue<br/>(exponential backoff, max 3 attempts)
                    S->>M: POST webhook (retry)
                    M-->>S: 200 OK
                end
            end
        end
    end
```
