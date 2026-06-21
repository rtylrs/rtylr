# License Flow

Self-hosted rtylr requires a license key provided by Voxire.

## How it works

1. You receive a license key from Voxire.
2. You set it as `RTYLR_LICENSE_KEY` (in `.env` for Compose, or `license.key` / `license.existingSecret` for Helm).
3. The platform validates your license automatically.

## Important

- There is **no separate license service** and **no `rtylr/license` image** to run. Licensing is handled by the backend automatically.
- MySQL, Redis, and object storage can run locally or on your own infrastructure.

## Network requirements

Ensure the backend is allowed to make outbound HTTPS requests (port 443). If your environment uses an egress proxy or firewall, allow outbound HTTPS from the backend.

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
