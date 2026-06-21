import { defineConfig } from 'vitepress';

export default defineConfig({
  title: 'rtylr Docs',
  description:
    'Documentation for rtylr — the all-in-one business operating system. POS, ERP, HR, CRM, Finance, Flow, Store, Book and more.',
  lang: 'en-US',
  cleanUrls: true,
  lastUpdated: true,

  head: [['link', { rel: 'icon', href: '/logo.png', type: 'image/png' }]],

  themeConfig: {
    logo: '/logo.png',

    search: {
      provider: 'local',
    },

    nav: [
      { text: 'Getting Started', link: '/getting-started/introduction' },
      { text: 'Apps', link: '/apps/hub' },
      { text: 'API', link: '/api/overview' },
      { text: 'Self-Hosting', link: '/self-hosting/architecture' },
      { text: 'FAQ', link: '/faq' },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        collapsed: false,
        items: [
          { text: 'Introduction', link: '/getting-started/introduction' },
          { text: 'Quick Start', link: '/getting-started/quick-start' },
          { text: 'Multi-Tenant Architecture', link: '/getting-started/multi-tenant' },
        ],
      },
      {
        text: 'Apps',
        collapsed: false,
        items: [
          { text: 'Hub (Command Center)', link: '/apps/hub' },
          { text: 'POS (Point of Sale)', link: '/apps/pos' },
          { text: 'ERP (Inventory & Supply Chain)', link: '/apps/erp' },
          { text: 'HR (Human Resources)', link: '/apps/hr' },
          { text: 'CRM (Customer Relations)', link: '/apps/crm' },
          { text: 'Finance', link: '/apps/finance' },
          { text: 'Flow (Project Management)', link: '/apps/flow' },
          { text: 'Insights (Analytics)', link: '/apps/insights' },
          { text: 'Book (Scheduling)', link: '/apps/book' },
          { text: 'Store (E-Commerce)', link: '/apps/store' },
          { text: 'Support (Help Desk)', link: '/apps/support' },
        ],
      },
      {
        text: 'Platform',
        collapsed: false,
        items: [{ text: 'Addon System', link: '/platform/addons' }],
      },
      {
        text: 'API',
        collapsed: false,
        items: [{ text: 'API Overview', link: '/api/overview' }],
      },
      {
        text: 'Self-Hosting',
        collapsed: false,
        items: [
          { text: 'Architecture', link: '/self-hosting/architecture' },
          { text: 'Image Contract', link: '/self-hosting/images' },
          { text: 'License Flow', link: '/self-hosting/license' },
          { text: 'Helm Install', link: '/self-hosting/helm' },
          { text: 'Docker Compose', link: '/self-hosting/compose' },
          { text: 'DNS Setup', link: '/self-hosting/dns' },
          { text: 'Backend Environment', link: '/self-hosting/backend-env' },
          { text: 'Upgrade Flow', link: '/self-hosting/upgrade' },
        ],
      },
      {
        text: 'Help',
        collapsed: false,
        items: [{ text: 'FAQ', link: '/faq' }],
      },
    ],

    socialLinks: [{ icon: 'github', link: 'https://github.com/voxire' }],

    footer: {
      message: 'Built by Voxire.',
      copyright: 'Copyright © 2026 rtylr',
    },
  },
});
