# Rtylr self-host starter

This compose stack keeps hosted and self-hosted behavior separate while using
the same Onion binaries:

- `api` serves business/auth/POS/ERP API routes.
- `upload` is the only runmode that accepts file uploads and serves uploaded media.
- `worker` runs background jobs.
- `license` exposes self-host license status and verification.
- `db`, `redis`, and `floci` provide local MySQL, cache, and S3-compatible storage.

## Run it

```sh
cp self-host/.env.example self-host/.env
docker compose --env-file self-host/.env -f self-host/docker-compose.yml up --build
```

Before exposing a self-host install, replace `JWT_SECRET`, `CDN_API_KEY`, and
the license values in `self-host/.env`. Self-host keys are issued by the
Voxire sales/admin team and are not generated inside a customer install.

## Ports

- API: `http://localhost:9999`
- Upload/CDN: `http://localhost:8081`
- License: internal-only container endpoint on port `8082`
- Floci object storage: `http://localhost:4566`
- MySQL: `localhost:3306`
- Redis: `localhost:6379`

## License status

Hosted deployments default to `DEPLOYMENT_MODE=hosted` and do not require a
license key. Self-hosted images must be built by the vendor with the vendor
public key embedded in the binary, then receive only `LICENSE_KEY` or
`LICENSE_FILE` at runtime. `LICENSE_PUBLIC_KEY` is intentionally unsupported.

The license runmode exposes:

- `GET /license/status`
- `POST /license/verify`
- `POST /license/activate`
- `POST /license/refresh`

All endpoints return a JSON status report with `valid`, `status`, plan, feature,
seat, and expiry fields. `POST /license/verify` can receive a body like:

```json
{ "license_key": "rtylr_license_v1....", "instance_id": "optional-local-id" }
```
