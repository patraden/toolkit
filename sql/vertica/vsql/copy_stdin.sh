#!/usr/bin/env bash
# Usage: cat data.csv | ./copy_stdin.sh target_table
# Required environment variables:
#   VERTICA_HOST     - host of Vertica server
#   VERTICA_PORT     - port (default 5433)
#   VERTICA_USER     - username (default dbadmin)
#   VERTICA_PASSWORD - password (can be empty)
#   VERTICA_DB       - database name (default bi)
#   VERTICA_DELIM    - field delimiter (default "#")
#   VERTICA_NULL     - string representing NULL (default "")
#   VERTICA_ENC      - enclosing character (default '"')

set -euo pipefail

TARGET_TABLE=${1:-}
if [[ -z "$TARGET_TABLE" ]]; then
    echo "Usage: $0 target_table"
    exit 1
fi

# Check required variables (allow empty password)
if [[ -z "${VERTICA_HOST+x}" ]]; then
    echo "Please set VERTICA_HOST"
    exit 1
fi

VERTICA_USER="${VERTICA_USER:-dbadmin}"
VERTICA_DB="${VERTICA_DB:-bi}"
VERTICA_PORT="${VERTICA_PORT:-5433}"
VERTICA_DELIM="${VERTICA_DELIM:-#}"
VERTICA_NULL="${VERTICA_NULL:-}"
VERTICA_ENC="${VERTICA_ENC:-\"}"
VERTICA_PASSWORD="${VERTICA_PASSWORD:-}"

# Build the COPY command
COPY_CMD=$(cat <<EOF
COPY $TARGET_TABLE
FROM STDIN
DELIMITER '$VERTICA_DELIM'
NULL AS '$VERTICA_NULL'
ENCLOSED BY '$VERTICA_ENC';
EOF
)

# Run the COPY command with vsql, reading from stdin
vsql -h "$VERTICA_HOST" \
      -p "$VERTICA_PORT" \
      -U "$VERTICA_USER" \
      -w "$VERTICA_PASSWORD" \
      -d "$VERTICA_DB" \
      -q \
      -c "$COPY_CMD"