import React, { useMemo, useState } from 'react';
import { Activity as ActivityIcon, Search } from 'lucide-react';
import { Input } from '../../shared/components/ui';
import {
  ActivityEntryRow,
  Breadcrumbs,
  EmptyState,
  SkeletonList,
} from '../../shared/components/dash';
import { useActivity } from '../../shared/hooks/queries/useActivity';

const Activity: React.FC = () => {
  const { data = [], isLoading } = useActivity(200);
  const [query, setQuery] = useState('');
  const [appFilter, setAppFilter] = useState('all');

  const filtered = useMemo(() => {
    return data.filter((e) => {
      const matchesQuery =
        !query.trim() ||
        e.actor.toLowerCase().includes(query.toLowerCase()) ||
        e.action.toLowerCase().includes(query.toLowerCase()) ||
        e.target.toLowerCase().includes(query.toLowerCase());
      const matchesApp = appFilter === 'all' || e.app === appFilter;
      return matchesQuery && matchesApp;
    });
  }, [data, query, appFilter]);

  return (
    <div className="space-y-5 max-w-5xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(69,131,196,0.15)' }}
        >
          <ActivityIcon size={16} style={{ color: '#4583C4' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Activity
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            A live stream of every action across your rtylr apps.
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
            style={{ color: 'var(--dash-text-dim)' }}
          />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search activity..."
            className="pl-9"
          />
        </div>
        <div className="flex items-center gap-1">
          {['all', 'dash', 'pos', 'erp', 'hr', 'crm', 'finance'].map((a) => (
            <button
              key={a}
              onClick={() => setAppFilter(a)}
              className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
              style={{
                background: appFilter === a ? 'rgba(69,131,196,0.15)' : 'var(--dash-hover)',
                color: appFilter === a ? '#4583C4' : 'var(--dash-text-muted)',
                border:
                  appFilter === a
                    ? '1px solid rgba(69,131,196,0.3)'
                    : '1px solid var(--dash-border)',
              }}
            >
              {a}
            </button>
          ))}
        </div>
      </div>

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        {isLoading ? (
          <div className="p-2">
            <SkeletonList rows={8} />
          </div>
        ) : filtered.length === 0 ? (
          <div className="p-4">
            <EmptyState
              icon={ActivityIcon}
              title="No activity"
              description="Actions will appear here as your team uses rtylr."
            />
          </div>
        ) : (
          <div className="p-2">
            {filtered.map((e) => (
              <ActivityEntryRow key={e.id} entry={e} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Activity;
