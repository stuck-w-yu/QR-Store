-- 000003_phase1_cleanup.down.sql
DROP TABLE IF EXISTS service_requests;

ALTER TABLE payments DROP COLUMN IF EXISTS verified_at;
ALTER TABLE payments DROP COLUMN IF EXISTS verified_by;
ALTER TABLE payments DROP COLUMN IF EXISTS reference_number;
ALTER TABLE payments DROP COLUMN IF EXISTS change_amount;
ALTER TABLE payments DROP COLUMN IF EXISTS paid_amount;

ALTER TABLE orders DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE orders DROP COLUMN IF EXISTS completed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_method;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_status;
