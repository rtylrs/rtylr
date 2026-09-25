# Self-host rtylr

Run rtylr on infrastructure you control. The self-hosting guides cover the deployment architecture, image contract, Helm and Docker Compose installs, DNS, backend configuration, and upgrades.

## Choose a deployment path

- **Helm** — the primary path for Kubernetes clusters and production environments. Start with [Helm Install](/self-hosting/helm).
- **Docker Compose** — a practical path for evaluation, smaller installations, and teams that prefer a single host. Start with [Docker Compose](/self-hosting/compose).

## Before you begin

Review the [architecture](/self-hosting/architecture), prepare the image and environment contract, and make sure your DNS and TLS plan is ready. For production, use managed or highly available datastores and keep credentials in a secret manager.

## Keep your deployment healthy

After installation, use the [DNS setup](/self-hosting/dns), [backend environment](/self-hosting/backend-env), and [upgrade flow](/self-hosting/upgrade) references to operate the installation safely.
