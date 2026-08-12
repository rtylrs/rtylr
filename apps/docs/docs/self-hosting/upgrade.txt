# Upgrade Flow

All rtylr images share a single version. Bump one value to upgrade the whole stack, Retool-style.

## The version

`RTYLR_VERSION` (e.g. `1.0.0`) pins every image — backend and frontends — to the same Docker tag. It lives in `versions.env` (and `.env` for Compose); the Helm equivalent is `image.tag`.

## Compose

```bash
./scripts/update.sh 3.300.14
```

This writes `RTYLR_VERSION=3.300.14` into `versions.env` and `.env`, then pulls the new images and restarts. Run without an argument to pull/restart the current pinned version. The script auto-detects whether the `local-infra` profile is in use.

Always back up first:

```bash
./scripts/backup.sh
```

## Helm

```bash
helm upgrade --install rtylr ./helm/rtylr \
  --reuse-values \
  --set image.tag=3.300.14
```

Or update `image.tag` in your values file and run `helm upgrade`.

## Rollback

- **Helm**: `helm rollback rtylr` (or to a specific revision).
- **Compose**: set `RTYLR_VERSION` back to the previous tag and run `./scripts/update.sh <previous>`.

## Notes

- Pin a specific version in production. Avoid `latest`.
- Database migrations (if any) are run by the `api`/`worker` images on start; take a backup before upgrading.
