-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- 01_cluster_up.sql
--
-- Verify the cluster is formed and every node reports node_state = 'UP'.
-- Expected: 3 UP nodes (matches VERTICA_COMPOSE_CONTAINERS in the Makefile).
--
-- Emits exactly one row:
--   PASS                      — 3 UP nodes
--   FAIL:up_nodes=<N>         — wrong UP count (includes total)
SELECT CASE
    WHEN up_nodes = 3 AND total_nodes = 3 THEN 'PASS'
    ELSE 'FAIL:up_nodes=' || up_nodes::VARCHAR
         || '_total_nodes=' || total_nodes::VARCHAR
END
FROM (
    SELECT
        SUM(CASE WHEN node_state = 'UP' THEN 1 ELSE 0 END) AS up_nodes,
        COUNT(*)                                            AS total_nodes
    FROM v_catalog.nodes
) AS node_counts;
