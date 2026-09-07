import React, { useState, useMemo } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Plug2, Search } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Input } from '../../shared/components/ui';
import {
  useIntegrations,
  useConnectIntegration,
  useDisconnectIntegration,
} from '../../shared/hooks/queries/useIntegrations';

const CATEGORIES = [
  { key: 'all', label: 'All' },
  { key: 'payments', label: 'Payments' },
  { key: 'messaging', label: 'Messaging' },
  { key: 'accounting', label: 'Accounting' },
  { key: 'analytics', label: 'Analytics' },
  { key: 'delivery', label: 'Delivery' },
] as const;

const Integrations: React.FC = () => {
  const { t } = useTranslation();
  const { data = [], isLoading } = useIntegrations();
  const connect = useConnectIntegration();
  const disconnect = useDisconnectIntegration();
  const [searchTerm, setSearchTerm] = useState('');
  const [cat, setCat] = useState<(typeof CATEGORIES)[number]['key']>('all');

  const integrations = useMemo(() => (Array.isArray(data) ? data : []), [data]);

  const filtered = useMemo(() => {
    return integrations.filter((i) => {
      const name = i.name;
      const matchesQuery =
        !searchTerm.trim() || name.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesCat = cat === 'all' || i.category === cat;
      return matchesQuery && matchesCat;
    });
  }, [integrations, searchTerm, cat]);

  return (
    <>
      <HelmetContainer title="Integrations · rtylr" />
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
              Integrations
            </h1>
            <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
              Connect rtylr to the tools you already use
            </p>
          </div>
          <div className="relative">
            <Search
              size={13}
              className="absolute left-3 top-1/2 -translate-y-1/2"
              style={{ color: 'var(--dash-text-dim)' }}
            />
            <Input
              className="pl-9 w-56"
              placeholder="Search integrations..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {/* Category pills */}
        <div className="flex items-center gap-1.5 flex-wrap">
          {CATEGORIES.map((c) => (
            <button
              type="button"
              key={c.key}
              onClick={() => setCat(c.key)}
              className="px-3 py-1 rounded-lg text-xs font-medium transition-colors"
              style={{
                background: cat === c.key ? 'rgba(69,131,196,0.12)' : 'var(--dash-hover)',
                color: cat === c.key ? '#4583C4' : 'var(--dash-text-muted)',
                border:
                  cat === c.key
                    ? '1px solid rgba(69,131,196,0.25)'
                    : '1px solid var(--dash-border)',
              }}
            >
              {c.label}
            </button>
          ))}
        </div>

        {/* Grid */}
        {isLoading ? (
          <div className="grid grid-cols-2 gap-3">
            {[1, 2, 3, 4].map((i) => (
              <div
                key={i}
                className="h-32 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Plug2}
            title={t('integrations.emptyMatchTitle')}
            action={{
              label: t('common.clearFilters'),
              onClick: () => {
                setSearchTerm('');
                setCat('all');
              },
            }}
          />
        ) : (
          <div className="grid grid-cols-2 gap-3">
            {filtered.map((i) => {
              const isConnected = i.status === 'connected';
              const name = i.name;

              return (
                <div
                  key={i.id}
                  className="rounded-xl p-5 flex flex-col gap-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <div className="flex items-start justify-between">
                    <div
                      className="h-10 w-10 rounded-xl flex items-center justify-center shrink-0"
                      style={{ background: 'rgba(69,131,196,0.12)' }}
                    >
                      <Plug2 size={19} style={{ color: '#4583C4' }} />
                    </div>
                    {i.status === 'coming_soon' ? (
                      <span
                        className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                        style={{ background: 'rgba(245,158,11,0.12)', color: '#f59e0b' }}
                      >
                        Coming Soon
                      </span>
                    ) : isConnected ? (
                      <span
                        className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                        style={{ background: 'rgba(16,185,129,0.12)', color: '#10b981' }}
                      >
                        Connected
                      </span>
                    ) : null}
                  </div>

                  <div className="flex-1">
                    <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                      {name}
                    </p>
                    {i.description && (
                      <p
                        className="text-xs mt-1 leading-relaxed"
                        style={{ color: 'var(--dash-text-muted)' }}
                      >
                        {i.description}
                      </p>
                    )}
                  </div>

                  <div className="flex items-center justify-end">
                    {i.status === 'coming_soon' ? (
                      <button
                        type="button"
                        disabled
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold opacity-50 cursor-not-allowed"
                        style={{
                          background: 'var(--dash-hover)',
                          color: 'var(--dash-text-muted)',
                          border: '1px solid var(--dash-border)',
                        }}
                      >
                        Coming Soon
                      </button>
                    ) : isConnected ? (
                      <button
                        type="button"
                        onClick={() => disconnect.mutate(i.id)}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                        style={{
                          background: 'var(--dash-hover)',
                          color: 'var(--dash-text-muted)',
                          border: '1px solid var(--dash-border)',
                        }}
                      >
                        Disconnect
                      </button>
                    ) : (
                      <button
                        type="button"
                        onClick={() => connect.mutate(i.id)}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                        style={{ background: '#4583C4', color: '#fff' }}
                      >
                        Connect
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </>
  );
};

export default Integrations;
