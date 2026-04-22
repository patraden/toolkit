# Copyright 2026 Denis Patrakhin
# SPDX-License-Identifier: Apache-2.0
#
# Three-node Vertica CE cluster on macOS (Colima + Docker).
# All targets assume you run `make` from this directory.
#
# Typical happy path:
#   make docker-build-node   # build the node image (RPM + catalog/data dirs baked in)
#   make compose-up          # start vertica1/vertica2/vertica3
#   make install-vertica     # run install_vertica on vertica1 (as root)
#   make create-db           # admintools -t create_db on vertica1 (as dbadmin)
#   make load-vmart          # generate + load the VMart sample schema
#   make test                # (optional) run the 3-node smoke/cluster tests

# ---- image / build args -----------------------------------------------------
#
# This Makefile is the single source of truth for every variable below.
# Docker Compose is invoked with `--env-file /dev/null` so it does NOT
# auto-load any `.env` file from this directory — values reach compose only
# via the explicit env prefix on the `docker compose` command lines below.
#
# To override defaults:
#   - `make VAR=value ...` on the command line (highest precedence), or
#   - export them in your shell before running make, e.g.
#       source ./.env.example && make compose-up
#     (`.env.example` ships with `export KEY=value` lines, so a plain `source`
#     is enough; it is NOT loaded automatically — it exists purely as a
#     reference/`source`-able helper.)
#
# `export` below makes every Makefile variable available in the environment
# of recipe subprocesses, so both `make FOO=bar compose-up` and
# `FOO=bar make compose-up` propagate FOO into docker compose.
export

TAG              ?= latest
IMAGE_NODE_NAME  ?= vertica-ce-node
VERTICA_NODE_IMAGE = $(IMAGE_NODE_NAME):$(TAG)

OS_IMAGE   ?= almalinux
OS_VERSION ?= 8.10

# Auto-detect the Vertica RPM under packages/. If TAG is a version string
# (e.g. 11.0.0-3), restrict the glob to that version.
ifeq ($(TAG),latest)
VERTICA_PACKAGE ?= $(notdir $(wildcard packages/vertica*.rpm))
else
VERTICA_PACKAGE ?= $(notdir $(wildcard packages/vertica-$(TAG)*.rpm))
endif

ifeq ($(VERTICA_PACKAGE),)
$(error No Vertica RPM found under packages/ (expected packages/vertica*.rpm or packages/vertica-$(TAG)*.rpm))
endif

# Vertica DB user / group / UID / GID / DB name. These flow BOTH as build args
# (so the image pre-creates catalog/data dirs with the right owner and name)
# AND as runtime values for install-vertica / create-db / load-vmart. Keep
# them in sync — changing VERTICA_DB_NAME means rebuilding the image.
VERTICA_DB_UID      ?= 1000
VERTICA_DB_GID      ?= 1000
VERTICA_DB_USER     ?= dbadmin
VERTICA_DB_GROUP    ?= verticadba
VERTICA_DB_NAME     ?= dockerdb
VERTICA_DB_PASSWORD ?=

BUILD_ARGS = \
    --build-arg os_image=$(OS_IMAGE) \
    --build-arg os_version=$(OS_VERSION) \
    --build-arg vertica_package=$(VERTICA_PACKAGE) \
    --build-arg dbadmin_id=$(VERTICA_DB_UID) \
    --build-arg dbadmin_gid=$(VERTICA_DB_GID) \
    --build-arg vertica_db_user=$(VERTICA_DB_USER) \
    --build-arg vertica_db_group=$(VERTICA_DB_GROUP) \
    --build-arg vertica_db_name=$(VERTICA_DB_NAME)

# ---- cluster runtime defaults ----------------------------------------------

VERTICA_COMPOSE_CONTAINERS ?= vertica1 vertica2 vertica3
VERTICA_LEADER             ?= vertica1
VERTICA_HOSTS              ?= 172.28.0.11,172.28.0.12,172.28.0.13

# Must match the Dockerfile: ${VERTICA_DATA_DIR}/${vertica_db_name}_catalog and
# ${VERTICA_DATA_DIR}/${vertica_db_name}_data (two sibling dirs).
# Note: VERTICA_DATA_ROOT here corresponds to the Dockerfile ENV VERTICA_DATA_DIR
# (deliberately renamed to avoid overloading the name: here VERTICA_DB_DATA_DIR
# is the per-DB data dir, not the parent).
VERTICA_DATA_ROOT      ?= /data/vertica
VERTICA_DB_CATALOG_DIR ?= $(VERTICA_DATA_ROOT)/$(VERTICA_DB_NAME)_catalog
VERTICA_DB_DATA_DIR    ?= $(VERTICA_DATA_ROOT)/$(VERTICA_DB_NAME)_data

# Path to the RPM inside each container (baked in by the Dockerfile).
VERTICA_INSTALL_RPM_PATH ?= /opt/vertica/install-rpm/$(VERTICA_PACKAGE)

# ---- phony targets ----------------------------------------------------------

.PHONY: all display docker-build-node compose-up compose-down \
        install-vertica create-db load-vmart \
        shell logs vsql clean test

all: docker-build-node

display:
	$(info Vertica cluster - build parameters)
	$(info   IMAGE_NODE_NAME     = $(IMAGE_NODE_NAME))
	$(info   TAG                 = $(TAG))
	$(info   VERTICA_NODE_IMAGE  = $(VERTICA_NODE_IMAGE))
	$(info   OS_IMAGE            = $(OS_IMAGE):$(OS_VERSION))
	$(info   VERTICA_PACKAGE     = $(VERTICA_PACKAGE))
	$(info   VERTICA_DB_USER     = $(VERTICA_DB_USER))
	$(info   VERTICA_DB_GROUP    = $(VERTICA_DB_GROUP))
	$(info   VERTICA_DB_UID      = $(VERTICA_DB_UID))
	$(info   VERTICA_DB_GID      = $(VERTICA_DB_GID))
	$(info   VERTICA_DB_NAME     = $(VERTICA_DB_NAME))
	$(info Runtime)
	$(info   CONTAINERS          = $(VERTICA_COMPOSE_CONTAINERS))
	$(info   LEADER              = $(VERTICA_LEADER))
	$(info   VERTICA_HOSTS       = $(VERTICA_HOSTS))
	$(info   CATALOG DIR         = $(VERTICA_DB_CATALOG_DIR))
	$(info   DATA DIR            = $(VERTICA_DB_DATA_DIR))
	$(info   RPM IN CONTAINER    = $(VERTICA_INSTALL_RPM_PATH))
	@:

# ---- image build ------------------------------------------------------------

docker-build-node: packages/$(VERTICA_PACKAGE) display Dockerfile
	docker build \
	    -f Dockerfile \
	    $(BUILD_ARGS) \
	    -t $(VERTICA_NODE_IMAGE) .

# ---- compose ----------------------------------------------------------------

# --env-file /dev/null disables compose's automatic `.env` auto-load. Every
# variable that compose needs is forwarded via the `export` directive above,
# which puts all Makefile variables into the environment of the recipe shell.
COMPOSE ?= docker compose --env-file /dev/null

compose-up:
	$(COMPOSE) up -d

# Safe to run: admintools config and dbadmin's ~/.ssh are redirected onto the
# /data named volume by the entrypoint, so a compose-down + compose-up cycle
# preserves the cluster. Use `make clean` below for a full reset.
compose-down:
	$(COMPOSE) down

# Full reset: stop + remove containers AND the /data volumes. You'll need to
# re-run install-vertica / create-db / load-vmart after this.
clean:
	$(COMPOSE) down -v

# ---- cluster bring-up helpers ----------------------------------------------

# Run install_vertica on the leader as root. The RPM is baked into each image
# at $(VERTICA_INSTALL_RPM_PATH). --no-system-configuration skips sysctl /
# limits.conf tweaks that can't be applied from inside a container, and
# --failure-threshold NONE downgrades hard-stop preflight checks (THP,
# swappiness, etc.) to warnings so install proceeds in the container env.
install-vertica:
	docker exec -it $(VERTICA_LEADER) bash -lc '/opt/vertica/sbin/install_vertica \
	    --accept-eula \
	    --license CE \
	    --data-dir $(VERTICA_DATA_ROOT) \
	    --dba-group $(VERTICA_DB_GROUP) \
	    --dba-user $(VERTICA_DB_USER) \
	    --dba-user-password-disabled \
	    --debug \
	    --failure-threshold NONE \
	    --hosts $(VERTICA_HOSTS) \
	    --no-system-configuration \
	    --rpm $(VERTICA_INSTALL_RPM_PATH)'

# Create the database as dbadmin. Use VERTICA_DB_PASSWORD= make ... to set it
# (defaults to empty; admintools still requires -p so we pass '' when unset).
# Catalog / data dirs are pre-created by the Dockerfile; admintools just uses them.
create-db:
	docker exec -u $(VERTICA_DB_USER) -w /home/$(VERTICA_DB_USER) -it $(VERTICA_LEADER) \
	    /opt/vertica/bin/admintools -t create_db \
	        --hosts "$(VERTICA_HOSTS)" \
	        --catalog_path $(VERTICA_DB_CATALOG_DIR) \
	        --data_path $(VERTICA_DB_DATA_DIR) \
	        -d $(VERTICA_DB_NAME) \
	        -p '$(VERTICA_DB_PASSWORD)' \
	        -i

# Generate the VMart sample data and load it. Uses the image-baked
# ${VMART_DIR}/01_load_vmart_schema.sh which in turn calls vmart_gen + vsql.
load-vmart:
	docker exec -u $(VERTICA_DB_USER) -w /home/$(VERTICA_DB_USER) \
	    -e VERTICA_DB_NAME=$(VERTICA_DB_NAME) \
	    -it $(VERTICA_LEADER) \
	    bash -lc '"$$VMART_DIR/$$VMART_ETL_SCRIPT"'

# ---- convenience ------------------------------------------------------------

CONTAINER ?= $(VERTICA_LEADER)

shell:
	docker exec -it $(CONTAINER) bash -l

logs:
	$(COMPOSE) logs -f $(CONTAINER)

vsql:
	docker exec -u $(VERTICA_DB_USER) -it $(CONTAINER) \
	    /opt/vertica/bin/vsql -U $(VERTICA_DB_USER) -d $(VERTICA_DB_NAME)

# ---- tests ------------------------------------------------------------------
#
# Each tests/*.sql file MUST emit exactly one line matching
#   ^(PASS|FAIL:<reason>|SKIP:<reason>)$
# on a 3-node cluster. The harness runs them in alphabetical order via vsql
# (streamed over stdin so no host-volume mount is required), greps for the
# first PASS/FAIL/SKIP line, and tallies the results. Anything else (vsql
# errors, empty output) counts as a FAIL.
#
# `make test` is optional — it is NOT a dependency of any other target. Run it
# after `make create-db` (for tests 01-04) and/or `make load-vmart`
# (for tests 05-07) to sanity-check the cluster.

TEST_DIR   ?= tests
TEST_FILES ?= $(sort $(wildcard $(TEST_DIR)/*.sql))

test:
	@if [ -z "$(TEST_FILES)" ]; then \
	    echo "No SQL test files found under $(TEST_DIR)/"; \
	    exit 1; \
	fi; \
	failed=0; skipped=0; total=0; \
	for f in $(TEST_FILES); do \
	    name=$$(basename $$f .sql); \
	    total=$$((total+1)); \
	    out=$$(docker exec -u $(VERTICA_DB_USER) -i $(VERTICA_LEADER) \
	        /opt/vertica/bin/vsql -U $(VERTICA_DB_USER) -d $(VERTICA_DB_NAME) \
	        -Atq 2>&1 < "$$f" \
	        | grep -E '^(PASS|FAIL|SKIP)' | head -n 1); \
	    case "$$out" in \
	        PASS) \
	            printf "  PASS  %s\n" "$$name" ;; \
	        SKIP*) \
	            printf "  SKIP  %s (%s)\n" "$$name" "$$out"; \
	            skipped=$$((skipped+1)) ;; \
	        FAIL*) \
	            printf "  FAIL  %s (%s)\n" "$$name" "$$out"; \
	            failed=$$((failed+1)) ;; \
	        *) \
	            printf "  FAIL  %s (no PASS/FAIL/SKIP emitted — check vsql output)\n" "$$name"; \
	            failed=$$((failed+1)) ;; \
	    esac; \
	done; \
	echo; \
	printf "Results: %d passed, %d skipped, %d failed (of %d)\n" \
	    "$$((total-failed-skipped))" "$$skipped" "$$failed" "$$total"; \
	exit $$failed
