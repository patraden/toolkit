-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- 03_license.sql
--
-- Verify a Vertica license is installed. For CE the license is self-issued by
-- `install_vertica --license CE` and has end_date = 'Perpetual'. A missing
-- license row means install_vertica never ran (or failed silently).
--
-- `end_date` in v_catalog.licenses is VARCHAR (can be 'Perpetual' or a date
-- string), so we keep this check simple: at least one license row present.
--
-- Emits:
--   PASS                       — at least one license row present
--   FAIL:no_license_row        — v_catalog.licenses empty
SELECT CASE
    WHEN COUNT(*) > 0 THEN 'PASS'
    ELSE 'FAIL:no_license_row'
END
FROM v_catalog.licenses;
