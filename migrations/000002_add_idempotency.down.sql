-- Rollback unique constraint on reference_id
DROP INDEX IF EXISTS idx_transactions_reference_unique;
