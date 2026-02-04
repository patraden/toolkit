#!/usr/bin/env bash
# Usage: ./list_schedulers.sh schema [--html]
# Required environment variables:
#   VERTICA_HOST     - host of Vertica server
# Optional environment variables:
#   VERTICA_PORT     - port (default 5433)
#   VERTICA_USER     - username (default dbadmin)
#   VERTICA_PASSWORD - password (can be empty)
#   VERTICA_DB       - database name (default bi)
#
# Lists schedulers (stream microbatches, load specs, sources, targets) for the given schema.
# By default uses vsql default formatting. Pass --html for HTML table output.
# Uses sql/vertica/schedulers/list.sql with vsql variable :schema.

set -euo pipefail

SCHEMA=${1:-}
HTML_FLAG=${2:-}
if [[ -z "$SCHEMA" ]]; then
    echo "Usage: $0 schema [--html]"
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
VERTICA_PASSWORD="${VERTICA_PASSWORD:-}"

# run from project root
ROOT_DIR="."
LIST_SQL="${ROOT_DIR}/sql/vertica/schedulers/list.sql"

if [[ ! -f "$LIST_SQL" ]]; then
    echo "SQL file not found: $LIST_SQL"
    exit 1
fi

set --
[[ "$HTML_FLAG" == "--html" ]] && set -- -H
vsql -h "$VERTICA_HOST" \
     -p "$VERTICA_PORT" \
     -U "$VERTICA_USER" \
     -w "$VERTICA_PASSWORD" \
     -d "$VERTICA_DB" \
     "$@" \
     -v schema="$SCHEMA" \
     -f "$LIST_SQL"
