# Portal workspace

`clients/app` owns the browser workspace at `portal.rtylr.com`.
The first module is Home: the account and business features previously deployed
as Dash. Auth remains at `sso.rtylr.com`.

## Ownership

- `main.tsx`: one query client, auth cookie provider, locale instance and theme.
- `shared/utils/BusinessRoute.tsx`: fresh account and business access hydration,
  onboarding and MFA gates before the workspace renders.
- `shell/`: persistent app rail, collapsible page navigation, profile,
  business switcher, notifications, command palette and Caret.
- `features/`: Home, team, groups, permissions, Extensions and account/business settings.
- `modules/insights/`: analytics pages with their own header inside the shared shell.
- `modules/flow/`, `modules/crm/`, `modules/erp/`: routes and page navigation in the shared workspace.
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

## Business modules and public entrypoints

All business modules now render inside Portal. Finance, HR, web POS and Recruit administration use `/finance`, `/hr`, `/pos` and `/recruit`. Home includes Menu management. Each module registers its page navigation and settings beneath `/settings`.

`clients/app` is the single web package. `entrypoints/auth`, `entrypoints/menu` and `entrypoints/recruit` preserve SSO and public-site URLs as separate builds without separate workspace packages. Native POS remains at `clients/pos`, and MCP is at `clients/mcp`.

POS reuses the native React components through a route adapter. Its local cart, terminal, employee session and offline queues are partitioned by the selected business. Native storage is preserved. The adapter does not bypass location selection or employee PIN authentication. Finance browser preferences are also partitioned by business; they are local preferences rather than server-backed business configuration.

Query keys are prefixed by app. Portal refreshes the account authorization snapshot before rendering, and business selection reloads the workspace to clear old query state. Backend handlers remain authoritative for permissions.

## Release order

1. Provision the Portal Pages project and `portal.rtylr.com` custom domain in infra.
2. Promote the source release's Portal build through `tryrtylr/deploy` and verify it.
3. Enable the Dash-to-Portal and Insights-to-Portal edge rules, then retire their
   standalone deployment manifests. Insights deep links map beneath `/insights`.
   Flow moves with the same order: deploy Portal first, redirect Flow links beneath
   `/flow`, then retire its Pages project and deployment manifest. CRM, ERP, Finance, HR and web POS follow the same sequence with their route prefixes. Recruit keeps its public host; old `/admin` links redirect into Portal.
4. Promote Auth and shared-URL consumers, including Onion's generated email links.

Do not activate the Dash redirect before Portal is deployed and reachable.
