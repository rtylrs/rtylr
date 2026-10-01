export type NavGroup = {
  label: string;
  items: Array<{ label: string; href: string; detail?: string }>;
};

export const navGroups: NavGroup[] = [
  {
    label: 'Build',
    items: [
      { label: 'Apps', href: '/apps/hub', detail: 'Run every business app from one workspace.' },
      { label: 'Getting started', href: '/getting-started/introduction', detail: 'Set up rtylr for your team.' },
      { label: 'Platform guide', href: '/platform/addons', detail: 'Understand the shared platform.' },
    ],
  },
  {
    label: 'Automate',
    items: [
      { label: 'Flow', href: '/apps/flow', detail: 'Turn repeatable work into clear workflows.' },
      { label: 'Integrations', href: '/platform/addons', detail: 'Connect the tools your business already uses.' },
      { label: 'API', href: '/api/overview', detail: 'Build on top of rtylr programmatically.' },
    ],
  },
  {
    label: 'Code',
    items: [
      { label: 'API reference', href: '/api/overview', detail: 'Explore the API surface.' },
      { label: 'MCP', href: '/api/overview', detail: 'Connect agents and developer tools.' },
    ],
  },
  {
    label: 'Manage',
    items: [
      { label: 'Hub', href: '/apps/hub', detail: 'Manage your organization from the command center.' },
      { label: 'Multi-tenant architecture', href: '/getting-started/multi-tenant', detail: 'Understand organizations and tenants.' },
      { label: 'FAQ', href: '/faq', detail: 'Find answers to common questions.' },
    ],
  },
  {
    label: 'Resources',
    items: [
      { label: 'Documentation map', href: '/', detail: 'Find the right rtylr guide.' },
      { label: 'FAQ', href: '/faq', detail: 'Find answers to common questions.' },
    ],
  },
];

export const sidebarGroups = [
  {
    label: 'Getting Started',
    items: [
      ['Introduction', '/getting-started/introduction'],
      ['Quick Start', '/getting-started/quick-start'],
      ['Multi-Tenant Architecture', '/getting-started/multi-tenant'],
    ],
  },
  {
    label: 'Apps',
    items: [
      ['Hub', '/apps/hub'],
      ['POS', '/apps/pos'],
      ['ERP', '/apps/erp'],
      ['HR', '/apps/hr'],
      ['CRM', '/apps/crm'],
      ['Finance', '/apps/finance'],
      ['Flow', '/apps/flow'],
      ['Insights', '/apps/insights'],
      ['Store', '/apps/store'],
      ['Support', '/apps/support'],
    ],
  },
  {
    label: 'Platform',
    items: [['Addon System', '/platform/addons']],
  },
  {
    label: 'API',
    items: [['API Overview', '/api/overview']],
  },
  {
    label: 'Help',
    items: [['FAQ', '/faq']],
  },
] as const;
