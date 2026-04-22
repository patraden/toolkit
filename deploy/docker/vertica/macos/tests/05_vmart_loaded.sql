-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- Verify the VMart sample schema is fully loaded. Looks at three canonical
-- fact tables (store_sales_fact, online_sales_fact, inventory_fact) and
-- checks their *live* row counts against the `vmart_gen` defaults. Skips
-- cleanly if VMart was never loaded.
--
-- VMart defaults (from vmart_gen):
--   store.store_sales_fact         = 5,000,000 rows
--   online_sales.online_sales_fact = 5,000,000 rows
--   public.inventory_fact          =   300,000 rows
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
