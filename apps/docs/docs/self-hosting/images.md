# Image Contract

Images are published to Docker Hub under the namespace configured by:

```
RTYLR_IMAGE_REGISTRY=docker.io
RTYLR_IMAGE_NAMESPACE=rtylr
RTYLR_VERSION=1.0.0
```

A single `RTYLR_VERSION` pins ALL images (backend and frontends) to the same tag. Images are tagged by version — not just `latest`.

## Backend images (Onion runmodes)

| Image          | Port | Notes                                                      |
| -------------- | ---- | ---------------------------------------------------------- |
| `rtylr/api`    | 9999 | Validates `RTYLR_LICENSE_KEY` against the hosted authority |
| `rtylr/worker` | —    | Background worker; no HTTP                                  |
| `rtylr/upload` | 8080 | Media/CDN service; self-host analog of cdn.voxire.com      |

## Frontend images (static nginx on :8080)

`rtylr/auth`, `rtylr/dash`, `rtylr/pos`, `rtylr/erp`, `rtylr/hr`, `rtylr/crm`, `rtylr/finance`, `rtylr/flow`, `rtylr/insights`, `rtylr/recruit`, `rtylr/menu`.

Each frontend serves the compiled application on port `8080` and reads runtime configuration (`API_URL`, `UPLOAD_URL`, `AUTH_URL`, `DASH_URL`) from environment variables at container start, so customers can change URLs without rebuilding.

## Not part of the self-host image set

- There is **no `rtylr/license` image**. Customers do not run a license runmode. The backend `api` validates `RTYLR_LICENSE_KEY` against the hosted authority `https://license.voxire.com`. See [License Flow](/self-hosting/license).
- There is **no `rtylr/docs` image**. The documentation site is hosted by Voxire and not shipped for self-host.
- The marketing `landing` site is **not** shipped for self-host.
- The shared `packages` are build-time dependencies only, consumed by the app image build pipelines — never mounted into this runtime.
