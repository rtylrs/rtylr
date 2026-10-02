# Portal workspace

`clients/web/rtylr/apps/portal` owns the browser workspace at `portal.rtylr.com`.
The first module is Home: the account and business features previously deployed
as Dash. Auth remains at `sso.rtylr.com`.

## Ownership

- `main.tsx`: one query client, auth cookie provider, locale instance and theme.
- `shared/utils/BusinessRoute.tsx`: fresh account and business access hydration,
  onboarding and MFA gates before the workspace renders.
- `shell/`: persistent app rail, collapsible page navigation, profile,
  business switcher, notifications, command palette and Caret.
- `features/`: Home, team, groups, permission profiles and account/business settings.
- `modules/insights/`: analytics pages with their own header inside the shared shell.
- `shared/utils/canAccessPortalPath.ts`: the route visibility rule shared by
  page navigation and the rendered route gate. The API enforces authorization.

The API's existing `dash` application identifier and `hub:*` permissions describe
account management. They do not determine the frontend hostname. The web package
and deployment are named Portal; `APP_URLS.dash` resolves to Portal for existing
API consumers. No additional Portal role or separate authorization document is
introduced.

## Navigation and sessions

Home is available to every authenticated business member. Owner and admin access
comes from the server's effective-access document. Member page links require the
corresponding permission. Changing pages retains the shell and shared providers.

At cutover, Dash links must redirect at Cloudflare's edge to Portal with their path and
query string preserved. A redirect without a target fragment retains the source
fragment in the browser. SSO receives a relative `returnTo` path and returns only
to the configured Portal origin; external and protocol-relative destinations are
rejected. The existing `.rtylr.com` auth cookie spans SSO and Portal.

## Moving the next app

Home, Insights and Flow now use the shared Portal entry point. Insights keeps its
page selector in a top header; its routes live under `/insights`. Flow has native
`/flow` routes, its own navigation and the shared launcher, Caret, business and
account context. Personal and app settings share `/settings`, including
`/settings/flow` and `/settings/insights`. ERP, HR,
CRM, Finance and web POS still open their standalone frontends in another tab.
Menu and Recruit are setup and public products rather than workspace app entries.
No standalone frontend is embedded in an iframe.

For each migration, add its routes beneath a stable prefix (for example `/flow`),
register its rail entry, and lift its bootstrap providers into the shared workspace.
Keep page state and drafts in the module's stores or query cache. Do not mount
its existing standalone `main.tsx`, a second router or a second locale instance.
Retire its domain only after the Portal routes and deep links are verified.

Native POS and its offline behavior remain separate from this browser migration.
Business subdomains are a later routing feature and are not enabled here.

## Release order

1. Provision the Portal Pages project and `portal.rtylr.com` custom domain in infra.
2. Promote the source release's Portal build through `tryrtylr/deploy` and verify it.
3. Enable the Dash-to-Portal and Insights-to-Portal edge rules, then retire their
   standalone deployment manifests. Insights deep links map beneath `/insights`.
   Flow moves with the same order: deploy Portal first, redirect Flow links beneath
   `/flow`, then retire its Pages project and deployment manifest.
4. Promote Auth and shared-URL consumers, including Onion's generated email links.

Do not activate the Dash redirect before Portal is deployed and reachable.
