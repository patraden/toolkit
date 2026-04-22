-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- Verify Vertica's hash-segmented projections actually balance rows evenly
-- across nodes — i.e. the cluster is not just "3 nodes up" but also
-- parallel-capable (queries will touch all 3 nodes with roughly equal work).
--
-- Uses store.store_sales_fact (5M rows) and picks a single deterministic
-- buddy projection (lowest projection_name alphabetically) so the row counts
-- aren't inflated by K-safety buddies. Then computes the coefficient of
-- variation (stddev / mean) across the 3 nodes. For a correctly hash-
-- segmented large fact table this should be effectively 0 (<0.01); we allow
-- up to 0.20 (20 %) to avoid flakiness on small machines.
WITH sentinel AS (
    SELECT CASE WHEN COUNT(*) > 0 THEN TRUE ELSE FALSE END AS loaded
    FROM v_catalog.tables
    WHERE table_schema = 'store'
      AND table_name   = 'store_sales_fact'
),
base_projection AS (
    SELECT MIN(projection_name) AS pname
    FROM v_catalog.projections
    WHERE projection_schema    = 'store'
      AND anchor_table_name    = 'store_sales_fact'
      AND is_super_projection  = TRUE
),
per_node AS (
    SELECT node_name,
           SUM(row_count) AS rc
    FROM v_monitor.projection_storage
    WHERE anchor_table_schema = 'store'
      AND anchor_table_name   = 'store_sales_fact'
      AND projection_name     = (SELECT pname FROM base_projection)
    GROUP BY node_name
),
stats AS (
    SELECT COUNT(*)           AS node_count,
           AVG(rc)             AS mean_rc,
           STDDEV(rc)          AS stddev_rc,
           CASE WHEN AVG(rc) = 0 THEN NULL
                ELSE STDDEV(rc) / AVG(rc) END AS cv
    FROM per_node
)
SELECT CASE
    WHEN NOT (SELECT loaded FROM sentinel)
        THEN 'SKIP:vmart_not_loaded'
    WHEN (SELECT node_count FROM stats) < 3
        THEN 'FAIL:only_' || (SELECT node_count FROM stats)::VARCHAR || '_nodes_have_rows'
    WHEN (SELECT cv FROM stats) IS NULL
        THEN 'FAIL:zero_rows_on_base_projection'
    WHEN (SELECT cv FROM stats) < 0.20
        THEN 'PASS'
    ELSE 'FAIL:cv=' || ROUND((SELECT cv FROM stats)::NUMERIC, 3)::VARCHAR
END;
