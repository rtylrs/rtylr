# Running a real backend without Docker

`self-host/docker-compose.yml` is the normal way to get a local backend, and it
is the right way. This page is for environments with **no container runtime** —
CI sandboxes, agent environments, a laptop without Docker — where the Go API
still needs a real MySQL to be exercised against.

Everything below was used to run `e2e/auth-real-login.spec.ts` and
`e2e/app-shells-real.spec.ts` against real data.

## 1. MySQL 8

Grab the official static build and run it from your own home directory:

```sh
curl -L -o mysql8.tar.xz \
  https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.40-linux-glibc2.17-aarch64.tar.xz
tar -xf mysql8.tar.xz
export MYSQL_HOME="$PWD/mysql-8.0.40-linux-glibc2.17-aarch64"
export PATH="$MYSQL_HOME/bin:$PATH"
```

`mysqld` and the `mysql` client need `libaio` and (on older glibc) `libncurses5`.
If `mysqld` fails with `error while loading shared libraries: libaio.so.1`, drop
the Debian packages' shared objects into one directory and point
`LD_LIBRARY_PATH` at it:

```sh
mkdir -p /tmp/mysqllibs/lib
# libaio1, libnuma1, libncurses5, libtinfo5 for your arch, then:
#   dpkg-deb -x <pkg>.deb ./root && cp root/usr/lib/<triple>/lib*.so* /tmp/mysqllibs/lib/
export LD_LIBRARY_PATH=/tmp/mysqllibs/lib:$LD_LIBRARY_PATH
```

Initialise and start it:

```sh
mkdir -p /tmp/mysqldata /tmp/mysqlrun
mysqld --initialize-insecure --user="$(whoami)" --basedir="$MYSQL_HOME" --datadir=/tmp/mysqldata
mysqld --user="$(whoami)" --basedir="$MYSQL_HOME" --datadir=/tmp/mysqldata \
       --port=3306 --socket=/tmp/mysqlrun/mysql.sock --mysqlx=OFF \
       --log-error=/tmp/mysqlrun/mysqld.err &
```

### Load the time-zone tables (do not skip this)

The API's DSN sets `time_zone`, and a freshly initialised datadir has no zone
tables, so every connection fails with `Error 1298 (HY000): Unknown or incorrect
time zone`. Load them once:

```sh
mysql_tzinfo_to_sql /usr/share/zoneinfo | mysql --socket=/tmp/mysqlrun/mysql.sock -uroot mysql
```

## 2. Database, migrations, seeds

```sh
mysql --socket=/tmp/mysqlrun/mysql.sock -uroot <<'SQL'
CREATE DATABASE IF NOT EXISTS onion_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'onion_user'@'localhost' IDENTIFIED BY 'onion_password';
GRANT ALL PRIVILEGES ON onion_dev.* TO 'onion_user'@'localhost';
SQL

cd services/onion
DB_HOST=127.0.0.1 DB_PORT=3306 DB_NAME=onion_dev DB_USER=onion_user DB_PASSWORD=onion_password \
  bash scripts/migrate.sh

for f in scripts/fixtures/*.sql; do
  [ "$(basename "$f")" = "70_direct_authorization.sql" ] && continue   # evidence-gated
  MYSQL_PWD=onion_password mysql --default-character-set=utf8mb4 \
    -h127.0.0.1 -P3306 -uonion_user onion_dev < "$f"
done
```

## 3. Users the specs need

`scripts/fixtures/*.sql` seeds businesses and users, but **not** the
`authorization_*` rows the resolver reads, and its password hashes are
placeholders. Create a usable pair — one user in each of two tenants:

```sql
-- memberships the resolver requires (the fixtures leave these empty)
INSERT INTO authorization_business_members
  (business_id, user_id, dash_role_id, dash_role_key, status, authorization_version)
VALUES (1, 1001, 1, 'dash.owner', 'active', 1)
ON DUPLICATE KEY UPDATE status = 'active';

-- per-app role + an active license, or module routes 403
INSERT INTO authorization_role_assignments
  (business_id, user_id, app_id, role_id, source, status, is_primary, granted_by)
SELECT 1, 1001, r.app_id, r.id, 'explicit', 'active', 1, 1001
FROM authorization_roles r
WHERE r.role_key IN ('crm.admin', 'flow.admin', 'hr.admin')
  AND NOT EXISTS (SELECT 1 FROM authorization_role_assignments a
                  WHERE a.user_id = 1001 AND a.app_id = r.app_id);

INSERT INTO authorization_license_assignments
  (business_id, user_id, app_id, source, status, granted_by)
SELECT 1, 1001, t.app_id, 'explicit', 'active', 1001
FROM (SELECT 'crm' app_id UNION ALL SELECT 'flow' UNION ALL SELECT 'hr'
      UNION ALL SELECT 'dash') t
WHERE NOT EXISTS (SELECT 1 FROM authorization_license_assignments l
                  WHERE l.user_id = 1001 AND l.app_id = t.app_id);
```

For real logins the seeded `password_hash` values are not valid bcrypt, so set
a real one using the repo's own hasher (cost 12) — never hardcode a hash in a
committed fixture:

```go
// throwaway, outside the repo
passwordutils.HashPassword("TestPass!123")
```

### A limited (viewer) user, to exercise the deny paths

An owner holds every permission, so an owner-only run never exercises a denial.
Add a viewer in the same business with read-only roles and a deliberately
narrow license set, so the permission boundary is actually tested:

```sql
INSERT INTO users (business_id, email, password_hash, full_name, onboarded, is_console_admin, auth_provider)
VALUES (1, 'viewer@example.test', '<bcrypt>', 'Dummy Viewer', 1, 0, 'password')
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash);

-- a member, not an owner
INSERT INTO authorization_business_members
  (business_id, user_id, dash_role_id, dash_role_key, status, authorization_version)
VALUES (1, (SELECT id FROM users WHERE email='viewer@example.test'), 4, 'dash.member', 'active', 1)
ON DUPLICATE KEY UPDATE status = 'active';

-- read-only roles: crm.viewer has crm:customers:view but no :create/:edit/:delete
INSERT INTO authorization_role_assignments
  (business_id, user_id, app_id, role_id, source, status, is_primary, granted_by)
SELECT 1, u.id, r.app_id, r.id, 'explicit', 'active', 1, 1001
FROM users u JOIN authorization_roles r ON r.role_key IN ('crm.viewer','hr.viewer','flow.viewer')
WHERE u.email = 'viewer@example.test'
  AND NOT EXISTS (SELECT 1 FROM authorization_role_assignments a
                  WHERE a.user_id = u.id AND a.app_id = r.app_id);

-- licensed for crm/hr/flow/dash only; erp, finance, pos, insights, menu and
-- recruit are left unlicensed on purpose so those routes must be denied
INSERT INTO authorization_license_assignments
  (business_id, user_id, app_id, source, status, granted_by)
SELECT 1, u.id, t.app_id, 'explicit', 'active', 1001
FROM users u JOIN (SELECT 'crm' app_id UNION ALL SELECT 'hr' UNION ALL SELECT 'flow'
                   UNION ALL SELECT 'dash') t
WHERE u.email = 'viewer@example.test'
  AND NOT EXISTS (SELECT 1 FROM authorization_license_assignments l
                  WHERE l.user_id = u.id AND l.app_id = t.app_id);

-- REQUIRED for a non-owner: the app list the app-access gate is evaluated
-- against. `GetMeByBusiness` -> `GetPermissionByBusiness` reads exactly this
-- row (SELECT ... FROM user_app_permissions WHERE business_id = ? AND user_id
-- = ?), and a missing row yields an empty app list, so the member is denied
-- every app route even while holding a valid role and license. `role` is an
-- enum here, so use one of owner/admin/operator/billing/menu_manager/
-- store_manager/staff/viewer -- 'member' is rejected by the column.
INSERT INTO user_app_permissions (business_id, owner_id, user_id, apps, role, is_active)
VALUES (1, 1001, (SELECT id FROM users WHERE email='viewer@example.test'),
        '["dash","crm","hr","flow"]', 'staff', 1)
ON DUPLICATE KEY UPDATE apps = VALUES(apps), role = VALUES(role), is_active = 1;
```

Two things to know about the owner/member split:

- **A member's** readable apps come from its own `user_app_permissions` row
  (`GetPermissionByBusiness`). The fixtures create that table but do not give
  the seeded users rows, so a member resolves to an empty app list and is denied
  every app route. Seed the row above.
- **An owner's** apps come from a different path: `GetOwnerApps` joins
  `subscriptions` to `plans` for the owner, and falls back to
  `defaultFreePlanApps` when there is no subscription. Owners are therefore
  unaffected by the missing `user_app_permissions` rows, which is why an
  owner-only run hides this entirely.

Restart the API after changing authorization rows: `EffectiveAccess` is resolved
per request but the surrounding profile is cached in-process, so a running
server keeps serving the pre-change access until it restarts.

## 4. Run the API

A `.env` in the working directory is what `cmd/api` reads:

```
RUN_MODE=api
PORT=9999
PROBE_PORT=8888
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=onion_dev
DB_USER=onion_user
DB_PASSWORD=onion_password
REDIS_ENABLED=false
JWT_SECRET=local_dev_secret_key_for_rtylr_testing_only_32chars
DEPLOYMENT_MODE=hosted
```

```sh
go build -o /tmp/onion-api ./cmd/api
cd /tmp/onion-run && /tmp/onion-api
```

`GET :8888/health` should return 200 and `GET :9999/` should return 401 (auth
enforced). Redis off means local cache only; object storage, media-engine and
email stay unconfigured and are reported as degraded, which is fine for auth and
tenant-boundary work.

## What this does not give you

- No container parity: the self-host `license` runmode, floci/S3, and
  media-engine are not exercised.
- No Redis, so cache behaviour and realtime are out of scope.
- The login route is rate limited to 5/min per IP; back off between manual runs
  or the real specs will 429.
