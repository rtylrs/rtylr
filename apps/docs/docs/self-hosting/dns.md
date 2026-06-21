# DNS Setup

rtylr serves each app on its own subdomain of your base domain. With Compose the base domain is `RTYLR_DOMAIN`; with Helm it is `ingress.domain`.

## Subdomains

Create an A/AAAA (or CNAME) record for each of these, all pointing at the host/load balancer that fronts the stack:

| Subdomain  | Service        |
| ---------- | -------------- |
| `api`      | backend api    |
| `upload`   | upload / media |
| `auth`     | auth frontend  |
| `dash`     | dashboard      |
| `pos`      | point of sale  |
| `erp`      | erp            |
| `hr`       | hr             |
| `crm`      | crm            |
| `finance`  | finance        |
| `flow`     | flow           |
| `insights` | insights       |
| `recruit`  | recruit        |
| `menu`     | menu           |

For base domain `example.com`, that is `api.example.com`, `dash.example.com`, and so on. A wildcard `*.example.com` record also works and covers every app.

## Local development

With `RTYLR_DOMAIN=localhost`, `*.localhost` resolves to `127.0.0.1` on most systems with no DNS configuration. Use `http://dash.localhost`, `http://api.localhost`, etc.

## TLS

- **Compose**: the bundled `Caddyfile` is HTTP-first. Remove `auto_https off` for automatic Let's Encrypt certificates on a real domain, or terminate TLS in a proxy/load balancer in front of the stack.
- **Helm**: set `ingress.tls.enabled=true` and `ingress.tls.secretName`, and let cert-manager (or your ingress controller) provision certificates.
