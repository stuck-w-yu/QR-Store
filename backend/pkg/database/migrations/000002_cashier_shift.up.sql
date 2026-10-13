-- 000002_cashier_shift.up.sql
-- Cashier Shift Management Schema

CREATE TABLE IF NOT EXISTS registers (
    id VARCHAR(64) PRIMARY KEY,
    restaurant_id VARCHAR(64) REFERENCES restaurants(id) ON DELETE CASCADE NOT NULL,
    outlet_id VARCHAR(64),
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS cashier_shifts (
    id VARCHAR(64) PRIMARY KEY,
    restaurant_id VARCHAR(64) REFERENCES restaurants(id) ON DELETE CASCADE NOT NULL,
    outlet_id VARCHAR(64),
    register_id VARCHAR(64) REFERENCES registers(id) ON DELETE RESTRICT NOT NULL,
    cashier_id VARCHAR(64) REFERENCES users(id) ON DELETE RESTRICT NOT NULL,
    status VARCHAR(20) NOT NULL, -- OPEN, CLOSING, CLOSED, CANCELLED
    opening_balance BIGINT DEFAULT 0 NOT NULL,
    opened_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    closed_at TIMESTAMPTZ,
    expected_total BIGINT DEFAULT 0 NOT NULL,
    actual_total BIGINT,
    difference BIGINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Ensure only one active shift per register
CREATE UNIQUE INDEX IF NOT EXISTS uk_active_register_shift ON cashier_shifts (register_id) WHERE status = 'OPEN';

-- Ensure only one active shift per cashier
CREATE UNIQUE INDEX IF NOT EXISTS uk_active_cashier_shift ON cashier_shifts (cashier_id) WHERE status = 'OPEN';

CREATE TABLE IF NOT EXISTS cashier_shift_transactions (
    id VARCHAR(64) PRIMARY KEY,
    shift_id VARCHAR(64) REFERENCES cashier_shifts(id) ON DELETE CASCADE NOT NULL,
    order_id VARCHAR(64) REFERENCES orders(id) ON DELETE SET NULL,
    payment_id VARCHAR(64) REFERENCES payments(id) ON DELETE SET NULL,
    type VARCHAR(30) NOT NULL, -- SALE, REFUND, VOID, ADJUSTMENT
    amount BIGINT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS shift_closings (
    id VARCHAR(64) PRIMARY KEY,
    shift_id VARCHAR(64) UNIQUE REFERENCES cashier_shifts(id) ON DELETE CASCADE NOT NULL,
    gross_sales BIGINT DEFAULT 0 NOT NULL,
    refund_total BIGINT DEFAULT 0 NOT NULL,
    void_total BIGINT DEFAULT 0 NOT NULL,
    adjustment_total BIGINT DEFAULT 0 NOT NULL,
    net_sales BIGINT DEFAULT 0 NOT NULL,
    expected_amount BIGINT DEFAULT 0 NOT NULL,
    actual_amount BIGINT,
    difference BIGINT,
    notes TEXT,
    closed_by VARCHAR(64) REFERENCES users(id) NOT NULL,
    closed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS refunds (
    id VARCHAR(64) PRIMARY KEY,
    restaurant_id VARCHAR(64) REFERENCES restaurants(id) ON DELETE CASCADE NOT NULL,
    order_id VARCHAR(64) REFERENCES orders(id) ON DELETE CASCADE NOT NULL,
    payment_id VARCHAR(64) REFERENCES payments(id) ON DELETE CASCADE NOT NULL,
    amount BIGINT NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(30) DEFAULT 'COMPLETED' NOT NULL, -- PENDING, COMPLETED, REJECTED
    requested_by VARCHAR(64) REFERENCES users(id) NOT NULL,
    approved_by VARCHAR(64) REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at TIMESTAMPTZ
);

-- Recommended Indexes
CREATE INDEX IF NOT EXISTS idx_registers_restaurant ON registers(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_restaurant ON cashier_shifts(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_cashier ON cashier_shifts(cashier_id);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_status ON cashier_shifts(status);
CREATE INDEX IF NOT EXISTS idx_shift_transactions_shift ON cashier_shift_transactions(shift_id);
CREATE INDEX IF NOT EXISTS idx_shift_transactions_created ON cashier_shift_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_refunds_order_id ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_restaurant_id ON refunds(restaurant_id);
