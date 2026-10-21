import { APP_URLS } from '@voxire/config';

/**
 * The header/footer nav's link data, in one place so the desktop header,
 * mobile header, and footer can't drift into three different English
 * copies of the same list. `key` addresses the copy in
 * common.nav.featureLinks/resourceLinks/companyLinks; `href` is the bare
 * English route, localised at render time via isLocalisedRoute/localisePath
 * for whichever of these routes have a translation.
 */

export interface FeatureLink {
  href: string;
  key: 'overview' | 'work' | 'finance' | 'customers' | 'team' | 'operations' | 'sales' | 'insights';
}

export interface LabelledLink {
  href: string;
  key: string;
  external?: boolean;
}

// Every area lives on one page; each link jumps to its section. The
// fragments match the section ids in pages/workspace/index.tsx.
export const FEATURE_LINKS: FeatureLink[] = [
  { href: '/workspace/', key: 'overview' },
  { href: '/workspace/#work', key: 'work' },
  { href: '/workspace/#finance', key: 'finance' },
  { href: '/workspace/#customers', key: 'customers' },
  { href: '/workspace/#team', key: 'team' },
  { href: '/workspace/#operations', key: 'operations' },
  { href: '/workspace/#sales', key: 'sales' },
  { href: '/workspace/#insights', key: 'insights' },
];

export const RESOURCE_LINKS: LabelledLink[] = [
  { href: '/hardware', key: 'hardware' },
  { href: '/integrations', key: 'integrations' },
  { href: '/blog', key: 'blog' },
  { href: 'https://docs.rtylr.com', key: 'documentation', external: true },
  { href: '/support', key: 'helpCenter' },
  { href: '/trust', key: 'trustCenter' },
  { href: '/legal/security', key: 'security' },
  { href: '/legal/privacy-policy', key: 'privacyPolicy' },
  { href: '/legal/terms', key: 'termsOfUse' },
  { href: APP_URLS.status, key: 'status', external: true },
];

export const COMPANY_LINKS: LabelledLink[] = [
  { href: '/about', key: 'about' },
  { href: '/partners', key: 'partners' },
  { href: 'https://voxire.com/careers', key: 'careers', external: true },
  { href: '/contact', key: 'contact' },
];
