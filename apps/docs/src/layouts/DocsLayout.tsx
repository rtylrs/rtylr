import React, { useState } from 'react';
import { Outlet, Link, useLocation, useNavigate } from '@tanstack/react-router';
import { Search, Book, Layers, Users, Code, FileText, Server, ChevronRight } from 'lucide-react';
import { content } from '../shared/content';

interface NavItem {
  label: string;
  href: string;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

// Filter out any nav entry whose slug isn't in the content registry,
// keeping `/docs/:slug` paths from ever 404'ing. The hash-routed
// non-doc pages (`/api`, `/changelog`) are passed through unchecked.
// See vault audit docs.md P1: sidebar drift used to require manual
// sync - this filter flips it to fail-closed at render time.
function pruneToContent(items: NavItem[]): NavItem[] {
  return items.filter((it) => {
    if (!it.href.startsWith('/docs/')) return true;
    const slug = it.href.slice('/docs/'.length);
    return slug in content;
  });
}

const rawSections: NavSection[] = [
  {
    title: 'Getting Started',
    items: [
      { label: 'Introduction', href: '/docs/introduction' },
      { label: 'Quick Start', href: '/docs/quick-start' },
      { label: 'Architecture', href: '/docs/architecture' },
    ],
  },
  {
    title: 'Apps',
    items: [
      { label: 'Hub', href: '/docs/hub' },
      { label: 'POS', href: '/docs/pos' },
      { label: 'ERP', href: '/docs/erp' },
      { label: 'HR', href: '/docs/hr' },
      { label: 'CRM', href: '/docs/crm' },
      { label: 'Finance', href: '/docs/finance' },
      { label: 'Flow', href: '/docs/flow' },
      { label: 'Store', href: '/docs/store' },
      { label: 'Book', href: '/docs/book' },
      { label: 'Support', href: '/docs/support' },
      { label: 'Insights', href: '/docs/insights' },
    ],
  },
  {
    title: 'Platform',
    items: [
      { label: 'Multi-Tenant', href: '/docs/multi-tenant' },
      { label: 'Addon System', href: '/docs/addons' },
      { label: 'Authentication', href: '/docs/authentication' },
      { label: 'Team Management', href: '/docs/team-management' },
    ],
  },
  {
    title: 'Self-Hosting',
    items: [
      { label: 'Overview', href: '/docs/self-hosting' },
      { label: 'License Flow', href: '/docs/self-hosting-license' },
      { label: 'Image Contract', href: '/docs/self-hosting-images' },
      { label: 'Helm Install', href: '/docs/self-hosting-helm' },
      { label: 'Docker Compose', href: '/docs/self-hosting-compose' },
      { label: 'DNS Setup', href: '/docs/self-hosting-dns' },
      { label: 'Backend Env', href: '/docs/self-hosting-env' },
      { label: 'Upgrade Flow', href: '/docs/self-hosting-upgrade' },
    ],
  },
  {
    title: 'API',
    items: [
      { label: 'Overview', href: '/api' },
      { label: 'Endpoints', href: '/docs/api-endpoints' },
      { label: 'Webhooks', href: '/docs/webhooks' },
      { label: 'Rate Limits', href: '/docs/rate-limits' },
    ],
  },
  {
    title: 'Resources',
    items: [
      { label: 'FAQ', href: '/docs/faq' },
      { label: 'Changelog', href: '/changelog' },
      { label: 'Contact', href: '/docs/contact' },
    ],
  },
];

const sections: NavSection[] = rawSections
  .map((s) => ({ ...s, items: pruneToContent(s.items) }))
  .filter((s) => s.items.length > 0);

const sectionIcons: Record<string, React.ReactNode> = {
  'Getting Started': <Book size={14} />,
  Apps: <Layers size={14} />,
  Platform: <Users size={14} />,
  'Self-Hosting': <Server size={14} />,
  API: <Code size={14} />,
  Resources: <FileText size={14} />,
};

const DocsLayout: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate({ to: '/search', search: { q: searchQuery.trim() } });
    }
  };

  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <aside
        className="w-64 border-r flex flex-col flex-shrink-0 h-full overflow-hidden"
        style={{ borderColor: 'var(--docs-border)', background: 'var(--docs-surface)' }}
      >
        {/* Logo */}
        <div className="px-5 py-4 border-b" style={{ borderColor: 'var(--docs-border)' }}>
          <a
            href="https://rtylr.com"
            className="flex items-center gap-2 no-underline"
            target="_blank"
            rel="noopener noreferrer"
          >
            <span className="text-base font-semibold" style={{ color: 'var(--docs-text)' }}>
              rtylr
            </span>
            <span
              className="text-xs font-medium px-1.5 py-0.5 rounded"
              style={{ background: 'var(--docs-accent-bg)', color: 'var(--docs-accent)' }}
            >
              docs
            </span>
          </a>
        </div>

        {/* Search */}
        <div className="px-3 py-3">
          <form onSubmit={handleSearch}>
            <div className="relative">
              <Search
                size={14}
                className="absolute left-2.5 top-1/2 -translate-y-1/2"
                style={{ color: 'var(--docs-text-dim)' }}
              />
              <input
                type="text"
                placeholder="Search docs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-8 pr-3 py-1.5 text-sm rounded-md border outline-none transition-colors"
                style={{
                  borderColor: 'var(--docs-border)',
                  background: 'var(--docs-bg)',
                  color: 'var(--docs-text)',
                }}
              />
            </div>
          </form>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto px-3 pb-4">
          {sections.map((section) => (
            <div key={section.title} className="mb-4">
              <div className="flex items-center gap-1.5 px-2 mb-1">
                <span style={{ color: 'var(--docs-text-muted)' }}>
                  {sectionIcons[section.title]}
                </span>
                <span
                  className="text-xs font-semibold uppercase tracking-wider"
                  style={{ color: 'var(--docs-text-muted)' }}
                >
                  {section.title}
                </span>
              </div>
              {section.items.map((item) => {
                const isActive = location.pathname === item.href;
                return (
                  <Link
                    key={item.href}
                    to={item.href}
                    className="flex items-center gap-1.5 px-2 py-1 text-sm rounded-md no-underline transition-colors"
                    style={{
                      color: isActive ? 'var(--docs-accent)' : 'var(--docs-text-muted)',
                      background: isActive ? 'var(--docs-accent-bg)' : 'transparent',
                      fontWeight: isActive ? 500 : 400,
                    }}
                  >
                    {isActive && <ChevronRight size={12} />}
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-3xl mx-auto px-8 py-10">
          <Outlet />
        </div>
      </main>
    </div>
  );
};

export default DocsLayout;
