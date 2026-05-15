/**
 * Build the Caret `systemContext` string for the dash app. The backend
 * splices this into the model's system prompt so Caret knows where the
 * user currently is in the product.
 */
export interface DashSystemContextInput {
  pathname: string;
  businessName?: string;
}

const PAGE_HINTS: Record<string, string> = {
  '/': 'on the dashboard home (KPIs and quick overview)',
  '/team': 'on the Team management page',
  '/apps': 'on the Apps marketplace',
  '/menus': 'managing menus',
  '/orders': 'reviewing orders',
  '/qr-codes': 'generating QR codes',
  '/design': 'editing storefront design',
  '/settings': 'configuring settings',
};

function pageHint(pathname: string): string {
  if (PAGE_HINTS[pathname]) return PAGE_HINTS[pathname];
  const match = Object.keys(PAGE_HINTS).find((p) => pathname.startsWith(p) && p !== '/');
  return match ? PAGE_HINTS[match] : 'navigating the dash app';
}

export function buildDashSystemContext({ pathname, businessName }: DashSystemContextInput): string {
  const where = pageHint(pathname);
  const biz = businessName ? `Business: ${businessName}. ` : '';
  return `${biz}User is in the Dash app (business command center) and is currently ${where}.`;
}
