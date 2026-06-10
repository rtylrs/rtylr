import type { AppId, UserRole } from '@voxire/types';

export type { AppId };

function isLocalDevHost(hostname: string): boolean {
  const host = hostname.toLowerCase();
  return host === 'localhost' || host === '127.0.0.1' || /^\d+\.\d+\.\d+\.\d+$/.test(host);
}

const IS_LOCAL = typeof window !== 'undefined' && isLocalDevHost(window.location.hostname);

export const APP_URLS: Record<AppId, string> & { api: string; status: string; console: string } = {
  dash: IS_LOCAL ? 'http://localhost:5175' : 'https://dash.rtylr.com',
  auth: IS_LOCAL ? 'http://localhost:5176' : 'https://auth.rtylr.com',
  menu: IS_LOCAL ? 'http://localhost:5177' : 'https://menu.rtylr.com',
  pos: IS_LOCAL ? 'http://localhost:5178' : 'https://pos.rtylr.com',
  erp: IS_LOCAL ? 'http://localhost:5179' : 'https://erp.rtylr.com',
  hr: IS_LOCAL ? 'http://localhost:5180' : 'https://hr.rtylr.com',
  crm: IS_LOCAL ? 'http://localhost:5181' : 'https://crm.rtylr.com',
  finance: IS_LOCAL ? 'http://localhost:5182' : 'https://finance.rtylr.com',
  flow: IS_LOCAL ? 'http://localhost:5183' : 'https://flow.rtylr.com',
  insights: IS_LOCAL ? 'http://localhost:5187' : 'https://insights.rtylr.com',
  recruit: IS_LOCAL ? 'http://localhost:5184' : 'https://recruit.rtylr.com',
  landing: IS_LOCAL ? 'http://localhost:4200' : 'https://rtylr.com',
  api: 'https://api.voxire.com',
  status: 'https://status.rtylr.com',
  console: 'https://console.voxire.com',
};

export const APP_CONFIG: Record<
  AppId,
  { name: string; devPort: number; allowedRoles: UserRole[] }
> = {
  dash: { name: 'Dash', devPort: 5175, allowedRoles: ['owner', 'admin', 'operator'] },
  auth: {
    name: 'Auth',
    devPort: 5176,
    allowedRoles: ['owner', 'admin', 'operator', 'staff', 'viewer'],
  },
  menu: { name: 'Menu', devPort: 5177, allowedRoles: [] },
  pos: { name: 'POS', devPort: 5178, allowedRoles: ['owner', 'admin', 'operator', 'staff'] },
  erp: { name: 'ERP', devPort: 5179, allowedRoles: ['owner', 'admin', 'operator'] },
  hr: { name: 'HR', devPort: 5180, allowedRoles: ['owner', 'admin', 'operator', 'staff'] },
  crm: { name: 'CRM', devPort: 5181, allowedRoles: ['owner', 'admin', 'operator'] },
  finance: { name: 'Finance', devPort: 5182, allowedRoles: ['owner', 'admin'] },
  flow: { name: 'Flow', devPort: 5183, allowedRoles: ['owner', 'admin', 'operator', 'staff'] },
  insights: { name: 'Insights', devPort: 5187, allowedRoles: ['owner', 'admin'] },
  // Recruit is public on the candidate side; the admin side is gated to
  // owner/admin only. allowedRoles reflects the admin gate - the public
  // pages don't go through AppGuard at all.
  recruit: { name: 'Recruit', devPort: 5184, allowedRoles: ['owner', 'admin'] },
  landing: { name: 'Landing', devPort: 4200, allowedRoles: [] },
};
