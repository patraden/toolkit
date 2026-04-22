-- (c) Copyright [2021-2023] Open Text.
-- Copyright 2026 Denis Patrakhin (modifications to this file)
-- SPDX-License-Identifier: Apache-2.0
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- Based on: https://github.com/vertica/vertica-containers/blob/main/one-node-ce/tests/flex_table_loaded.sql
--
-- 04_flex_table_loaded.sql
--
-- Verify the bundled FlexTableLib UDx library is registered on the cluster.
-- `install_vertica` + `create_db` should auto-register it; if it's missing
-- something went wrong during image/DB bootstrap.
--
-- Adapted from vertica/vertica-containers one-node-ce tests/flex_table_loaded.sql.
-- The upstream `__MD5__` placeholder check is dropped — we don't bake a pinned
-- md5 here, so we just assert the library is linked and has >0 functions in the
-- manifest.
--
-- Emits:
--   PASS                                 — FlexTableLib is loaded with >=1 fn
--   FAIL:flex_table_lib_not_registered   — no match in user_libraries/manifest
SELECT CASE
    WHEN COUNT(*) > 0 THEN 'PASS'
    ELSE 'FAIL:flex_table_lib_not_registered'
END
FROM v_monitor.user_libraries ul
JOIN v_monitor.user_library_manifest ulm
  ON ul.lib_name = ulm.lib_name
WHERE ulm.lib_name = 'FlexTableLib';
