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
#   1. Persist admintools state and dbadmin's home dir onto the /data volume
#      via symlinks (see preserve_config + preserve_dbadmin_home below), so
#      that `docker compose down` + `up` does not erase the cluster's
#      topology / ssh keys.
#   2. Start sshd (install_vertica needs passwordless root SSH between nodes).
#   3. On the cluster leader (VERTICA_CLUSTER_ROLE=leader), if the database
#      has already been created (`admintools -t list_db -d $VERTICA_DB_NAME`
#      returns zero), bring the DB back up with `admintools -t start_db`.
#      Followers stay idle.
#   4. On SIGTERM/SIGINT/SIGHUP, the leader gracefully stops the DB with
#      `admintools -t stop_db` before the container exits. Followers just exit.
#
# Cluster formation (install_vertica + create_db) is still manual the first
# time a /data volume is created — see README.md. After that, the cluster
# survives both `compose stop/start` and `compose down/up`; only `make clean`
# (which removes the /data volumes too) forces a re-install.

set -uo pipefail

VERTICA_DB_USER="${VERTICA_DB_USER:-dbadmin}"
VERTICA_DB_NAME="${VERTICA_DB_NAME:-dockerdb}"
VERTICA_OPT_DIR="${VERTICA_OPT_DIR:-/opt/vertica}"
VERTICA_CLUSTER_ROLE="${VERTICA_CLUSTER_ROLE:-follower}"

ADMINTOOLS="${VERTICA_OPT_DIR}/bin/admintools"

log() { echo "[entrypoint] $*"; }

# Redirect admintools state onto the /data volume so it survives
# `docker compose down`. First boot: seed /data/config from the pristine
# /opt/vertica/config that shipped with the image. Every subsequent boot:
# just (re)create the symlink — the fresh container has a new plain
# /opt/vertica/config from the image that we need to replace. Idempotent.
#
# Adapted from upstream vertica/vertica-containers one-node-ce preserve_config()
# (https://github.com/vertica/vertica-containers/blob/main/one-node-ce/docker-entrypoint.sh).
preserve_config() {
    if [[ ! -d /data/config ]]; then
        log "First boot: seeding /data/config from /opt/vertica/config"
        cp -a /opt/vertica/config /data/config
    fi
    if [[ ! -L /opt/vertica/config ]]; then
        log "Redirecting /opt/vertica/config -> /data/config"
        rm -rf /opt/vertica/config
        ln -s /data/config /opt/vertica/config
    fi
}

# Same trick for dbadmin's home directory. install_vertica puts dbadmin's
# inter-node SSH keys in ~dbadmin/.ssh/; `admintools -t start_db/stop_db`
# rides on those keys to coordinate the cluster. If /home/dbadmin is just a
# normal dir inside the container filesystem, it vanishes on `compose down`
# and the next `start_db` fails with "Permission denied (publickey)" even
# though /data/config is intact. Upstream doesn't need this (single-node),
# so it's our extension.
preserve_dbadmin_home() {
    if [[ ! -d /data/dbadmin_home ]]; then
        log "First boot: seeding /data/dbadmin_home from /home/${VERTICA_DB_USER}"
        cp -a "/home/${VERTICA_DB_USER}" /data/dbadmin_home
    fi
    if [[ ! -L "/home/${VERTICA_DB_USER}" ]]; then
        log "Redirecting /home/${VERTICA_DB_USER} -> /data/dbadmin_home"
        rm -rf "/home/${VERTICA_DB_USER}"
        ln -s /data/dbadmin_home "/home/${VERTICA_DB_USER}"
    fi
}

# Ask admintools directly whether the DB is registered on this host. In
# Vertica 11 `admintools -t list_db -d <name>` returns zero iff the DB is
# known to admintools.conf; the bare `list_db` (no -d) fails with an error
# so we can't use it as a "list all" probe.
db_is_configured() {
    [[ -x "${ADMINTOOLS}" ]] || return 1
    su - "${VERTICA_DB_USER}" -c \
        "${ADMINTOOLS} -t list_db -d ${VERTICA_DB_NAME}" \
        >/dev/null 2>&1
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

preserve_config
preserve_dbadmin_home

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
