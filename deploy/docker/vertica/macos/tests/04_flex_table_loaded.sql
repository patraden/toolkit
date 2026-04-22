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
SELECT CASE
    WHEN COUNT(*) > 0 THEN 'PASS'
    ELSE 'FAIL:flex_table_lib_not_registered'
END
FROM v_monitor.user_libraries ul
JOIN v_monitor.user_library_manifest ulm
  ON ul.lib_name = ulm.lib_name
WHERE ulm.lib_name = 'FlexTableLib';
