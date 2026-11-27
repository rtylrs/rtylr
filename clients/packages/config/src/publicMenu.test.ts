import { describe, expect, it } from 'vitest';
import { PORTAL_URL } from './apps';
import {
  RESERVED_BUSINESS_SLUGS,
  buildPublicMenuUrl,
  buildPublicMenuUrlById,
  buildTenantSubdomainMenuUrl,
  extractTenantSubdomain,
  isReservedBusinessSlug,
  isTenantMenuHost,
} from './publicMenu';

const ORIGIN = PORTAL_URL.replace(/\/+$/, '');

describe('RESERVED_BUSINESS_SLUGS', () => {
  it('covers the Portal route names and POS-style prefixes', () => {
    expect(RESERVED_BUSINESS_SLUGS).toEqual([
      'automations',
      'webhooks',
      'menus',
      'menu',
      'create-item',
      'create-banner',
      'orders',
      'qr-codes',
      'design',
      'chat',
      'caret',
      'team',
      'roles',
      'audit',
      'notifications',
      'account',
      'billing',
      'theme',
      'help',
      'extensions',
      'store',
      'integrations',
      'insights',
      'flow',
      'crm',
      'erp',
      'finance',
      'hr',
      'recruit',
      'settings',
      'security',
      'api-keys',
      'pos',
      'b',
      'api',
      'auth',
      'login',
      'signup',
      'assets',
      'static',
      'public',
      'admin',
      'www',
      'app',
      'portal',
      'board',
      'display',
      'poster',
      'try',
      'setup',
      't',
    ]);
  });
});

describe('isReservedBusinessSlug', () => {
  it.each(RESERVED_BUSINESS_SLUGS)('reserves %s', (slug) => {
    expect(isReservedBusinessSlug(slug)).toBe(true);
  });

  it('trims and lowercases before comparing', () => {
    expect(isReservedBusinessSlug('API-KEYS')).toBe(true);
    expect(isReservedBusinessSlug('  Menu  ')).toBe(true);
  });

  it('leaves ordinary slugs and empty input alone', () => {
    expect(isReservedBusinessSlug('acme-cafe')).toBe(false);
    expect(isReservedBusinessSlug('menu-2')).toBe(false);
    expect(isReservedBusinessSlug('')).toBe(false);
    expect(isReservedBusinessSlug('   ')).toBe(false);
  });
});

describe('buildPublicMenuUrl', () => {
  it('builds the slug menu URL on the Portal origin', () => {
    expect(buildPublicMenuUrl('acme')).toBe(`${ORIGIN}/acme/menu`);
  });

  it('appends a location segment when given', () => {
    expect(buildPublicMenuUrl('acme', 'downtown')).toBe(`${ORIGIN}/acme/menu/downtown`);
  });

  it('encodes the slug and location and never leaves a trailing slash', () => {
    const url = buildPublicMenuUrl('acme cafe', 'a/b');
    expect(url).toBe(`${ORIGIN}/acme%20cafe/menu/a%2Fb`);
    expect(url.endsWith('/')).toBe(false);
  });
});

describe('buildPublicMenuUrlById', () => {
  it('uses the POS-style /b/<id>/menu form', () => {
    expect(buildPublicMenuUrlById(42)).toBe(`${ORIGIN}/b/42/menu`);
    expect(buildPublicMenuUrlById('42')).toBe(`${ORIGIN}/b/42/menu`);
  });
});

describe('tenant subdomain and custom domain host detection', () => {
  it('extracts valid non-reserved <slug>.rtylr.com subdomains', () => {
    expect(extractTenantSubdomain('voxire.rtylr.com')).toBe('voxire');
    expect(extractTenantSubdomain('Acme-Cafe.RTYLR.COM')).toBe('acme-cafe');
  });

  it('rejects reserved platform subdomains and multi-level hosts', () => {
    expect(extractTenantSubdomain('portal.rtylr.com')).toBeNull();
    expect(extractTenantSubdomain('pos.rtylr.com')).toBeNull();
    expect(extractTenantSubdomain('sso.rtylr.com')).toBeNull();
    expect(extractTenantSubdomain('rennili.rtylr.com')).toBeNull();
    expect(extractTenantSubdomain('a.b.rtylr.com')).toBeNull();
    expect(extractTenantSubdomain('rtylr.com')).toBeNull();
  });

  it('distinguishes tenant menu hosts from platform and local hosts', () => {
    expect(isTenantMenuHost('voxire.rtylr.com')).toBe(true);
    expect(isTenantMenuHost('menu.myrestaurant.com')).toBe(true);
    expect(isTenantMenuHost('portal.rtylr.com')).toBe(false);
    expect(isTenantMenuHost('rtylr.com')).toBe(false);
    expect(isTenantMenuHost('rtylr-portal.pages.dev')).toBe(false);
    expect(isTenantMenuHost('localhost')).toBe(false);
    expect(isTenantMenuHost('127.0.0.1')).toBe(false);
  });

  it('builds branded tenant subdomain menu URLs', () => {
    expect(buildTenantSubdomainMenuUrl('voxire')).toBe('https://voxire.rtylr.com');
    expect(buildTenantSubdomainMenuUrl('voxire', 'voxire-cafe')).toBe(
      'https://voxire.rtylr.com/voxire-cafe'
    );
  });
});
