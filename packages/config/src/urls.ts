/**
 * Global URL configuration for all apps
 * Automatically switches between local and production based on hostname
 *
 * Apps:
 *   landing  → rtylr.com
 *   auth     → auth.rtylr.com
 *   hub      → hub.rtylr.com
 *   menu     → menu.rtylr.com
 *   pos      → pos.rtylr.com  (vbitez)
 *   api      → api.rtylr.com
 */

const isLocalhost =
  typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1');

const LOCAL_URLS = {
  landing: 'http://localhost:4200',
  auth: 'http://localhost:5176',
  hub: 'http://localhost:5175',
  menu: 'http://localhost:5177',
  pos: 'http://localhost:5178',
  api: 'https://api.voxire.dev',
  status: 'https://status.rtylr.com',
} as const;

const PROD_URLS = {
  landing: 'https://rtylr.com',
  auth: 'https://auth.rtylr.com',
  hub: 'https://hub.rtylr.com',
  menu: 'https://menu.rtylr.com',
  pos: 'https://pos.rtylr.com',
  api: 'https://api.voxire.dev',
  status: 'https://status.rtylr.com',
} as const;

export const APP_URLS = isLocalhost ? LOCAL_URLS : PROD_URLS;
