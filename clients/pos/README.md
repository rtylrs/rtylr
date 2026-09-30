# rtylr POS

Tauri + React POS terminal for rtylr.

## Setup

```bash
pnpm install
```

Copy `.env.example` to `.env` for local overrides.

## Verification

```bash
pnpm typecheck
pnpm test
pnpm build
pnpm native:check
pnpm tauri:build
```

`pnpm tauri:build` produces the native app bundle under `core/target/release/bundle/`.

## Desktop Releases

Pushing a tag named `pos-v<version>` or running the `Desktop Release` workflow manually builds downloadable macOS, Windows, and Linux installers from the private `tryrtylr/pos` source repo and deploys them to `https://pos-downloads.rtylr.com`.

The workflow also generates the Tauri updater `latest.json` release asset. The desktop app checks that file on startup and installs any newer signed release. Large installer binaries are published to `https://pos-download-assets.rtylr.com` from the private CI job, so GitHub source archives are never the customer-facing download surface. CI needs these secrets:

- `TAURI_SIGNING_PRIVATE_KEY`
- `TAURI_SIGNING_PRIVATE_KEY_PASSWORD` (optional; leave empty for the current generated key)
- `CLOUDFLARE_ACCOUNT_ID`
- `CLOUDFLARE_API_TOKEN`

The public updater key is stored in `core/tauri.conf.json`; keep the private key only in GitHub Secrets.

## Production Gates

Before cutting a release:

- Run the full verification command set above.
- Confirm `.env`/deployment secrets provide `VITE_API_URL`, `VITE_APP_ENV`, and observability keys as needed.
- Verify cash checkout both online and offline: cash sales must create a local receipt, record a cash drawer `sale` movement, and sync queued orders after reconnect.
- Verify card and QR checkout require the operator attestation and are labelled as operator-attested on screen. Do **not** treat either as a verified payment: this client has no terminal or payment-provider integration, and `ProcessPayment` sends only `{order_id, payment_method, amount_paid}`, so the server has nothing to check an authorisation against. A card sale here is the cashier's word, nothing more.
- Verify printer and cash drawer hardware on the target register OS.
- Verify the Tauri app bundle opens on the target OS and redirects unauthenticated users to auth.

## Known limitations

**Card and QR payments are not provider-confirmed.** No terminal, PSP or
gateway client exists in this codebase, and the payment contract
(`voxie/api/pos/orders/payments/v1`) carries no provider reference, auth code or
transaction id. Card and QR sales are recorded on the operator's explicit
attestation and land in the day summary as paid. Closing that gap needs a real
integration, not a checkbox.

**Tips cannot be collected on the single-sale screen.** `pos_orders.tip_amount`
exists and is read, but nothing writes it, and the payment contract has no tip
field, so `day_summary.total_tips` is structurally always `0.00`. The tip
control is therefore absent from the single-sale checkout. Adding one back
without the contract and column changes below books tip money as `change_due`,
which overstates cash against the drawer and makes the tip unrefundable.

## Native Notes

Tauri build output is intentionally ignored in git. Keep `core/Cargo.lock` committed for reproducible native application builds.
