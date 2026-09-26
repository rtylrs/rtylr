import type { AppId } from "@voxire/types";

export type { AppId };

function isLocalDevHost(hostname: string): boolean {
  const host = hostname.toLowerCase();
  return (
    host === "localhost" ||
    host === "127.0.0.1" ||
    /^\d+\.\d+\.\d+\.\d+$/.test(host)
  );
}

const IS_LOCAL =
  typeof window !== "undefined" && isLocalDevHost(window.location.hostname);

export const APP_URLS: Record<AppId, string> & {
  api: string;
  status: string;
  console: string;
} = {
  dash: IS_LOCAL ? "http://localhost:5175" : "https://dash.rtylr.com",
  auth: IS_LOCAL ? "http://localhost:5176" : "https://sso.rtylr.com",
  menu: IS_LOCAL ? "http://localhost:5174" : "https://menu.rtylr.com",
  pos: IS_LOCAL ? "http://localhost:5178" : "https://pos.rtylr.com",
  erp: IS_LOCAL ? "http://localhost:5179" : "https://erp.rtylr.com",
  hr: IS_LOCAL ? "http://localhost:5180" : "https://hr.rtylr.com",
  crm: IS_LOCAL ? "http://localhost:5181" : "https://crm.rtylr.com",
  finance: IS_LOCAL ? "http://localhost:5182" : "https://finance.rtylr.com",
  flow: IS_LOCAL ? "http://localhost:5183" : "https://flow.rtylr.com",
  insights: IS_LOCAL ? "http://localhost:5187" : "https://insights.rtylr.com",
  recruit: IS_LOCAL ? "http://localhost:5184" : "https://recruit.rtylr.com",
  landing: IS_LOCAL ? "http://localhost:4200" : "https://rtylr.com",
  api: "https://api.voxire.com",
  status: "https://status.rtylr.com",
  console: "https://console.voxire.com",
};

/**
 * Per-app display name and dev port.
 *
 * This carried an `allowedRoles: UserRole[]` per app until tryrtylr/rtylr#795.
 * Nothing ever read it: `AppGuard` does not import APP_CONFIG, and its actual
 * rule is `role === 'owner' || role === 'admin' || apps.includes(appId)` - the
 * per-user `apps` array is authoritative and the role name carries no weight
 * beyond that owner/admin special case.
 *
 * The table is not deleted lightly; it encoded a real intent, and it is recorded
 * on #795 so re-introducing app-level role gating does not start from scratch.
 * It is removed because a table that looks like authorization and enforces
 * nothing is worse than no table: the RBAC audit that produced #795 was itself
 * misled by it before tracing the real path.
 */
export const APP_CONFIG: Record<AppId, { name: string; devPort: number }> = {
  dash: { name: "Dash", devPort: 5175 },
  auth: { name: "Auth", devPort: 5176 },
  menu: { name: "Menu", devPort: 5174 },
  pos: { name: "POS", devPort: 5178 },
  erp: { name: "ERP", devPort: 5179 },
  hr: { name: "HR", devPort: 5180 },
  crm: { name: "CRM", devPort: 5181 },
  finance: { name: "Finance", devPort: 5182 },
  flow: { name: "Flow", devPort: 5183 },
  insights: { name: "Insights", devPort: 5187 },
  recruit: { name: "Recruit", devPort: 5184 },
  landing: { name: "Landing", devPort: 4200 },
};
