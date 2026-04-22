-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- Verify K-safety of the 3-node cluster. 
SELECT CASE
    WHEN designed_fault_tolerance = 1
     AND current_fault_tolerance  = 1 THEN 'PASS'
    ELSE 'FAIL:designed=' || designed_fault_tolerance::VARCHAR
         || '_current='   || current_fault_tolerance::VARCHAR
END
FROM v_monitor.system;
