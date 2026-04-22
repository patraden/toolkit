-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- 02_ksafety.sql
--
-- Verify K-safety of the 3-node cluster. `admintools -t create_db` on 3 nodes
-- defaults to K=1 (one-node failure tolerated, buddy projections). We check
-- BOTH values match: designed_fault_tolerance stays constant across restarts,
-- current_fault_tolerance drops if a node is down.
--
-- Emits:
--   PASS                                        — designed=1 AND current=1
--   FAIL:designed=<D>_current=<C>               — otherwise
SELECT CASE
    WHEN designed_fault_tolerance = 1
     AND current_fault_tolerance  = 1 THEN 'PASS'
    ELSE 'FAIL:designed=' || designed_fault_tolerance::VARCHAR
         || '_current='   || current_fault_tolerance::VARCHAR
END
FROM v_monitor.system;
