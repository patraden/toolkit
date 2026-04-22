# Vertica 3-node cluster on macOS

A self-contained three-node Vertica Community Edition (CE) cluster that runs
under Colima + Docker on Apple Silicon. Based on
[vertica/vertica-containers/one-node-ce](https://github.com/vertica/vertica-containers/tree/main/one-node-ce),
modified for a full cluster with manual `install_vertica` and a host-provided
root SSH keypair.

Key differences from the upstream one-node example:

- Three identical node containers on a fixed subnet (`172.28.0.0/24`).
- Root SSH keypair is copied from the host (not generated per container), so
  passwordless root SSH between nodes works out of the box for `install_vertica`.
- The Vertica RPM is baked into the image at `/opt/vertica/install-rpm/` so
  `install_vertica --rpm` works without any host-side mounts at runtime.
- Catalog and data directories for the target DB are pre-created at image
  build time (`/data/vertica/${VERTICA_DB_NAME}_catalog` and `_data`, owned by
  `dbadmin`), so `admintools -t create_db` can run without a separate dir
  setup step.
- Cluster formation (`install_vertica`, `create_db`, VMart load) is driven by
  `make` targets; nothing happens automatically on container start beyond sshd
  and, for the leader, re-starting an already-created database.

## Prerequisites

1. Colima and Docker:

   ```shell
   brew install colima docker qemu lima-additional-guestagents
   colima start --arch x86_64 --cpu 6 --memory 12 --disk 100
   ```

2. Vertica CE RPM placed at `packages/vertica-*.x86_64.RHEL*.rpm`. OpenText no
   longer publishes Vertica CE on Docker Hub; download the RPM via a trial
   request or from an internal mirror.

3. A root SSH keypair placed at `.ssh/id_rsa` and `.ssh/id_rsa.pub`. Both are
   gitignored. The build copies them into `/root/.ssh/` inside the image
   (mode `0600` / `0644`) and seeds `/root/.ssh/authorized_keys`. `dbadmin`
   gets an empty `~/.ssh/` only; `install_vertica` populates it.

4. (Optional) Environment overrides. The Makefile is the single source of
   truth for configuration and passes every value explicitly to
   `docker compose`, which is invoked with `--env-file /dev/null` so it does
   NOT auto-load any `.env` file. The built-in defaults cover the happy path;
   to customize, either pass variables on the command line
   (`make TAG=11.0.0-3 VERTICA_DB_NAME=mydb docker-build-node`) or source
   `.env.example` (or a copy of it) into your shell before running make:

   ```shell
   source ./.env.example            # every line uses `export`, so plain source works
   # or, for local edits without polluting git:
   cp .env.example .env && source ./.env   # .env is gitignored
   make compose-up
   ```

## Build and run

Run each target from this directory.

| Step | Command | What it does |
| :--- | :------ | :----------- |
| 1 | `make docker-build-node` | Build the node image (includes the Vertica RPM at `/opt/vertica/install-rpm/` and pre-created catalog/data dirs for `$VERTICA_DB_NAME`). |
| 2 | `make compose-up` | Start `vertica1`, `vertica2`, `vertica3` on the fixed `172.28.0.0/24` subnet. |
| 3 | `make install-vertica` | Run `install_vertica` on `vertica1` as root against all three hosts. Uses the baked-in RPM. |
| 4 | `make create-db` | `admintools -t create_db` on `vertica1` as `dbadmin`. Creates the `dockerdb` database across all three nodes. |
| 5 | `make load-vmart` | Generate the VMart sample data (`vmart_gen`), load the schema, and run the ETL SQL. |
| 6 (optional) | `make test` | Run the 3-node smoke/cluster tests under [`tests/`](./tests/). Safe to rerun. |

Verify with:

```shell
make vsql                 # vsql -U dbadmin -d dockerdb on vertica1
# or
docker exec -u dbadmin -it vertica1 /opt/vertica/bin/admintools -t view_cluster
```

## Tests

`make test` runs every `tests/*.sql` file in alphabetical order against
`vertica1` as `dbadmin`, and tallies the results. Each test emits a single
line starting with `PASS`, `FAIL:<reason>` or `SKIP:<reason>`; anything else
is treated as a failure. The command exits non-zero if any test failed.

| # | File | What it checks | Needs |
| :- | :--- | :------------- | :---- |
| 01 | `01_cluster_up.sql` | All 3 nodes report `node_state = 'UP'` in `v_catalog.nodes`. | `create-db` |
| 02 | `02_ksafety.sql` | `designed_fault_tolerance = current_fault_tolerance = 1` (K=1). | `create-db` |
| 03 | `03_license.sql` | At least one row in `v_catalog.licenses` (proves `install_vertica --license CE` ran). | `install-vertica` |
| 04 | `04_flex_table_loaded.sql` | `FlexTableLib` UDx library is registered. Adapted from the upstream one-node-ce smoke test. | `create-db` |
| 05 | `05_vmart_loaded.sql` | `store_sales_fact`, `online_sales_fact`, `inventory_fact` have the expected row counts from `vmart_gen`. | `load-vmart` (else SKIP) |
| 06 | `06_data_distribution.sql` | `store_sales_fact` has rows on all 3 nodes (segmentation actually reached the whole cluster). | `load-vmart` (else SKIP) |
| 07 | `07_segmentation_balance.sql` | Coefficient of variation of `store_sales_fact` row counts across nodes is < 0.20 (hash segmentation is balanced). | `load-vmart` (else SKIP) |

Typical output after a full happy-path run:

```shell
$ make test
  PASS  01_cluster_up
  PASS  02_ksafety
  PASS  03_license
  PASS  04_flex_table_loaded
  PASS  05_vmart_loaded
  PASS  06_data_distribution
  PASS  07_segmentation_balance

Results: 7 passed, 0 skipped, 0 failed (of 7)
```

Tests 05–07 return `SKIP:vmart_not_loaded` if `make load-vmart` hasn't run
yet, so `make test` is also useful immediately after `make create-db`.

## What you get

- Containers: `vertica1`, `vertica2`, `vertica3`
- IPs (internal): `172.28.0.11`, `172.28.0.12`, `172.28.0.13`
- Host port mappings (override via `make VERTICA1_DB_PORT=... compose-up` or by
  exporting the variable from your shell — see `.env.example`):

  | Service | DB (5433) | SSH (22) |
  | :------ | :-------- | :------- |
  | vertica1 | `54331` | `10022` |
  | vertica2 | `54332` | `10023` |
  | vertica3 | `54333` | `10024` |

- Volumes: `vertica1_data`, `vertica2_data`, `vertica3_data` for `/data` on each
  node.

## Restarting

The entrypoint script auto-starts the database on `vertica1` (the leader) if
`admintools -t list_db` already reports `$VERTICA_DB_NAME` on this host
(i.e. `create_db` has run and `admintools.conf` is intact inside the
container filesystem). Followers wait for the leader. On `docker compose
stop` / container shutdown the leader gracefully runs `admintools -t stop_db`
via a `SIGTERM` trap.

**Safe restart commands**: `docker compose restart` and
`docker compose stop && docker compose start` both preserve the container
filesystem, so `/opt/vertica/config/admintools.conf` survives and the
leader's auto `start_db` picks the cluster back up.

**Not safe**: `docker compose down` (even without `-v`) removes the
containers. The next `up` recreates them from the image, so admintools
forgets about the cluster (its config lives in the image, not on the
`/data` volume). The DB files under `/data` are still there but become
orphaned — you need to re-run `make install-vertica` and `make create-db`.
If you need `compose down` semantics, use `make clean` to wipe `/data`
volumes too and start fresh.

## Reset

```shell
make clean    # docker compose down -v — removes all three volumes
```

## Customization

`Makefile` accepts the following overrides:

- `TAG` (default `latest`) — image tag.
- `IMAGE_NODE_NAME` (default `vertica-ce-node`).
- `VERTICA_PACKAGE` — autodetected from `packages/vertica*.rpm`; override to
  pin a specific RPM when multiple are present.
- `VERTICA_DB_UID` (default `1000`), `VERTICA_DB_GID` (default `1000`),
  `VERTICA_DB_USER` (default `dbadmin`), `VERTICA_DB_GROUP` (default
  `verticadba`), `VERTICA_DB_NAME` (default `dockerdb`) — always passed as
  `--build-arg` so the image pre-creates
  `/data/vertica/${VERTICA_DB_NAME}_catalog` and `_data` with the right
  owner. Changing any of these requires a rebuild.
- Runtime only: `VERTICA_HOSTS`, `VERTICA_LEADER`, `VERTICA_COMPOSE_CONTAINERS`,
  `VERTICA_DB_PASSWORD`.

Run `make display` to see the effective values.

## Troubleshooting

- `install_vertica` fails with "Failed to create or export SSH key on
  localhost" or similar PAM / session / TTY errors under `su - dbadmin`:
  usually a stale `~dbadmin/.ssh` from a previous half-run. `make clean`
  (wipes volumes) then re-run the happy path.
- `make compose-up` fails with "Pool overlaps with other one on this address
  space": another Docker network is already bound to `172.28.0.0/24`. Inspect
  with `docker network ls` and either remove the conflicting network or
  change the subnet in `docker-compose.yml`.
- Container is up but `make vsql` hangs: the leader's auto-`start_db` (from
  `docker-entrypoint.sh`) can take ~30–60 s after `compose-up` on cold boot.
  Tail `docker logs vertica1` for `Database dockerdb started successfully`.

This setup is a fork of
[vertica/vertica-containers/one-node-ce](https://github.com/vertica/vertica-containers/tree/main/one-node-ce);
refer to that repo for the single-node reference Dockerfile and entrypoint
behaviour we diverged from.
