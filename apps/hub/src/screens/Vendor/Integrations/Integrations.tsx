import React, { useMemo, useState } from 'react';
import { Plug2, Search } from 'lucide-react';
import { Input } from '../../../components/ui';
import { Breadcrumbs, EmptyState, IntegrationCard, SkeletonGrid } from '../../../components/hub';
import {
  useIntegrations,
  useConnectIntegration,
  useDisconnectIntegration,
} from '../../../hooks/queries/useIntegrations';

const CATEGORIES = [
  { key: 'all',         label: 'All' },
  { key: 'payments',    label: 'Payments' },
  { key: 'messaging',   label: 'Messaging' },
  { key: 'accounting',  label: 'Accounting' },
  { key: 'analytics',   label: 'Analytics' },
  { key: 'delivery',    label: 'Delivery' },
  { key: 'data',        label: 'Data' },
] as const;

const Integrations: React.FC = () => {
  const { data = [], isLoading } = useIntegrations();
  const connect = useConnectIntegration();
  const disconnect = useDisconnectIntegration();
  const [query, setQuery] = useState('');
  const [cat, setCat] = useState<(typeof CATEGORIES)[number]['key']>('all');

  const filtered = useMemo(() => {
    return data.filter((i) => {
      const matchesQuery = !query.trim() || i.name.toLowerCase().includes(query.toLowerCase()) || i.description.toLowerCase().includes(query.toLowerCase());
      const matchesCat = cat === 'all' || i.category === cat;
      return matchesQuery && matchesCat;
    });
  }, [data, query, cat]);

  return (
    <div className="space-y-5 max-w-5xl mx-auto">
      <Breadcrumbs />
      <div className="flex items-center gap-3">
        <div className="h-9 w-9 rounded-xl flex items-center justify-center" style={{ background: 'rgba(69,131,196,0.15)' }}>
          <Plug2 size={16} style={{ color: '#4583C4' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Integrations</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            Connect rtylr to the rest of your stack.
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" style={{ color: 'var(--hub-text-dim)' }} />
          <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search integrations..." className="pl-9" />
        </div>
        <div className="flex items-center gap-1 flex-wrap">
          {CATEGORIES.map((c) => (
            <button
              key={c.key}
              onClick={() => setCat(c.key)}
              className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
              style={{
                background: cat === c.key ? 'rgba(69,131,196,0.15)' : 'var(--hub-hover)',
                color:      cat === c.key ? '#4583C4' : 'var(--hub-text-muted)',
                border:     cat === c.key ? '1px solid rgba(69,131,196,0.3)' : '1px solid var(--hub-border)',
              }}
            >
              {c.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <SkeletonGrid cols={3} rows={2} />
      ) : filtered.length === 0 ? (
        <EmptyState icon={Plug2} title="No integrations found" description="Try a different search or category." />
      ) : (
        <div className="grid grid-cols-3 gap-4">
          {filtered.map((i) => (
            <IntegrationCard
              key={i.id}
              integration={i}
              onConnect={(id) => connect.mutate(id)}
              onDisconnect={(id) => disconnect.mutate(id)}
              isLoading={connect.isPending || disconnect.isPending}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default Integrations;
