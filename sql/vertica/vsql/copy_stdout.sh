#!/usr/bin/env bash
# Usage: ./copy_stdout.sh query.sql
# Required environment variables:
#   VERTICA_HOST     - host of Vertica server
#   VERTICA_PORT     - port (default 5433)
#   VERTICA_USER     - username (default dbadmin)
#   VERTICA_PASSWORD - password
#   VERTICA_DB       - database name (default bi)
#   VERTICA_SEP      - field separator (default "#")

set -euo pipefail

SQL_FILE=${1:-}
if [[ -z "$SQL_FILE" || ! -f "$SQL_FILE" ]]; then
    echo "Usage: $0 path/to/query.sql"
    exit 1
fi

: "${VERTICA_HOST:?Please set VERTICA_HOST}"
: "${VERTICA_PASSWORD:?Please set VERTICA_PASSWORD}"
VERTICA_USER="${VERTICA_USER:-dbadmin}"
VERTICA_DB="${VERTICA_DB:-bi}"
VERTICA_PORT="${VERTICA_PORT:-5433}"
VERTICA_SEP="${VERTICA_SEP:-#}"

vsql -h "$VERTICA_HOST" \
      -p "$VERTICA_PORT" \
      -U "$VERTICA_USER" \
      -w "$VERTICA_PASSWORD" \
      -d "$VERTICA_DB" \
      -q \
      -t \
      -A \
      -F "$VERTICA_SEP" \
      -f "$SQL_FILE"