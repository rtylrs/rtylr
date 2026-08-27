# Role and Access Control V2 migration runbook

This runbook is for a disposable or production-style MySQL database. Never put a
real password, token, or connection string in this document or in shell history.

## Run

Set `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME` in the
environment, then run:

```sh
cd /Users/abedamouneh/Development/onion
bash scripts/migrate.sh
```

The runner records each successful file in `schema_migrations`, prints MySQL
affected-row output and warnings, stops on the first failed file, and does not
record a failed file as applied. The cross-schema migrations are deliberately
deferred until their owning schemas exist:

- `hr_time/zz_flow_client_key_tenant_scope.sql`
- `shared/zz_role_access_v2_integrity.sql`

The six Role and Access Control V2 files therefore execute in this order within
the complete run:

1. `shared/zz_role_access_v2.sql`
2. `shared/zz_role_access_v2_backfill.sql`
3. `shared/zz_role_access_v2_email_verification.sql`
4. `shared/zz_role_access_v2_provisioning.sql`
5. `shared/zz_role_access_v2_usernames.sql`
6. `shared/zz_role_access_v2_integrity.sql`

The integrity file is last so its foreign keys, indexes, ownership checks, and
fail-closed preflight checks can see POS locations and Flow tables.

## Repeatability and recovery

Run the command a second time against the same database. A healthy rerun must
report zero applied, all previously recorded files skipped, and zero failures.
To verify a restored snapshot, restore the dump into a new database name, point
the same five variables at that database, and run the same command. Do not use
the production database as the disposable target.

The operational rollback switch is:

```sh
AUTHORIZATION_V2_MODE=legacy
```

Set it only through deployment configuration and restart the API. The default
is normalized V2 authorization. Legacy mode preserves the old licensed app-level
boundary for recovery while retaining normalized V2 data. Remove the variable or
set it to a value other than `legacy` to return to V2 mode after the incident.

If a migration fails, retain the database and the complete command output. Do
not mark the file applied manually. Restore the disposable database from its
pre-migration dump, correct the failing migration, and rerun. Production
recovery requires the normal database backup and change-control procedure.

## Release-gate evidence recorded on 2026-08-26

The disposable fresh database `rtylr_gate_runner_release_20260826` completed
with `Applied: 170, Skipped: 0, Failed: 0`. Post-migration counts were:

| Check | Count |
| --- | ---: |
| `schema_migrations` | 170 |
| permissions | 104 |
| roles | 37 |
| role_permissions | 393 |
| members | 0 |
| licenses | 0 |
| assignments | 0 |
| audit_events | 0 |

The exact six-file RBAC ordering completed with `Applied: 170, Skipped: 0,
Failed: 0` and no numeric warnings. A second run completed with
`Applied: 0, Skipped: 170, Failed: 0`. A fresh dump restored into a separate
database also completed with `Applied: 0, Skipped: 170, Failed: 0`, retaining
the expected 170 migration rows, 104 permissions, 37 roles, and 393 role
permissions.

Disposable preflight tests rejected orphaned, duplicate-active, invalid-target,
invalid-scope, and cross-business records. The synthetic invalid rows were
inserted only after disabling the exact constraint needed for each test, then
`zz_role_access_v2_integrity.sql` aborted with
`role_access_v2_integrity_abort`. The migration did not silently repair any of
those rows.

The full repository DB fixture harness is a separate gate. Its migration and
seeding phase passes, but the current `make test-db` run still exposes unrelated
fixture and schema drift in legacy integration packages. Treat those failures as
release blockers until the fixture suite is repaired; do not convert them into
migration success.
