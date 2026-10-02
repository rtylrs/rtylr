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

export const PORTAL_URL = IS_LOCAL
  ? "http://localhost:5175"
  : "https://portal.rtylr.com";

export const APP_URLS: Record<AppId, string> & {
  api: string;
  status: string;
  console: string;
} = {
  dash: PORTAL_URL,
  auth: IS_LOCAL ? "http://localhost:5176" : "https://sso.rtylr.com",
  menu: IS_LOCAL ? "http://localhost:5174" : "https://menu.rtylr.com",
  pos: IS_LOCAL ? "http://localhost:5178" : "https://pos.rtylr.com",
  erp: IS_LOCAL ? "http://localhost:5179" : "https://erp.rtylr.com",
  hr: IS_LOCAL ? "http://localhost:5180" : "https://hr.rtylr.com",
  crm: IS_LOCAL ? "http://localhost:5181" : "https://crm.rtylr.com",
  finance: IS_LOCAL ? "http://localhost:5182" : "https://finance.rtylr.com",
  flow: IS_LOCAL ? "http://localhost:5183" : "https://flow.rtylr.com",
  insights: `${PORTAL_URL}/insights`,
  recruit: IS_LOCAL ? "http://localhost:5184" : "https://recruit.rtylr.com",
  landing: IS_LOCAL ? "http://localhost:4200" : "https://rtylr.com",
  api: "https://api.voxire.com",
  status: "https://status.rtylr.com",
  console: "https://voxire.partials.app",
};

export const APP_CONFIG: Record<AppId, { name: string; devPort: number }> = {
  dash: { name: "Home", devPort: 5175 },
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
