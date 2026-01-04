-- Add unique constraint on reference_id to enforce idempotency
-- This prevents duplicate transaction processing when network fails

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_reference_unique 
ON transactions(reference_id);
