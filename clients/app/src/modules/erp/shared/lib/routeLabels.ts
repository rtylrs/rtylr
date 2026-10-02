/**
 * The name of every ERP route, in one place, as i18n keys.
 *
 * Three tables used to spell these out in English: the sidebar groups in
 * `DefaultLayout`, the page title beside them, and the command palette. Naming
 * a route therefore meant remembering three files, and they had already
 * drifted: `/erp/purchasing` reads "Purchase Orders" in the sidebar and the title
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
  '/erp': label('erpNav.dashboard'),
  '/erp/inventory': label('erpNav.inventory'),
  '/erp/products': label('erpNav.inventory'),
  '/erp/categories': label('erpNav.categories'),
  '/erp/brands': label('erpNav.brands'),
  '/erp/variants': label('erpNav.variants', 'erpNav.titles.variants'),
  '/erp/price-lists': label('erpNav.priceLists'),
  '/erp/barcodes': label('erpNav.barcodes', 'erpNav.titles.barcodes'),
  '/erp/warehouses': label('erpNav.warehouses'),
  '/erp/transfers': label('erpNav.stockTransfers'),
  '/erp/stock-takes': label('erpNav.stockTakes'),
  '/erp/receiving': label('erpNav.receiving', 'erpNav.titles.receiving'),
  '/erp/expiry': label('erpNav.expiry', 'erpNav.titles.expiry'),
  '/erp/alerts': label('erpNav.lowStockAlerts'),
  '/erp/reorder': label('erpNav.reorder', 'erpNav.titles.reorder'),
  '/erp/valuation': label('erpNav.valuation', 'erpNav.titles.valuation'),
  '/erp/purchasing': label('erpNav.purchaseOrders'),
  '/erp/suppliers': label('erpNav.suppliers'),
  '/erp/settings': label('erpNav.settings'),
};

/** Shown when a path matches no route, such as a deep link into a detail page. */
export const FALLBACK_TITLE_KEY = 'erpNav.appName';

/**
 * The page title key for a path.
 *
 * A nested path falls back to its longest matching parent, so
 * `/erp/products/42/edit` is titled from `/erp/products`. Longest first matters:
 * `/` prefixes everything, and sorting by length keeps it last.
 */
export function pageTitleKey(pathname: string): string {
  const exact = routeLabels[pathname];
  if (exact) return exact.title;

  const match = Object.keys(routeLabels)
    .filter((key) => key !== '/erp' && pathname.startsWith(`${key}/`))
    .sort((a, b) => b.length - a.length)[0];

  return match ? routeLabels[match].title : FALLBACK_TITLE_KEY;
}
