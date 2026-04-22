-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- Verify a Vertica license is installed. 
SELECT CASE
    WHEN COUNT(*) > 0 THEN 'PASS'
    ELSE 'FAIL:no_license_row'
END
FROM v_catalog.licenses;
