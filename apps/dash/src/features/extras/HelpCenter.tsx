import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { LifeBuoy, Search, ExternalLink } from 'lucide-react';
import { Input } from '../../shared/components/ui';

interface Article {
  id: string;
  title: string;
  category: string;
  href: string;
  excerpt: string;
}

const ARTICLES: Article[] = [
  {
    id: '1',
    title: 'Getting started with rtylr Hub',
    category: 'Getting started',
    href: 'https://docs.rtylr.com/getting-started/introduction',
    excerpt: 'Everything you need to know about the Hub.',
  },
  {
    id: '2',
    title: 'Inviting your first team member',
    category: 'Team',
    href: 'https://docs.rtylr.com/getting-started/multi-tenant',
    excerpt: 'Send invitations and assign apps and permissions.',
  },
  {
    id: '3',
    title: 'Connecting Stripe for payments',
    category: 'Billing',
    href: 'https://docs.rtylr.com/faq',
    excerpt: 'Accept online payments with Stripe.',
  },
  {
    id: '4',
    title: 'Configuring permissions per role',
    category: 'Security',
    href: 'https://docs.rtylr.com/getting-started/multi-tenant',
    excerpt: 'Give the right people the right access.',
  },
  {
    id: '5',
    title: 'Building a menu in the Menus app',
    category: 'POS',
    href: 'https://docs.rtylr.com/apps/pos',
    excerpt: 'Create menus and categories.',
  },
  {
    id: '6',
    title: 'Reading the analytics dashboard',
    category: 'Analytics',
    href: 'https://docs.rtylr.com/apps/insights',
    excerpt: 'Understand revenue, orders, and trends.',
  },
  {
    id: '8',
    title: 'Enabling two-factor authentication',
    category: 'Security',
    href: 'https://docs.rtylr.com/faq',
    excerpt: 'Add an extra layer to your account.',
  },
];

const HelpCenter: React.FC = () => {
  const [query, setQuery] = useState('');
  const filtered = useMemo(() => {
    if (!query.trim()) return ARTICLES;
    return ARTICLES.filter(
      (a) =>
        a.title.toLowerCase().includes(query.toLowerCase()) ||
        a.category.toLowerCase().includes(query.toLowerCase()) ||
        a.excerpt.toLowerCase().includes(query.toLowerCase())
    );
  }, [query]);

  return (
    <>
      <HelmetContainer title="Help center · rtylr" />
      <div className="space-y-5 max-w-3xl">
        <div className="flex items-center gap-3">
          <div
            className="h-9 w-9 rounded-xl flex items-center justify-center"
            style={{ background: 'rgba(100,197,185,0.15)' }}
          >
            <LifeBuoy size={16} style={{ color: '#64C5B9' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              Help center
            </h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              Guides and answers to common questions.
            </p>
          </div>
        </div>

        <div className="relative">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
            style={{ color: 'var(--dash-text-dim)' }}
          />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search help articles..."
            className="pl-9"
          />
        </div>

        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {filtered.map((a) => (
            <a
              key={a.id}
              href={a.href}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-start gap-3 px-4 py-3 no-underline transition-colors hover:bg-[var(--dash-hover)]"
              style={{ borderBottom: '1px solid var(--dash-border)' }}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {a.title}
                  </p>
                  <span
                    className="text-[10px] font-semibold uppercase tracking-wider"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    {a.category}
                  </span>
                </div>
                <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
                  {a.excerpt}
                </p>
              </div>
              <ExternalLink
                size={12}
                className="mt-1 shrink-0"
                style={{ color: 'var(--dash-text-dim)' }}
              />
            </a>
          ))}
        </div>
      </div>
    </>
  );
};

export default HelpCenter;
