-- 000002_cashier_shift.down.sql
-- Revert Cashier Shift Management Schema

DROP TABLE IF EXISTS refunds CASCADE;
DROP TABLE IF EXISTS shift_closings CASCADE;
DROP TABLE IF EXISTS cashier_shift_transactions CASCADE;
DROP TABLE IF EXISTS cashier_shifts CASCADE;
DROP TABLE IF EXISTS registers CASCADE;
