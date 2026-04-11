import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, ArrowRight, Book, Code, Layers, Zap } from 'lucide-react';

const featuredGuides = [
  {
    title: 'Introduction',
    description: 'Learn what rtylr is and how it can help your business.',
    href: '/docs/introduction',
    icon: <Book size={18} />,
  },
  {
    title: 'Quick Start',
    description: 'Get up and running in minutes with this step-by-step guide.',
    href: '/docs/quick-start',
    icon: <Zap size={18} />,
  },
  {
    title: 'API Reference',
    description: 'Build integrations with the rtylr REST API.',
    href: '/api',
    icon: <Code size={18} />,
  },
  {
    title: 'Platform Overview',
    description: 'Explore the apps that make up the rtylr platform.',
    href: '/docs/hub',
    icon: <Layers size={18} />,
  },
];

const Home: React.FC = () => {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = React.useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  return (
    <div>
      {/* Hero */}
      <div className="mb-10">
        <h1 className="text-3xl font-bold mb-3" style={{ color: 'var(--docs-text)' }}>
          rtylr Documentation
        </h1>
        <p className="text-base mb-6" style={{ color: 'var(--docs-text-muted)' }}>
          Everything you need to build, configure, and integrate with rtylr.
        </p>

        {/* Search */}
        <form onSubmit={handleSearch} className="max-w-lg">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2"
              style={{ color: 'var(--docs-text-dim)' }} />
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

      {/* Featured Guides */}
      <div className="mb-10">
        <h2 className="text-sm font-semibold uppercase tracking-wider mb-4"
          style={{ color: 'var(--docs-text-muted)' }}>
          Getting Started
        </h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {featuredGuides.map((guide) => (
            <Link
              key={guide.href}
              to={guide.href}
              className="group flex flex-col p-4 rounded-lg border no-underline transition-colors"
              style={{ borderColor: 'var(--docs-border)', background: 'var(--docs-bg)' }}
            >
              <div className="flex items-center gap-2 mb-2">
                <span style={{ color: 'var(--docs-accent)' }}>{guide.icon}</span>
                <span className="text-sm font-medium" style={{ color: 'var(--docs-text)' }}>
                  {guide.title}
                </span>
                <ArrowRight size={14} className="ml-auto opacity-0 group-hover:opacity-100 transition-opacity"
                  style={{ color: 'var(--docs-text-dim)' }} />
              </div>
              <p className="text-sm" style={{ color: 'var(--docs-text-muted)' }}>
                {guide.description}
              </p>
            </Link>
          ))}
        </div>
      </div>

      {/* Quick Links */}
      <div>
        <h2 className="text-sm font-semibold uppercase tracking-wider mb-4"
          style={{ color: 'var(--docs-text-muted)' }}>
          Popular Topics
        </h2>
        <div className="flex flex-col gap-1">
          {[
            { label: 'Authentication', href: '/docs/authentication' },
            { label: 'Team Management', href: '/docs/team-management' },
            { label: 'Webhooks', href: '/docs/webhooks' },
            { label: 'Changelog', href: '/changelog' },
          ].map((link) => (
            <Link
              key={link.href}
              to={link.href}
              className="flex items-center gap-2 px-3 py-2 text-sm rounded-md no-underline transition-colors"
              style={{ color: 'var(--docs-accent)' }}
            >
              <ArrowRight size={12} />
              {link.label}
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Home;
