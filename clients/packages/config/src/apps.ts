import type { AppId } from '@voxire/types';

export type { AppId };

function isLocalDevHost(hostname: string): boolean {
  const host = hostname.toLowerCase();
  return host === 'localhost' || host === '127.0.0.1' || /^\d+\.\d+\.\d+\.\d+$/.test(host);
}

const IS_LOCAL = typeof window !== 'undefined' && isLocalDevHost(window.location.hostname);

export const PORTAL_URL = IS_LOCAL ? 'http://localhost:5175' : 'https://portal.rtylr.com';

export const APP_URLS: Record<AppId, string> & {
  api: string;
  status: string;
  console: string;
} = {
  dash: PORTAL_URL,
  auth: IS_LOCAL ? 'http://localhost:5176' : 'https://sso.rtylr.com',
  menu: IS_LOCAL ? 'http://localhost:5174' : 'https://menu.rtylr.com',
  pos: IS_LOCAL ? 'http://localhost:5178' : 'https://pos.rtylr.com',
  erp: `${PORTAL_URL}/erp`,
  hr: `${PORTAL_URL}/hr`,
  crm: `${PORTAL_URL}/crm`,
  finance: `${PORTAL_URL}/finance`,
  flow: `${PORTAL_URL}/flow`,
  insights: `${PORTAL_URL}/insights`,
  recruit: `${PORTAL_URL}/recruit`,
  landing: IS_LOCAL ? 'http://localhost:4200' : 'https://rtylr.com',
  api: 'https://api.voxire.com',
  status: 'https://status.rtylr.com',
  console: 'https://voxire.partials.app',
};

export const RECRUIT_PUBLIC_URL = IS_LOCAL ? 'http://localhost:5184' : 'https://recruit.rtylr.com';

export const APP_CONFIG: Record<AppId, { name: string; devPort: number }> = {
  dash: { name: 'Home', devPort: 5175 },
  auth: { name: 'Auth', devPort: 5176 },
  menu: { name: 'Menu', devPort: 5174 },
  pos: { name: 'POS', devPort: 5178 },
  erp: { name: 'ERP', devPort: 5175 },
  hr: { name: 'HR', devPort: 5175 },
  crm: { name: 'CRM', devPort: 5175 },
  finance: { name: 'Finance', devPort: 5175 },
  flow: { name: 'Flow', devPort: 5175 },
  insights: { name: 'Insights', devPort: 5175 },
  recruit: { name: 'Recruit', devPort: 5175 },
  landing: { name: 'Landing', devPort: 4200 },
};
