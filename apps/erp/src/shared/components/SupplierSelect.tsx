import React, { useState, useMemo } from 'react';
import { Search, Check } from 'lucide-react';
import { useSuppliers } from '../../features/suppliers/useSuppliers';

interface SupplierSelectProps {
  value?: string | null;
  onChange: (id: string) => void;
  placeholder?: string;
}

const SupplierSelect: React.FC<SupplierSelectProps> = ({
  value,
  onChange,
  placeholder = 'Select supplier…',
}) => {
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState('');
  const { data: supplierResult } = useSuppliers();
  const suppliers = useMemo(() => supplierResult?.suppliers ?? [], [supplierResult]);

  const filtered = useMemo(
    () => suppliers.filter((s) => (s.name ?? '').toLowerCase().includes(q.toLowerCase())),
    [suppliers, q]
  );

  const selected = suppliers.find((s) => s.id === value);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-full px-3 py-2 rounded-lg text-sm text-left flex items-center justify-between"
        style={{
          background: 'var(--hub-input-bg)',
          border: '1px solid var(--hub-border)',
          color: selected ? 'var(--hub-text)' : 'var(--hub-text-dim)',
        }}
      >
        <span className="truncate">{selected ? selected.name : placeholder}</span>
      </button>
      {open && (
        <div
          className="absolute top-full left-0 right-0 mt-1 rounded-xl p-2 z-20 max-h-64 overflow-y-auto"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            boxShadow: 'var(--hub-shadow)',
          }}
        >
          <div className="relative mb-2">
            <Search
              size={12}
              className="absolute left-2 top-1/2 -translate-y-1/2"
              style={{ color: 'var(--hub-text-dim)' }}
            />
            <input
              autoFocus
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Search suppliers…"
              className="w-full pl-7 pr-2 py-1.5 rounded-lg text-xs outline-none"
              style={{
                background: 'var(--hub-hover)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
          </div>
          {filtered.map((s) => (
            <button
              type="button"
              key={s.id}
              onClick={() => {
                onChange(s.id ?? '');
                setOpen(false);
                setQ('');
              }}
              className="w-full flex items-center justify-between px-2 py-1.5 rounded-lg text-sm text-left"
              style={{ color: 'var(--hub-text)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <span className="truncate">{s.name}</span>
              {value === s.id && <Check size={12} style={{ color: 'var(--erp-accent)' }} />}
            </button>
          ))}
          {filtered.length === 0 && (
            <p className="text-xs text-center py-3" style={{ color: 'var(--hub-text-dim)' }}>
              No suppliers found
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default SupplierSelect;
