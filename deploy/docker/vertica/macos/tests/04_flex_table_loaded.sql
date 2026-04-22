-- Copyright 2026 Denis Patrakhin
-- SPDX-License-Identifier: Apache-2.0
--
-- Functional flex-table smoke test: create a flex table, ingest JSON via
-- FJSONPARSER, compute virtual keys, and verify MapLookup-based virtual-column
-- extraction against the ingested rows. This implicitly covers that the
-- FlexTableLib UDx library is registered and loadable on every node
-- (CREATE FLEX TABLE, FJSONPARSER, COMPUTE_FLEXTABLE_KEYS, and MapLookup all
-- resolve against it).
DROP TABLE IF EXISTS public.test04_flex CASCADE;

CREATE FLEX TABLE public.test04_flex();

COPY public.test04_flex FROM STDIN PARSER FJSONPARSER() ABORT ON ERROR;
{"name":"alice","n":1}
{"name":"bob","n":2}
{"name":"carol","n":3}
\.

SELECT COMPUTE_FLEXTABLE_KEYS('public.test04_flex');

WITH extracted AS (
    SELECT MapLookup(__raw__, 'name')::VARCHAR AS name
    FROM public.test04_flex
)
SELECT CASE
    WHEN COUNT(*) = 3
     AND SUM((name IS NOT NULL)::INT) = 3
     AND MAX(name) = 'carol'
        THEN 'PASS'
    ELSE 'FAIL:c=' || COUNT(*)::VARCHAR
         || '_named=' || SUM((name IS NOT NULL)::INT)::VARCHAR
         || '_maxname=' || COALESCE(MAX(name), 'null')
END
FROM extracted;

DROP TABLE public.test04_flex CASCADE;
