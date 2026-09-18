/**
 * The name of every ERP route, in one place, as i18n keys.
 *
 * Three tables used to spell these out in English: the sidebar groups in
 * `DefaultLayout`, the page title beside them, and the command palette. Naming
 * a route therefore meant remembering three files, and they had already
 * drifted: `/purchasing` reads "Purchase Orders" in the sidebar and the title
 * and read "Purchasing" in the palette.
 *
 * Two vocabularies are deliberate, and both are kept. The sidebar wants a short
 * label under an icon ("Variants", "Expiry"). The page title and the command
 * palette want the full name ("Product Variants", "Expiry Tracking"). A route
 * whose full name is its short label carries one key for both.
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
  '/inventory': label('nav.inventory'),
  '/products': label('nav.inventory'),
  '/categories': label('nav.categories'),
  '/brands': label('nav.brands'),
  '/variants': label('nav.variants', 'nav.titles.variants'),
  '/price-lists': label('nav.priceLists'),
  '/barcodes': label('nav.barcodes', 'nav.titles.barcodes'),
  '/warehouses': label('nav.warehouses'),
  '/transfers': label('nav.stockTransfers'),
  '/stock-takes': label('nav.stockTakes'),
  '/receiving': label('nav.receiving', 'nav.titles.receiving'),
  '/expiry': label('nav.expiry', 'nav.titles.expiry'),
  '/alerts': label('nav.lowStockAlerts'),
  '/reorder': label('nav.reorder', 'nav.titles.reorder'),
  '/valuation': label('nav.valuation', 'nav.titles.valuation'),
  '/purchasing': label('nav.purchaseOrders'),
  '/suppliers': label('nav.suppliers'),
  '/settings': label('nav.settings'),
};

/** Shown when a path matches no route, such as a deep link into a detail page. */
export const FALLBACK_TITLE_KEY = 'nav.appName';

/**
 * The page title key for a path.
 *
 * A nested path falls back to its longest matching parent, so
 * `/products/42/edit` is titled from `/products`. Longest first matters:
 * `/` prefixes everything, and sorting by length keeps it last.
 */
export function pageTitleKey(pathname: string): string {
  const exact = routeLabels[pathname];
  if (exact) return exact.title;

  const match = Object.keys(routeLabels)
    .filter((key) => key !== '/' && pathname.startsWith(`${key}/`))
    .sort((a, b) => b.length - a.length)[0];

  return match ? routeLabels[match].title : FALLBACK_TITLE_KEY;
}
