-- create enum
CREATE TYPE transaction_status AS ENUM (
    'pending',
    'authorized',
    'captured',
    'failed',
    'voided',
    'refunded'
);

CREATE TYPE payment_type AS ENUM (
    'direct_charge',
    'authorize_capture'
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,
    gateway_ref VARCHAR(255),
    payment_type payment_type NOT NULL,
    card_token VARCHAR(255) NOT NULL,
    card_last_four VARCHAR(4) NOT NULL,
    card_brand VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    status transaction_status DEFAULT 'pending',
    description VARCHAR(500),
    idempotency_key VARCHAR(255)  UNIQUE,
    failed_reason VARCHAR(500),
    captured_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_transactions_merchant
    FOREIGN KEY (merchant_id)
    REFERENCES merchants(id)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    merchant_id UUID NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_idempotency_keys_merchant
    FOREIGN KEY (merchant_id)
    REFERENCES merchants(id)
    ON DELETE CASCADE
);

-- indexes
CREATE INDEX idx_transactions_merchant_id ON transactions(merchant_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_gateway_ref ON transactions(gateway_ref);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

CREATE INDEX idx_idempotency_merchant_id ON idempotency_keys(merchant_id);
CREATE INDEX idx_idempotency_expires_at ON idempotency_keys(expires_at);