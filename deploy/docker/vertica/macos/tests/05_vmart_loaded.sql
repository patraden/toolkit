-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- 05_vmart_loaded.sql
--
-- Verify the VMart sample schema is fully loaded. Looks at three canonical
-- fact tables (store_sales_fact, online_sales_fact, inventory_fact) and
-- checks their *live* row counts against the `vmart_gen` defaults. Skips
-- cleanly if VMart was never loaded.
--
-- Implementation notes:
--   * We never reference the VMart tables directly — only v_monitor +
--     v_catalog system tables — so this SQL parses even if VMart was
--     never loaded (the tables / schemas may not exist).
--   * v_monitor.projection_storage.row_count INCLUDES tombstones. VMart's
--     `02_vmart_etl.sql` runs a full-table UPDATE on store_sales_fact which
--     leaves 5M deleted rows alongside the 5M live ones. Use
--     v_monitor.storage_containers.(total_row_count - deleted_row_count)
--     instead for accurate live-row counts.
--   * Storage counts are per buddy projection; divide by (K+1) to get the
--     underlying table rowcount.
--
-- VMart defaults (from vmart_gen):
--   store.store_sales_fact         = 5,000,000 rows
--   online_sales.online_sales_fact = 5,000,000 rows
--   public.inventory_fact          =   300,000 rows
--
-- Emits:
--   PASS                                              — all three counts match
--   SKIP:vmart_not_loaded                             — sentinel table missing
--   FAIL:store=<N>_online=<N>_inv=<N>                 — row counts mismatch
WITH ksafety AS (
    SELECT GREATEST(designed_fault_tolerance, 0) + 1 AS replicas
    FROM v_monitor.system
),
sentinel AS (
    SELECT CASE WHEN COUNT(*) > 0 THEN TRUE ELSE FALSE END AS loaded
    FROM v_catalog.tables
    WHERE table_schema = 'public'
      AND table_name   = 'vmart_load_success'
),
-- Per-table live-row sums (across all projections + all nodes).
live_rows AS (
    SELECT p.projection_schema AS sch,
           p.anchor_table_name AS tbl,
           SUM(sc.total_row_count - sc.deleted_row_count) AS rows_sum
    FROM v_monitor.storage_containers sc
    JOIN v_catalog.projections p
      ON p.projection_id = sc.projection_id
    WHERE p.is_super_projection = TRUE
      AND (   (p.projection_schema = 'store'        AND p.anchor_table_name = 'store_sales_fact')
           OR (p.projection_schema = 'online_sales' AND p.anchor_table_name = 'online_sales_fact')
           OR (p.projection_schema = 'public'       AND p.anchor_table_name = 'inventory_fact'))
    GROUP BY 1, 2
),
pivoted AS (
    SELECT
        COALESCE(MAX(CASE WHEN sch='store'        AND tbl='store_sales_fact'  THEN rows_sum END), 0) AS store_sum,
        COALESCE(MAX(CASE WHEN sch='online_sales' AND tbl='online_sales_fact' THEN rows_sum END), 0) AS online_sum,
        COALESCE(MAX(CASE WHEN sch='public'       AND tbl='inventory_fact'    THEN rows_sum END), 0) AS inv_sum
    FROM live_rows
)
SELECT CASE
    WHEN NOT s.loaded THEN 'SKIP:vmart_not_loaded'
    WHEN p.store_sum  / k.replicas <> 5000000
      OR p.online_sum / k.replicas <> 5000000
      OR p.inv_sum    / k.replicas <>  300000
        THEN 'FAIL:store=' || (p.store_sum  / k.replicas)::VARCHAR
              || '_online=' || (p.online_sum / k.replicas)::VARCHAR
              || '_inv='    || (p.inv_sum    / k.replicas)::VARCHAR
    ELSE 'PASS'
END
FROM pivoted p CROSS JOIN ksafety k CROSS JOIN sentinel s;
