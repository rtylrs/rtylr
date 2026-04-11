import React, { useState, useMemo } from 'react';
import { Plug2, Search, ExternalLink, Check } from 'lucide-react';
import { Input } from "../../shared/components/ui";
import {
  useIntegrations,
  useConnectIntegration,
  useDisconnectIntegration,
} from "../../shared/hooks/queries/useIntegrations";

const CATEGORIES = [
  { key: 'all',         label: 'All' },
  { key: 'payments',    label: 'Payments' },
  { key: 'messaging',   label: 'Messaging' },
  { key: 'accounting',  label: 'Accounting' },
  { key: 'analytics',   label: 'Analytics' },
  { key: 'delivery',    label: 'Delivery' },
] as const;

const Integrations: React.FC = () => {
  const { data = [], isLoading } = useIntegrations();
  const connect = useConnectIntegration();
  const disconnect = useDisconnectIntegration();
  const [searchTerm, setSearchTerm] = useState('');
  const [cat, setCat] = useState<(typeof CATEGORIES)[number]['key']>('all');

  const integrations = Array.isArray(data) ? data : [];

  const filtered = useMemo(() => {
    return integrations.filter((i: any) => {
      const name = i.name || i.provider || '';
      const matchesQuery = !searchTerm.trim() || name.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesCat = cat === 'all' || i.category === cat;
      return matchesQuery && matchesCat;
    });
  }, [integrations, searchTerm, cat]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--hub-text)' }}>Integrations</h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Connect rtylr to the tools you already use
          </p>
        </div>
        <div className="relative">
          <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--hub-text-dim)' }} />
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
            key={c.key}
            onClick={() => setCat(c.key)}
            className="px-3 py-1 rounded-lg text-xs font-medium transition-colors"
            style={{
              background: cat === c.key ? 'rgba(69,131,196,0.12)' : 'var(--hub-hover)',
              color: cat === c.key ? '#4583C4' : 'var(--hub-text-muted)',
              border: cat === c.key ? '1px solid rgba(69,131,196,0.25)' : '1px solid var(--hub-border)',
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
            <div key={i} className="h-32 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Plug2 size={28} className="mb-2" style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>
            No integrations match your search
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-3">
          {filtered.map((i: any) => {
            const isConnected = i.status === 'connected';
            const name = i.name || i.provider || 'Unknown';

            return (
              <div
                key={i.id || i.provider}
                className="rounded-xl p-5 flex flex-col gap-4"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-start justify-between">
                  <div
                    className="h-10 w-10 rounded-xl flex items-center justify-center shrink-0"
                    style={{ background: 'rgba(69,131,196,0.12)' }}
                  >
                    <Plug2 size={19} style={{ color: '#4583C4' }} />
                  </div>
                  {isConnected && (
                    <span
                      className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                      style={{ background: 'rgba(16,185,129,0.12)', color: '#10b981' }}
                    >
                      Connected
                    </span>
                  )}
                </div>

                <div className="flex-1">
                  <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>{name}</p>
                  {i.description && (
                    <p className="text-xs mt-1 leading-relaxed" style={{ color: 'var(--hub-text-muted)' }}>
                      {i.description}
                    </p>
                  )}
                </div>

                <div className="flex items-center justify-end">
                  {isConnected ? (
                    <button
                      onClick={() => disconnect.mutate(i.id)}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                      style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)', border: '1px solid var(--hub-border)' }}
                    >
                      Disconnect
                    </button>
                  ) : (
                    <button
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
  );
};

export default Integrations;
