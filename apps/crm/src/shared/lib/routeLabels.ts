export const routeLabels: Record<string, string> = {
  '/': 'Dashboard',
  '/customers': 'Customers',
  '/customers/new': 'New customer',
  '/customers/import': 'Import customers',
  '/loyalty': 'Loyalty',
  '/loyalty/members': 'Loyalty members',
  '/loyalty/rewards': 'Rewards catalog',
  '/loyalty/tiers': 'Loyalty tiers',
  '/campaigns': 'Campaigns',
  '/campaigns/email': 'New email campaign',
  '/campaigns/sms': 'New SMS campaign',
  '/segments': 'Segments',
  '/segments/new': 'New segment',
  '/automations': 'Automations',
  '/automations/new': 'New automation',
  '/templates': 'Templates',
  '/feedback': 'Feedback & NPS',
  '/settings': 'Settings',
  '/sales': 'Pipeline',
  '/sales/leads': 'Leads',
  '/sales/contracts': 'Contracts',
};

export function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];

  const match = Object.keys(routeLabels)
    .filter((key) => key !== '/' && pathname.startsWith(`${key}/`))
    .sort((a, b) => b.length - a.length)[0];

  return match ? routeLabels[match] : 'CRM';
}
