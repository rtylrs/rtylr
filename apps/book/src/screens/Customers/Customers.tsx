import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Users, Search, Calendar, Phone, Mail } from 'lucide-react';
import { useCustomers } from '../../hooks/useCustomers';

const SkeletonRow: React.FC = () => (
  <div className="h-12 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

const Customers: React.FC = () => {
  const [search, setSearch] = useState('');

  const { data: customers, isLoading, isError } = useCustomers({
    search: search || undefined,
  });

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-7xl"
    >
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Customers
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Booking history and customer preferences
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <div
          className="flex items-center gap-2 flex-1 max-w-sm px-3 py-1.5 rounded-lg"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={13} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search customers..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-xs"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
      </div>

      {isError && (
        <div
          className="rounded-xl p-4 text-sm"
          style={{ background: 'rgba(239,68,68,0.08)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.2)' }}
        >
          Failed to load customers. Please try again later.
        </div>
      )}

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid grid-cols-6 px-4 py-2 text-[10px] font-semibold uppercase tracking-wider"
          style={{ color: 'var(--hub-text-dim)', borderBottom: '1px solid var(--hub-border)' }}
        >
          <span>Customer</span>
          <span>Email</span>
          <span>Phone</span>
          <span>Total Bookings</span>
          <span>Last Visit</span>
          <span className="text-right">Preferences</span>
        </div>

        {isLoading && (
          <div className="p-3 space-y-2">
            {Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        )}

        {!isLoading && !isError && (!customers || customers.length === 0) && (
          <div className="py-12 text-center">
            <Users size={32} className="mx-auto mb-2" style={{ color: 'var(--hub-text-dim)' }} />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No customers found
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              {search ? 'Try a different search term' : 'Customers will appear after their first booking'}
            </p>
          </div>
        )}

        {!isLoading && !isError && customers && customers.length > 0 && (
          <div>
            {customers.map((customer: any) => (
              <div
                key={customer.id}
                className="grid grid-cols-6 items-center px-4 py-3 transition-colors"
                style={{ borderBottom: '1px solid var(--hub-border)' }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
              >
                <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{customer.name}</p>
                <div className="flex items-center gap-1">
                  <Mail size={10} style={{ color: 'var(--hub-text-dim)' }} />
                  <span className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>{customer.email ?? '-'}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Phone size={10} style={{ color: 'var(--hub-text-dim)' }} />
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{customer.phone ?? '-'}</span>
                </div>
                <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{customer.totalBookings ?? 0}</p>
                <div className="flex items-center gap-1">
                  <Calendar size={10} style={{ color: 'var(--hub-text-dim)' }} />
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{customer.lastVisit ?? '-'}</span>
                </div>
                <p className="text-xs text-right truncate" style={{ color: 'var(--hub-text-muted)' }}>
                  {customer.preferences ?? '-'}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>
    </motion.div>
  );
};

export default Customers;
