/**
 * Global URL configuration for all apps
 * Automatically switches between local and production based on hostname
 */

const isLocalhost =
  typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1');

const LOCAL_URLS = {
  auth: 'http://localhost:5176',
  landing: 'http://localhost:4200',
  menu: 'http://localhost:5177',
  account: 'http://localhost:5178',
  api: 'http://localhost:8080/api',
} as const;

const PROD_URLS = {
  auth: 'https://auth.rtylr.com',
  landing: 'https://rtylr.com',
  menu: 'https://menu.rtylr.com',
  account: 'https://account.rtylr.com',
  api: 'https://api.rtylr.com',
} as const;

export const APP_URLS = isLocalhost ? LOCAL_URLS : PROD_URLS;
