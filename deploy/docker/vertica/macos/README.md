# Vertica 3-node cluster on macOS

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](../../../../LICENSE)

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

3. An SSH keypair at `.ssh/id_rsa` and `.ssh/id_rsa.pub`.
   The build copies them into `/root/.ssh/` inside the image and seeds
   `/root/.ssh/authorized_keys` for passwordless root SSH between nodes.

   Either reuse your personal keypair from `~/.ssh`:
   ```shell
   mkdir -p .ssh
   cp ~/.ssh/id_rsa     .ssh/id_rsa
   cp ~/.ssh/id_rsa.pub .ssh/id_rsa.pub
   chmod 0600 .ssh/id_rsa
   chmod 0644 .ssh/id_rsa.pub
   ```

   Or generate a dedicated one for this cluster (recommended if your personal
   key is passphrase-protected — `install_vertica` needs a key without a
   passphrase):

   ```shell
   mkdir -p .ssh
   ssh-keygen -t rsa -b 4096 -N '' -C "vertica-macos-cluster" -f .ssh/id_rsa
   ```

## Customization

The Makefile is the single source of truth for every variable and passes
them explicitly to `docker compose`, which is invoked with
`--env-file /dev/null` so it does NOT auto-load any `.env` file. Override a
default in one of two ways:

- pass it on the command line:
  `make TAG=11.0.0-3 VERTICA_DB_NAME=mydb docker-build-node`, or
- export it in your shell before running `make`, e.g. by sourcing
  [`.env.example`](./.env.example) (which ships with `export KEY=value`
  lines).

Run `make display` to see the effective values.

### Image properties

Build args baked into the image. Changing any of these requires a rebuild
(`make docker-build-node`).

| Environment Variable | Description | Default Value |
| :------------------- | :---------- | :------------ |
| `TAG` | Image tag. Also narrows the `VERTICA_PACKAGE` auto-detect glob when set to a specific version. | `latest` |
| `IMAGE_NODE_NAME` | Image name; the full reference is `$(IMAGE_NODE_NAME):$(TAG)`. | `vertica-ce-node` |
| `OS_IMAGE` | Base OS image. | `almalinux` |
| `OS_VERSION` | Base OS version. | `8.10` |
| `VERTICA_PACKAGE` | RPM filename under `packages/`. Autodetected from `packages/vertica*.rpm` (or `packages/vertica-$(TAG)*.rpm` when `TAG` is a version). Override to pin a specific RPM when multiple are present. | autodetected |

### Database user and name

Build args that also flow through at runtime to `install-vertica`,
`create-db`, and `load-vmart`. Changing them requires a rebuild so that the
pre-created catalog/data dirs (`/data/vertica/${VERTICA_DB_NAME}_{catalog,data}`)
match what `admintools` expects.

> **Note**: if a cluster is already running when you change any of these,
> you also need `make clean` to drop the `/data` volumes. Otherwise the
> persisted `admintools.conf`, catalog/data dirs and `dbadmin` SSH keys will
> still reference the previous values and startup will fail.

| Environment Variable | Description | Default Value |
| :------------------- | :---------- | :------------ |
| `VERTICA_DB_USER` | OS user and implicit database [superuser](https://www.vertica.com/docs/latest/HTML/Content/Authoring/AdministratorsGuide/DBUsersAndPrivileges/Privileges/AboutSuperuserPrivileges.htm). | `dbadmin` |
| `VERTICA_DB_UID` | UID for `VERTICA_DB_USER`. | `1000` |
| `VERTICA_DB_GROUP` | Group for database administrator users. | `verticadba` |
| `VERTICA_DB_GID` | GID for `VERTICA_DB_GROUP`. | `1000` |
| `VERTICA_DB_NAME` | Vertica database name. | `dockerdb` |

### Runtime only

No rebuild required; override on the Make command line or export before
running `make`.

| Environment Variable | Description | Default Value |
| :------------------- | :---------- | :------------ |
| `VERTICA_DB_PASSWORD` | Password for `VERTICA_DB_USER`, passed once to `admintools -t create_db -p`. Empty means no password. | _(empty)_ |
| `VERTICA1_DB_PORT` / `VERTICA1_SSH_PORT` | Host ports mapped to `vertica1:5433` / `vertica1:22`. | `54331` / `10022` |
| `VERTICA2_DB_PORT` / `VERTICA2_SSH_PORT` | Host ports mapped to `vertica2:5433` / `vertica2:22`. | `54332` / `10023` |
| `VERTICA3_DB_PORT` / `VERTICA3_SSH_PORT` | Host ports mapped to `vertica3:5433` / `vertica3:22`. | `54333` / `10024` |

### Fixed in `docker-compose.yml`

Cluster topology is not overridable from the Make command line — it is
hardcoded in [`docker-compose.yml`](./docker-compose.yml) to keep the
`install_vertica` step reproducible. To change any of the following, edit
`docker-compose.yml` directly (and keep the matching defaults in `Makefile`
aligned) and then `make clean && make compose-up && make install-vertica`:

- Container names (`vertica1`, `vertica2`, `vertica3`) — pinned via
  `container_name` on each service.
- Per-node IPs (`172.28.0.11/12/13`) and subnet (`172.28.0.0/24`) — pinned
  via `networks.vertica_cluster.ipv4_address` and `networks.*.ipam`.
- Which node auto-starts the database on `compose up` — pinned via
  `VERTICA_CLUSTER_ROLE: leader` on `vertica1` (all others are followers).
- Number of nodes — add/remove services in `docker-compose.yml` and update
  `VERTICA_HOSTS` in `Makefile` to match.

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
| 04 | `04_flex_table_loaded.sql` | End-to-end flex-table smoke test: creates a flex table, ingests JSON via `FJSONPARSER`, runs `COMPUTE_FLEXTABLE_KEYS`, and verifies extracted values via `MapLookup`. Implicitly covers that `FlexTableLib` is registered on every node. | `create-db` |
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

## Stop and restart

Once the cluster has been created, cycling it is a three-step loop:

```shell
make compose-down     # stops + removes containers; /data volumes survive
make compose-up       # recreates containers from the image
make test             # optional: confirm the cluster came back up (~30–60 s warm-up)
```

The cluster survives this cycle because the entrypoint redirects
`/opt/vertica/config` and `/home/dbadmin` onto the `/data` named volume on
first boot (see `preserve_config` / `preserve_dbadmin_home` in
`docker-entrypoint.sh`). On restart the leader runs
`admintools -t list_db -d $VERTICA_DB_NAME`, sees that the DB is known,
and auto-starts it with `admintools -t start_db`. Followers wait for the
leader. On shutdown the leader traps `SIGTERM` and runs `stop_db`.

Use `make logs` to tail the leader and watch for
`Database ${VERTICA_DB_NAME}: Startup Succeeded. All Nodes are UP`.

## Reset

```shell
make clean    # docker compose down -v — also removes the /data volumes
```

Unlike `compose-down`, this wipes `admintools.conf`, the DB files, and
`dbadmin`'s SSH keys, so you need the full happy path
(`install-vertica` → `create-db` → `load-vmart`) again.

---

## Credits

This setup is a fork of
[**vertica/vertica-containers/one-node-ce**](https://github.com/vertica/vertica-containers/tree/main/one-node-ce) —
refer to that repo for the single-node reference Dockerfile and entrypoint
behaviour we diverged from.
