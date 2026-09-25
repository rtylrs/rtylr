# Docker Compose Reference

Docker Compose is a supported install path (the Helm chart is the primary one). The stack is defined in `docker-compose.yaml`.

## Prerequisites

- Docker Engine / Docker Desktop with Compose v2
- A license key from Voxire

## Setup

```bash
./scripts/setup.sh
```

This copies `.env.example` to `.env`, generates secrets (`JWT_SECRET`, `INTERNAL_TOKEN`, `CDN_API_KEY`, MySQL/S3 credentials), and prompts for `RTYLR_LICENSE_KEY`.

## Start

Self-contained (bundles MySQL, Redis, floci via the `local-infra` profile):

```bash
docker compose --profile local-infra --env-file .env up -d
```

Bring your own MySQL/Redis/S3 (set `DB_HOST`, `REDIS_URL`, `S3_*` in `.env`, then omit the profile):

```bash
docker compose --env-file .env up -d
```

## Profiles

| Profile       | Services added                               |
| ------------- | -------------------------------------------- |
| (none)        | caddy, api, worker, upload, all 11 frontends |
| `local-infra` | + db (MySQL), redis, floci (S3), floci-init  |

When the `local-infra` profile is off, the backend points at whatever you set in `DB_HOST` / `REDIS_URL` / `S3_*`.

## Services and ports

| Service                                                      | Image                          | Internal port        |
| ------------------------------------------------------------ | ------------------------------ | -------------------- |
| caddy                                                        | caddy:2-alpine                 | 80 / 443 (published) |
| api                                                          | rtylr/api                      | 9999                 |
| worker                                                       | rtylr/worker                   | —                    |
| upload                                                       | rtylr/upload                   | 8080                 |
| auth/dash/pos/erp/hr/crm/finance/flow/insights/recruit/menu  | rtylr/&lt;app&gt;              | 8080                 |
| db                                                           | mysql:8.0 (local-infra)        | 3306                 |
| redis                                                        | redis:7-alpine (local-infra)   | 6379                 |
| floci                                                        | localstack/localstack:3 (local-infra) | 4566         |

Routing is handled by Caddy off `RTYLR_DOMAIN` — see `Caddyfile` and [DNS Setup](/self-hosting/dns).

## Common operations

```bash
./scripts/doctor.sh          # validate environment + compose config
./scripts/update.sh 3.300.14 # bump RTYLR_VERSION, pull, restart
./scripts/backup.sh          # dump MySQL + archive floci storage
docker compose --env-file .env logs -f api
docker compose --profile local-infra --env-file .env down
```

## Validate the compose file

```bash
docker compose --env-file .env config >/dev/null
docker compose --profile local-infra --env-file .env config >/dev/null
```
