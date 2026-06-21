# License Flow

Self-hosted rtylr requires a license key issued by Voxire sales.

## How it works

1. You receive a license key from Voxire sales.
2. You set it as `RTYLR_LICENSE_KEY` (in `.env` for Compose, or `license.key` / `license.existingSecret` for Helm).
3. The backend `api` service validates the key against the hosted authority `https://license.voxire.com`.

```
api  --(validate RTYLR_LICENSE_KEY)-->  https://license.voxire.com
```

## Important

- The authority URL `https://license.voxire.com` is **hardcoded in the backend** and is **not** customer-configurable. Do not look for a `LICENSE_AUTHORITY_URL` setting — it does not exist.
- There is **no local license service** and **no `rtylr/license` image**. Customers do not run a license runmode.
- `https://license.voxire.com` is the **only** external dependency of the entire self-host stack. Everything else (MySQL, Redis, object storage) can run locally or on your own infrastructure.

## Network requirements

The host running the `api` service must be able to reach `https://license.voxire.com` over HTTPS (port 443). If your environment uses an egress proxy or firewall, allowlist that host.

## Setting the key

Compose (`.env`):

```bash
RTYLR_LICENSE_KEY=your-key-here
DEPLOYMENT_MODE=self_hosted
```

Helm:

```bash
helm upgrade --install rtylr ./helm/rtylr \
  --set-string license.key=your-key-here
```

Or reference an existing Kubernetes Secret with a `RTYLR_LICENSE_KEY` key:

```bash
helm upgrade --install rtylr ./helm/rtylr \
  --set license.existingSecret=rtylr-license
```
