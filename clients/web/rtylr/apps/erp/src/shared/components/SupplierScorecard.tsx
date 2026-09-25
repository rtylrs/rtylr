import React from 'react';
import type { Supplier } from '../../features/suppliers/useSuppliers';
import { Mail, MapPin, ReceiptText, Phone } from 'lucide-react';

interface SupplierScorecardProps {
  supplier: Supplier;
}

const Detail: React.FC<{
  label: string;
  value: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}> = ({ label, value, icon: Icon }) => (
  <div className="min-w-0">
    <div className="mb-1 flex items-center gap-1.5">
      <Icon size={12} style={{ color: 'var(--erp-accent)' }} />
      <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
        {label}
      </p>
    </div>
    <p className="truncate text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
      {value || '-'}
    </p>
  </div>
);

const SupplierScorecard: React.FC<SupplierScorecardProps> = ({ supplier }) => {
  const location = [supplier.city, supplier.country].filter(Boolean).join(', ');

  return (
    <div
      className="rounded-xl p-4"
      style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
    >
      <div className="mb-3 flex items-center justify-between gap-3">
        <p className="truncate text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
          {supplier.name}
        </p>
        <span
          className="shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium"
          style={{
            background: supplier.isActive
              ? 'color-mix(in srgb, var(--erp-success) 12%, transparent)'
              : 'var(--hub-hover)',
            color: supplier.isActive ? 'var(--erp-success)' : 'var(--hub-text-muted)',
          }}
        >
          {supplier.isActive ? 'Active' : 'Inactive'}
        </span>
      </div>
      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        <Detail label="Email" value={supplier.email ?? ''} icon={Mail} />
        <Detail label="Phone" value={supplier.phone ?? ''} icon={Phone} />
        <Detail label="Location" value={location} icon={MapPin} />
        <Detail label="Terms" value={supplier.paymentTerms ?? ''} icon={ReceiptText} />
      </div>
    </div>
  );
};

export default SupplierScorecard;
