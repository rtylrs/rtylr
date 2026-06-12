import React from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import {
  Search,
  ArrowRight,
  Book,
  Code,
  Zap,
  Sparkles,
  ShoppingCart,
  Package,
  Users,
  Headphones,
  Wallet,
  CalendarCheck,
  Store as StoreIcon,
  ListChecks,
  BarChart3,
  LayoutGrid,
} from 'lucide-react';

interface AppCard {
  slug: string;
  name: string;
  tagline: string;
  icon: React.ReactNode;
  accent: string;
}

const apps: AppCard[] = [
  {
    slug: 'hub',
    name: 'Hub',
    tagline: 'Business command center - branches, team, settings, billing.',
    icon: <LayoutGrid size={18} />,
    accent: '#4583c4',
  },
  {
    slug: 'pos',
    name: 'POS',
    tagline: 'Multi-branch, multi-terminal point of sale with offline mode.',
    icon: <ShoppingCart size={18} />,
    accent: '#4583c4',
  },
  {
    slug: 'erp',
    name: 'ERP',
    tagline: 'Inventory, suppliers, purchase orders, multi-location stock.',
    icon: <Package size={18} />,
    accent: '#f59e0b',
  },
  {
    slug: 'hr',
    name: 'HR',
    tagline: 'Employees, shifts, leave, payroll, performance.',
    icon: <Users size={18} />,
    accent: '#8b5cf6',
  },
  {
    slug: 'crm',
    name: 'CRM',
    tagline: 'Customers, deals, loyalty, segments, marketing.',
    icon: <Sparkles size={18} />,
    accent: '#ec4899',
  },
  {
    slug: 'finance',
    name: 'Finance',
    tagline: 'Invoices, payments, P&L, balance sheet, cash flow.',
    icon: <Wallet size={18} />,
    accent: '#10b981',
  },
  {
    slug: 'flow',
    name: 'Flow',
    tagline: 'Projects, tasks, kanban, automations across the business.',
    icon: <ListChecks size={18} />,
    accent: '#6366f1',
  },
  {
    slug: 'store',
    name: 'Store',
    tagline: 'Online storefront with custom domains and checkout.',
    icon: <StoreIcon size={18} />,
    accent: '#0ea5e9',
  },
  {
    slug: 'book',
    name: 'Book',
    tagline: 'Appointments, reservations, calendar, reminders.',
    icon: <CalendarCheck size={18} />,
    accent: '#14b8a6',
  },
  {
    slug: 'support',
    name: 'Support',
    tagline: 'Shared inbox, tickets, SLAs, knowledge base.',
    icon: <Headphones size={18} />,
    accent: '#f97316',
  },
  {
    slug: 'insights',
    name: 'Insights',
    tagline: 'Cross-app analytics, dashboards, custom reports.',
    icon: <BarChart3 size={18} />,
    accent: '#3b82f6',
  },
];

const Home: React.FC = () => {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = React.useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate({ to: '/search', search: { q: searchQuery.trim() } });
    }
  };

  return (
    <>
      <HelmetContainer title="Documentation · rtylr" />
      <div>
        <div className="mb-10">
          <h1 className="text-3xl font-bold mb-3" style={{ color: 'var(--docs-text)' }}>
            rtylr Documentation
          </h1>
          <p className="text-base mb-6" style={{ color: 'var(--docs-text-muted)' }}>
            Pick an app to see what it does, every feature it ships with, and how to use it.
          </p>

          <form onSubmit={handleSearch} className="max-w-lg">
            <div className="relative">
              <Search
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2"
                style={{ color: 'var(--docs-text-dim)' }}
              />
              <input
                type="text"
                placeholder="Search the docs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2.5 text-sm rounded-lg border outline-none transition-colors"
                style={{
                  borderColor: 'var(--docs-border)',
                  background: 'var(--docs-bg)',
                  color: 'var(--docs-text)',
                }}
              />
            </div>
          </form>
        </div>

        <div className="mb-12">
          <h2
            className="text-sm font-semibold uppercase tracking-wider mb-4"
            style={{ color: 'var(--docs-text-muted)' }}
          >
            Apps
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {apps.map((app) => (
              <Link
                key={app.slug}
                to="/docs/$slug"
                params={{ slug: app.slug }}
                className="group flex flex-col p-4 rounded-lg border no-underline transition-all hover:-translate-y-0.5"
                style={{ borderColor: 'var(--docs-border)', background: 'var(--docs-bg)' }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span
                    className="h-7 w-7 rounded-md inline-flex items-center justify-center"
                    style={{ background: `${app.accent}1A`, color: app.accent }}
                  >
                    {app.icon}
                  </span>
                  <span className="text-sm font-semibold" style={{ color: 'var(--docs-text)' }}>
                    {app.name}
                  </span>
                  <ArrowRight
                    size={14}
                    className="ml-auto opacity-0 group-hover:opacity-100 transition-opacity"
                    style={{ color: 'var(--docs-text-dim)' }}
                  />
                </div>
                <p className="text-sm" style={{ color: 'var(--docs-text-muted)' }}>
                  {app.tagline}
                </p>
              </Link>
            ))}
          </div>
        </div>

        <div className="mb-10">
          <h2
            className="text-sm font-semibold uppercase tracking-wider mb-4"
            style={{ color: 'var(--docs-text-muted)' }}
          >
            Get Started
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {[
              {
                title: 'Introduction',
                description: 'What rtylr is and how it fits your business.',
                href: '/docs/introduction',
                icon: <Book size={16} />,
              },
              {
                title: 'Quick Start',
                description: 'From sign-up to first sale in minutes.',
                href: '/docs/quick-start',
                icon: <Zap size={16} />,
              },
              {
                title: 'API Reference',
                description: 'Build integrations with the rtylr REST API.',
                href: '/api',
                icon: <Code size={16} />,
              },
              {
                title: 'Changelog',
                description: 'Everything that shipped, sorted by date.',
                href: '/changelog',
                icon: <Sparkles size={16} />,
              },
            ].map((card) => (
              <Link
                key={card.href}
                to={card.href}
                className="group flex items-start gap-3 p-4 rounded-lg border no-underline transition-all hover:-translate-y-0.5"
                style={{ borderColor: 'var(--docs-border)', background: 'var(--docs-bg)' }}
              >
                <span
                  className="h-8 w-8 rounded-md inline-flex items-center justify-center shrink-0"
                  style={{ background: 'var(--docs-accent-bg)', color: 'var(--docs-accent)' }}
                >
                  {card.icon}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium mb-0.5" style={{ color: 'var(--docs-text)' }}>
                    {card.title}
                  </p>
                  <p className="text-sm" style={{ color: 'var(--docs-text-muted)' }}>
                    {card.description}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </>
  );
};

export default Home;
