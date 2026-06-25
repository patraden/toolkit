#!/bin/bash
set -euo pipefail

usage() {
	echo "Usage: $(basename "$0") <csv-file>" >&2
	exit 1
}

[[ $# -eq 1 ]] || usage
[[ -f "$1" ]] || { echo "File not found: $1" >&2; exit 1; }

# Jira table: header row uses ||, data rows use |
sed -En '1{s/,/||/g;s/^/||/;s/$/||/;p;};1!{s/,/|/g;s/^/|/;s/$/|/;p;}' "$1"
