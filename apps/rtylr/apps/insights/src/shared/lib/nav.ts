export const routeLabels: Record<string, string> = {
  '/': 'Overview',
  '/revenue': 'Revenue',
  '/customers': 'Customers',
  '/products': 'Products',
  '/operations': 'Operations',
  '/financial': 'Financial',
  '/reports': 'Reports',
  '/settings': 'Settings',
};

export function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  const match = Object.keys(routeLabels).find((key) => pathname.startsWith(key) && key !== '/');
  return match ? routeLabels[match] : 'Insights';
}
