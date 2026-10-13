-- 000003_phase1_cleanup.up.sql
-- Add payment_status, payment_method, completed_at, cancelled_at to orders
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_status VARCHAR(50) DEFAULT 'UNPAID' NOT NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

-- Add paid_amount, change_amount, reference_number, verified_by, verified_at to payments
ALTER TABLE payments ADD COLUMN IF NOT EXISTS paid_amount BIGINT DEFAULT 0 NOT NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS change_amount BIGINT DEFAULT 0 NOT NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reference_number VARCHAR(100);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS verified_by VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;

-- Create service_requests table
CREATE TABLE IF NOT EXISTS service_requests (
    id VARCHAR(64) PRIMARY KEY,
    restaurant_id VARCHAR(64) REFERENCES restaurants(id) ON DELETE CASCADE NOT NULL,
    table_id VARCHAR(64) REFERENCES tables(id) ON DELETE CASCADE NOT NULL,
    type VARCHAR(50) NOT NULL, -- CALL_WAITER, BILL, RECEIPT, HELP
    status VARCHAR(50) DEFAULT 'PENDING' NOT NULL, -- PENDING, IN_PROGRESS, RESOLVED, CANCELLED
    notes TEXT,
    resolved_by VARCHAR(64) REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_service_requests_resto ON service_requests(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_service_requests_status ON service_requests(status);
