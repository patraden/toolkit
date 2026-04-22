#!/usr/bin/env bash
# Copyright 2026 Denis Patrakhin
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
# https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Inspired by — but largely rewritten from —
# https://github.com/vertica/vertica-containers/blob/main/one-node-ce/docker-entrypoint.sh
# (c) Copyright [2021-2023] Open Text. The upstream single-node entrypoint
# creates the database, loads VMart, manages app users and MC agent; our
# 3-node version delegates create_db / install_vertica to manual `make`
# targets and only handles sshd, auto-restart on the leader, and a SIGTERM
# trap. No upstream code is copied verbatim.
#
# Node entrypoint for the three-node Vertica cluster on macOS.
#
# Runs as PID 1 as root. Responsibilities:
#   1. Start sshd (install_vertica needs passwordless root SSH between nodes).
#   2. On the cluster leader (VERTICA_CLUSTER_ROLE=leader), if the database has
#      already been created (admintools -t list_db knows VERTICA_DB_NAME), bring
#      the DB back up with `admintools -t start_db`. Followers stay idle.
#   3. On SIGTERM/SIGINT/SIGHUP, the leader gracefully stops the DB with
#      `admintools -t stop_db` before the container exits. Followers just exit.
#
# Cluster formation (install_vertica + create_db) is manual — see README.md.
#
# Known limitation (vs. the upstream one-node-ce entrypoint):
#   admintools state lives inside the image under /opt/vertica/config/, not on
#   the /data volume. `docker compose restart` and `stop/start` preserve it
#   (container filesystem is intact), but `docker compose down` removes
#   containers, so the next `up` recreates them from the image and admintools
#   forgets the cluster. The DB files on /data then become orphaned and you
#   must re-run `make install-vertica` + `make create-db`. To survive
#   `compose down`, port the upstream preserve_config() pattern (symlink
#   /opt/vertica/config -> /data/config on first boot).

set -uo pipefail

VERTICA_DB_USER="${VERTICA_DB_USER:-dbadmin}"
VERTICA_DB_NAME="${VERTICA_DB_NAME:-dockerdb}"
VERTICA_OPT_DIR="${VERTICA_OPT_DIR:-/opt/vertica}"
VERTICA_CLUSTER_ROLE="${VERTICA_CLUSTER_ROLE:-follower}"

ADMINTOOLS="${VERTICA_OPT_DIR}/bin/admintools"

log() { echo "[entrypoint] $*"; }

# Ask admintools directly whether the DB is registered on this host. This is
# robust across Vertica versions (the layout of admintools.conf has changed
# between releases, so grepping it is fragile).
db_is_configured() {
    [[ -x "${ADMINTOOLS}" ]] || return 1
    su - "${VERTICA_DB_USER}" -c "${ADMINTOOLS} -t list_db 2>/dev/null" \
        | grep -qw "${VERTICA_DB_NAME}"
}

db_is_active() {
    local active
    active=$(su - "${VERTICA_DB_USER}" -c "${ADMINTOOLS} -t show_active_db" 2>/dev/null || true)
    [[ -n "${active// /}" ]]
}

start_db() {
    if ! db_is_configured; then
        log "admintools.conf has no entry for ${VERTICA_DB_NAME}; skipping auto start_db (run 'make install-vertica' + 'make create-db')."
        return 0
    fi
    log "Starting database ${VERTICA_DB_NAME} via admintools"
    su - "${VERTICA_DB_USER}" -c \
        "${ADMINTOOLS} -t start_db -d ${VERTICA_DB_NAME} --noprompts" \
        || log "start_db returned non-zero (DB may already be up or peers not ready)"
}

stop_db() {
    if ! db_is_configured; then
        return 0
    fi
    if ! db_is_active; then
        log "Database ${VERTICA_DB_NAME} is not active; nothing to stop."
        return 0
    fi
    log "Stopping database ${VERTICA_DB_NAME} via admintools"
    su - "${VERTICA_DB_USER}" -c \
        "${ADMINTOOLS} -t stop_db -d ${VERTICA_DB_NAME} -i" \
        || log "stop_db returned non-zero"
}

shutdown() {
    log "Received shutdown signal (role=${VERTICA_CLUSTER_ROLE})"
    if [[ "${VERTICA_CLUSTER_ROLE}" == "leader" ]]; then
        stop_db
    fi
    exit 0
}

trap shutdown SIGTERM SIGINT SIGHUP

mkdir -p /var/run/sshd
/usr/sbin/sshd
log "sshd started"

if [[ "${VERTICA_CLUSTER_ROLE}" == "leader" ]]; then
    start_db
else
    log "role=follower; waiting for leader to manage the database"
fi

sleep infinity &
wait $!
