import React, { useCallback, useMemo, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, Search } from 'lucide-react';
import { unwrap } from '@voxire/api';
import { listCrmCustomers } from '../lib/clients';
import { parseClientsResponse } from '../schemas/clients';
import { ACCENT } from '../lib/tokens';

const statusStyle: React.CSSProperties = {
  padding: 10,
  textAlign: 'center',
  fontSize: 12,
  color: 'var(--hub-text-dim)',
};

export interface Client {
  /** CRM customer id when this client came from the CRM directory; absent
   *  for free-typed or cloned clients, which must not claim a customer link
   *  (#840). */
  crmCustomerId?: string;
  id: string;
  name: string;
  email?: string | null;
  company?: string | null;
}

interface ClientSelectProps {
  value: Client | null;
  onChange: (client: Client | null) => void;
}

const ClientSelect: React.FC<ClientSelectProps> = ({ value, onChange }) => {
  const { t } = useTranslation('financeCommon');
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [clients, setClients] = useState<Client[]>([]);
  // Distinguishing a failed load from an empty CRM: the picker used to render
  // 'No clients found' for loading, failure and truly-empty alike, which read
  // as a definitive empty directory and blocked NewInvoice step 1 with no
  // retry path.
  const [loadState, setLoadState] = useState<'loading' | 'error' | 'ready'>('loading');

  const loadClients = useCallback(() => {
    setLoadState('loading');
    void unwrap(listCrmCustomers())
      .then((data) => {
        setClients(parseClientsResponse(data));
        setLoadState('ready');
      })
      .catch(() => setLoadState('error'));
  }, []);

  useEffect(() => {
    loadClients();
  }, [loadClients]);

  const filtered = useMemo(
    () => clients.filter((c) => c.name.toLowerCase().includes(query.toLowerCase())),
    [query, clients]
  );

  return (
    <div style={{ position: 'relative', width: '100%' }}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        style={{
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '8px 12px',
          borderRadius: 8,
          border: '1px solid var(--hub-border)',
          background: 'var(--hub-input-bg)',
          color: 'var(--hub-text)',
          fontSize: 13,
          cursor: 'pointer',
        }}
      >
        <span>{value ? value.name : 'Select client'}</span>
        <ChevronDown size={14} />
      </button>
      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            insetInlineStart: 0,
            insetInlineEnd: 0,
            marginTop: 4,
            background: 'var(--hub-surface)',
            border: '1px solid var(--hub-border)',
            borderRadius: 8,
            padding: 8,
            zIndex: 20,
            maxHeight: 260,
            overflowY: 'auto',
            boxShadow: '0 10px 28px var(--finance-overlay-light)',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 8px',
              marginBottom: 6,
              background: 'var(--hub-hover)',
              borderRadius: 6,
            }}
          >
            <Search size={13} style={{ color: 'var(--hub-text-dim)' }} />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t('finance.clients.search')}
              style={{
                background: 'transparent',
                border: 'none',
                outline: 'none',
                color: 'var(--hub-text)',
                fontSize: 12,
                flex: 1,
              }}
            />
          </div>
          {filtered.map((c) => (
            <button
              type="button"
              key={c.id}
              onClick={() => {
                // Preserve the CRM id so invoices can carry a verified
                // customer link (#840). The generic id stays untouched for
                // existing consumers of this component.
                onChange({ ...c, crmCustomerId: c.id });
                setOpen(false);
                setQuery('');
              }}
              style={{
                width: '100%',
                textAlign: 'start',
                background: 'transparent',
                border: 'none',
                padding: '8px 10px',
                borderRadius: 6,
                cursor: 'pointer',
                color: 'var(--hub-text)',
              }}
            >
              <div style={{ fontSize: 13, fontWeight: 600 }}>{c.name}</div>
              <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{c.email}</div>
            </button>
          ))}
          {loadState === 'loading' && <div style={statusStyle}>{t('financeCommon.loading')}</div>}
          {loadState === 'error' && (
            <div role="alert" style={statusStyle}>
              <div>{t('finance.clients.loadError')}</div>
              <button
                type="button"
                onClick={loadClients}
                style={{
                  marginTop: 6,
                  background: 'transparent',
                  border: 'none',
                  padding: 0,
                  color: ACCENT,
                  fontSize: 12,
                  fontWeight: 600,
                  cursor: 'pointer',
                  textDecoration: 'underline',
                }}
              >
                {t('finance.clients.retry')}
              </button>
            </div>
          )}
          {loadState === 'ready' && filtered.length === 0 && (
            <div style={statusStyle}>{t('finance.clients.none')}</div>
          )}
        </div>
      )}
    </div>
  );
};

export default ClientSelect;
