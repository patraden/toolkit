START TRANSACTION;

-- current trx id
SELECT current_trans_id(), current_statement();
-- isolation level (not always reflect the current level)
SHOW CURRENT TransactionIsolationLevel;
-- let's lock the table
LOCK TABLE schema.table IN OWNER MODE;

COMMIT;