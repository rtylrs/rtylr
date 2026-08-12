# Helm Install (primary path)

The Helm chart lives at `helm/rtylr/`. It deploys the backend (`api`, `worker`, `upload`) and all 11 frontends as Deployments + Services behind a single Ingress, plus a ConfigMap/Secret for environment.

## Prerequisites

- A Kubernetes cluster and `kubectl` context
- Helm 3.8+ (or Helm 4)
- An ingress controller (e.g. ingress-nginx) and DNS per [DNS Setup](/self-hosting/dns)
- A license key from Voxire

## Minimal install (bring your own datastores)

```bash
helm upgrade --install rtylr ./helm/rtylr \
  --set image.tag=1.0.0 \
  --set-string license.key=YOUR_LICENSE_KEY \
  --set-string secrets.jwtSecret=$(openssl rand -hex 32) \
  --set ingress.domain=example.com \
  --set database.host=mysql.internal --set-string database.password=... \
  --set redis.url=redis://redis.internal:6379 \
  --set s3.endpoint=https://s3.amazonaws.com \
  --set s3.bucket=rtylr --set s3.region=eu-west-1 \
  --set-string s3.accessKeyId=... --set-string s3.secretAccessKey=...
```

## Self-contained install (bundled datastores, evaluation)

```bash
helm upgrade --install rtylr ./helm/rtylr \
  --set image.tag=1.0.0 \
  --set-string license.key=YOUR_LICENSE_KEY \
  --set-string secrets.jwtSecret=$(openssl rand -hex 32) \
  --set ingress.domain=example.com \
  --set localInfra.mysql.enabled=true --set-string localInfra.mysql.rootPassword=$(openssl rand -hex 16) \
  --set localInfra.redis.enabled=true \
  --set localInfra.floci.enabled=true \
  --set-string database.password=$(openssl rand -hex 16) \
  --set-string s3.secretAccessKey=$(openssl rand -hex 16)
```

The bundled datastores are single-replica and intended for evaluation. Use a managed/HA database, cache, and object store for production.

## Key values

| Value                                                 | Purpose                                          |
| ----------------------------------------------------- | ------------------------------------------------ |
| `image.registry` / `image.namespace` / `image.tag`   | Image source + global version                    |
| `license.key` / `license.existingSecret`              | Your rtylr license key (provided by Voxire) |
| `deploymentMode`                                      | `self_hosted`                                    |
| `urls.*`                                              | Configurable app URLs                            |
| `secrets.jwtSecret` / `secrets.existingSecret`        | JWT secret (required)                            |
| `database.*` / `redis.url` / `s3.*`                   | External datastore config                        |
| `extraEnv`                                            | Plain passthrough env applied to backend (ConfigMap) |
| `extraSecretEnv`                                      | Secret passthrough env applied to backend (Secret) |
| `backend.<svc>.replicas` / `frontends.<app>.replicas` | Replica counts                                   |
| `frontends.<app>.enabled`                             | Toggle individual frontends                      |
| `ingress.*`                                           | Ingress class, domain, annotations, TLS          |
| `localInfra.{mysql,redis,floci}.enabled`              | Deploy bundled datastores                        |

### Passthrough env example

```yaml
extraEnv:
  AXIOM_DATASET: rtylr-prod
  AI_PRIMARY_PROVIDER: openai
extraSecretEnv:
  OPENAI_API_KEY: sk-...
  RESEND_API_KEY: re_...
```

## Verify the chart

```bash
helm lint helm/rtylr --set secrets.jwtSecret=$(openssl rand -hex 32)
helm template rtylr helm/rtylr --set secrets.jwtSecret=$(openssl rand -hex 32) | less
```

## Upgrade

Bump `image.tag` (a single value pins all images) and re-run `helm upgrade --install`. See [Upgrade Flow](/self-hosting/upgrade).
