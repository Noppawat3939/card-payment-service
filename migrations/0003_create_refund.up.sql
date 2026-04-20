-- create enum
CREATE TYPE refund_status AS ENUM (
    'processing',
    'completed',
    'failed'
);

CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    refund_ref VARCHAR(255),
    amount BIGINT NOT NULL,
    status refund_status DEFAULT 'processing',
    failed_reason VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_refuns_merchant
    FOREIGN KEY (merchant_id)
    REFERENCES merchants(id)
    ON DELETE CASCADE,

    CONSTRAINT fk_refuns_transactions
    FOREIGN KEY (transaction_id)
    REFERENCES transactions(id)
    ON DELETE CASCADE
);

-- indexes
CREATE INDEX idx_refunds_refund_ref ON refunds(refund_ref);
CREATE INDEX idx_refunds_transaction_id ON refunds(transaction_id);