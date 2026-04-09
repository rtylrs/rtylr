import React, { useMemo, useState } from 'react';
import { ChevronDown, Search } from 'lucide-react';

export interface Client {
  id: string;
  name: string;
  email?: string;
  company?: string;
}

// Mock client list. In a real app this reads from the CRM service.
const MOCK_CLIENTS: Client[] = [
  { id: 'c1', name: 'Acme Corporation', email: 'billing@acme.com', company: 'Acme' },
  { id: 'c2', name: 'Globex Industries', email: 'ap@globex.com', company: 'Globex' },
  { id: 'c3', name: 'Initech Ltd.', email: 'finance@initech.com', company: 'Initech' },
  { id: 'c4', name: 'Umbrella LLC', email: 'pay@umbrella.com', company: 'Umbrella' },
  { id: 'c5', name: 'Wayne Enterprises', email: 'ar@wayne.com', company: 'Wayne' },
  { id: 'c6', name: 'Stark Industries', email: 'billing@stark.com', company: 'Stark' },
  { id: 'c7', name: 'Wonka Inc', email: 'candy@wonka.com', company: 'Wonka' },
  { id: 'c8', name: 'Pied Piper', email: 'richard@pp.com', company: 'Pied Piper' },
];

interface ClientSelectProps {
  value: Client | null;
  onChange: (client: Client | null) => void;
}

const ClientSelect: React.FC<ClientSelectProps> = ({ value, onChange }) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const filtered = useMemo(
    () => MOCK_CLIENTS.filter((c) => c.name.toLowerCase().includes(query.toLowerCase())),
    [query]
  );

  return (
    <div style={{ position: 'relative', width: '100%' }}>
      <button
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
            left: 0,
            right: 0,
            marginTop: 4,
            background: 'var(--hub-surface)',
            border: '1px solid var(--hub-border)',
            borderRadius: 8,
            padding: 8,
            zIndex: 20,
            maxHeight: 260,
            overflowY: 'auto',
            boxShadow: '0 10px 28px rgba(0,0,0,0.3)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '6px 8px', marginBottom: 6, background: 'var(--hub-hover)', borderRadius: 6 }}>
            <Search size={13} style={{ color: 'var(--hub-text-dim)' }} />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search clients"
              style={{ background: 'transparent', border: 'none', outline: 'none', color: 'var(--hub-text)', fontSize: 12, flex: 1 }}
            />
          </div>
          {filtered.map((c) => (
            <button
              key={c.id}
              onClick={() => {
                onChange(c);
                setOpen(false);
                setQuery('');
              }}
              style={{
                width: '100%',
                textAlign: 'left',
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
          {filtered.length === 0 && (
            <div style={{ padding: 10, textAlign: 'center', fontSize: 12, color: 'var(--hub-text-dim)' }}>
              No clients found
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ClientSelect;
export { MOCK_CLIENTS };
