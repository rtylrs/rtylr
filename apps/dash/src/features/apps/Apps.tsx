import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Search, Grid3x3 } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Input } from '../../shared/components/ui';
import { AppCard, Breadcrumbs, SkeletonGrid } from '../../shared/components/dash';
import { useApps } from '../../shared/hooks/queries/useApps';
import { useUserStore } from '../../shared/stores/userStore';

const CATEGORIES: {
  key: 'all' | 'core' | 'sales' | 'operations' | 'finance' | 'growth';
  label: string;
}[] = [
  { key: 'all', label: 'All' },
  { key: 'core', label: 'Core' },
  { key: 'sales', label: 'Sales' },
  { key: 'operations', label: 'Operations' },
  { key: 'finance', label: 'Finance' },
  { key: 'growth', label: 'Growth' },
];

const Apps: React.FC = () => {
  const { t } = useTranslation();
  const { data: apps = [], isLoading } = useApps();
  const { user } = useUserStore();
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState<string>('all');

  const userApps: string[] =
    user?.effectiveAccess?.apps
      .filter((app) => app.licensed && app.hasAccess)
      .map((app) => app.appId) ??
    user?.apps ??
    [];
  const isOwner = user?.effectiveAccess
    ? user.effectiveAccess.isOwner || user.effectiveAccess.isAdmin
    : user?.role === 'owner' || user?.role === 'admin';

  const filtered = useMemo(() => {
    return apps
      .filter((a) => {
        // Staff/invited users only see apps they have explicit permission for.
        if (!isOwner && !userApps.includes(a.key)) return false;
        const matchesQuery =
          !query.trim() ||
          a.label.toLowerCase().includes(query.toLowerCase()) ||
          a.description.toLowerCase().includes(query.toLowerCase());
        const matchesCat = category === 'all' || a.category === category;
        return matchesQuery && matchesCat;
      })
      .map((a) => ({
        ...a,
        // Owners see all apps but locked ones link to the upgrade page.
        locked: isOwner && userApps.length > 0 && !userApps.includes(a.key),
      }));
  }, [apps, query, category, isOwner, userApps]);

  return (
    <>
      <HelmetContainer title="Apps · rtylr" />
      <div className="dash-apps-page space-y-5 w-full min-w-0">
        <div className="flex items-center gap-3">
          <div
            className="h-9 w-9 rounded-xl flex items-center justify-center"
            style={{ background: 'rgba(69,131,196,0.15)' }}
          >
            <Grid3x3 size={16} style={{ color: '#4583C4' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              All apps
            </h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              Everything rtylr offers, in one place.
            </p>
          </div>
        </div>

        <div className="dash-apps-controls flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-[12rem] max-w-sm">
            <Search
              size={14}
              className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
              style={{ color: 'var(--dash-text-dim)' }}
            />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search apps..."
              className="pl-9"
            />
          </div>
          <div className="flex items-center gap-1 flex-wrap">
            {CATEGORIES.map((c) => (
              <button
                type="button"
                key={c.key}
                onClick={() => setCategory(c.key)}
                className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
                style={{
                  background: category === c.key ? 'rgba(69,131,196,0.15)' : 'var(--dash-hover)',
                  color: category === c.key ? '#4583C4' : 'var(--dash-text-muted)',
                  border:
                    category === c.key
                      ? '1px solid rgba(69,131,196,0.3)'
                      : '1px solid var(--dash-border)',
                }}
              >
                {c.label}
              </button>
            ))}
          </div>
        </div>

        {isLoading ? (
          <SkeletonGrid cols={3} rows={3} />
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Grid3x3}
            title={t('apps.emptyTitle')}
            description={t('apps.emptyDescription')}
            action={{
              label: t('common.clearFilters'),
              onClick: () => {
                setQuery('');
                setCategory('all');
              },
            }}
          />
        ) : (
          <div className="dash-app-grid grid gap-4">
            {filtered.map((app) => (
              <AppCard key={app.key} app={app} />
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default Apps;
