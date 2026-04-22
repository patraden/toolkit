-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- Verify data for a segmented fact table actually lands on every node in the
-- cluster (i.e. segmentation + rebalancing work on 3 nodes, not just 1).
-- Uses store.store_sales_fact (5M rows) as the witness table.
WITH sentinel AS (
    SELECT CASE WHEN COUNT(*) > 0 THEN TRUE ELSE FALSE END AS loaded
    FROM v_catalog.tables
    WHERE table_schema = 'store'
      AND table_name   = 'store_sales_fact'
),
nodes_with_data AS (
    SELECT COUNT(DISTINCT node_name) AS n
    FROM v_monitor.projection_storage
    WHERE anchor_table_schema = 'store'
      AND anchor_table_name   = 'store_sales_fact'
      AND row_count > 0
)
SELECT CASE
    WHEN NOT (SELECT loaded FROM sentinel) THEN 'SKIP:vmart_not_loaded'
    WHEN (SELECT n FROM nodes_with_data) = 3 THEN 'PASS'
    ELSE 'FAIL:distinct_nodes_with_rows=' || (SELECT n FROM nodes_with_data)::VARCHAR
END;
