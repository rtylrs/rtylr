# rtylr Self-Host Architecture

This repository is an installer and runtime orchestrator. It does not contain product source code. The source repositories build public Docker images, and this repo wires those images together for a customer-owned deployment via two supported install paths: a **Helm chart** (primary) and **Docker Compose**.

## Runtime Shape

```
browser
  |
  v
ingress / caddy proxy
  |-- auth | dash | pos | erp | hr | crm | finance      (frontends, :8080)
  |   flow | insights | recruit | menu
  |-- api:9999
  |-- upload:8080

api / worker / upload
  |-- MySQL                 (bundled local OR bring-your-own)
  |-- Redis                 (bundled local OR bring-your-own)
  |-- floci object storage  (bundled local OR bring-your-own S3-compatible)
  |
  +--> https://license.voxire.com   (ONLY external dependency)
```

The upload service is intentionally separate from the API service. Uploads route to the upload service (the self-host analog of `cdn.voxire.com`, reached via the configurable `UPLOAD_URL`); the API only stores and reads media metadata.

## Self-Sustaining Stack

The stack is self-contained by default. A customer may bring their own MySQL, Redis, and S3-compatible object storage by pointing `DB_HOST`, `REDIS_URL`, and `S3_*` at their services. If they do not, the stack spins up local MySQL, Redis, and **floci** (local S3-compatible object storage, bucket created by `floci/init/create-bucket.sh`).

- **Compose**: local infra is gated behind the `local-infra` profile.
- **Helm**: local infra is gated behind `localInfra.{mysql,redis,floci}.enabled`.

The only thing the stack ever reaches on the public internet is the hosted license authority `https://license.voxire.com`.

## License Flow

Self-hosted customers receive a license key from Voxire sales and set it as `RTYLR_LICENSE_KEY`. The backend `api` validates that key against the hosted authority `https://license.voxire.com`. That authority URL is hardcoded in the backend and is **not** customer-configurable. There is **no local license service** in the self-host stack. See [License Flow](/self-hosting/license) for details.

## Configuration

Every app URL is configurable (`AUTH_URL`, `DASH_URL`, `POS_URL`, `ERP_URL`, `HR_URL`, `CRM_URL`, `FINANCE_URL`, `FLOW_URL`, `INSIGHTS_URL`, `RECRUIT_URL`, `MENU_URL`, `API_URL`, `UPLOAD_URL`). The backend derives `CORS_ALLOWED_ORIGINS` from the configured frontend URLs. No hosted domains are hardcoded in any image.

## Further Reading

- [License Flow](/self-hosting/license) — license validation flow
- [DNS Setup](/self-hosting/dns) — DNS records to create
- [Helm Install](/self-hosting/helm) — Helm install
- [Docker Compose](/self-hosting/compose) — Compose reference
- [Backend Environment](/self-hosting/backend-env) — backend environment reference
- [Image Contract](/self-hosting/images) — image list and contract
- [Upgrade Flow](/self-hosting/upgrade) — upgrade flow
