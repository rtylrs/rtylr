/**
 * The name of every CRM route, in one place, as i18n keys.
 *
 * This table already existed and was already shared, which is why the page
 * title and the breadcrumb agreed. What it held was English, so the sidebar in
 * `DefaultLayout` and the command palette each kept their own copy of the same
 * names and the app had three vocabularies for one set of routes.
 *
 * Two of those are deliberate and both are kept. The sidebar wants a short
 * label under an icon ("Feedback"); the page title and the palette want the
 * full name ("Feedback & NPS"). A route whose full name is its short label
 * carries one key for both.
 */
export interface RouteLabel {
  /** Short label, for the sidebar. */
  readonly nav: string;
  /** Full name, for the page title and the command palette. */
  readonly title: string;
}

const label = (nav: string, title: string = nav): RouteLabel => ({ nav, title });

export const routeLabels: Readonly<Record<string, RouteLabel>> = {
  '/': label('nav.dashboard'),
  '/customers': label('nav.customers'),
  '/customers/new': label('nav.newCustomer'),
  '/customers/import': label('nav.importCustomers'),
  '/loyalty': label('nav.loyalty'),
  '/loyalty/members': label('nav.loyaltyMembers'),
  '/loyalty/rewards': label('nav.rewardsCatalog'),
  '/loyalty/tiers': label('nav.loyaltyTiers'),
  '/campaigns': label('nav.campaigns'),
  '/campaigns/email': label('nav.newEmailCampaign'),
  '/campaigns/sms': label('nav.newSmsCampaign'),
  '/segments': label('nav.segments'),
  '/segments/new': label('nav.newSegment'),
  '/automations': label('nav.automations'),
  '/automations/new': label('nav.newAutomation'),
  '/templates': label('nav.templates'),
  '/feedback': label('nav.feedback', 'nav.titles.feedback'),
  '/settings': label('nav.settings'),
  '/sales': label('nav.pipeline'),
  '/sales/leads': label('nav.leads'),
  '/sales/contracts': label('nav.contracts'),
};

/** Shown when a path matches no route, such as a deep link into a detail page. */
export const FALLBACK_TITLE_KEY = 'nav.appName';

/**
 * The page title key for a path.
 *
 * A nested path falls back to its longest matching parent, so
 * `/customers/42` is titled from `/customers`. Longest first matters: both
 * `/loyalty` and `/loyalty/members` prefix `/loyalty/members/3`, and only the
 * longer one is right.
 */
export function pageTitleKey(pathname: string): string {
  const exact = routeLabels[pathname];
  if (exact) return exact.title;

  const match = Object.keys(routeLabels)
    .filter((key) => key !== '/' && pathname.startsWith(`${key}/`))
    .sort((a, b) => b.length - a.length)[0];

  return match ? routeLabels[match].title : FALLBACK_TITLE_KEY;
}
