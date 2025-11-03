START TRANSACTION;

-- let's lock the table. Useful for debugging and performance tuning.
LOCK TABLE schema.table IN OWNER MODE;

COMMIT;

-- transactions are waiting for locks. Useful for debugging and performance tuning.
SELECT *
FROM locks
WHERE grant_timestamp is null
ORDER BY request_timestamp DESC;