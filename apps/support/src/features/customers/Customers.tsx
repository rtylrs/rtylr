import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Users, Search, Mail, Ticket } from 'lucide-react';
import { useDebounce } from '@voxire/utils';
import { useCustomers } from './useCustomers';

const Customers: React.FC = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const { data, isLoading, isError } = useCustomers({ search: debouncedSearch || undefined });

  const customers: any[] = data ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Customers
          </h2>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Customer directory with ticket history
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3 mb-4">
        <div
          className="flex items-center gap-2 flex-1 h-9 px-3 rounded-lg text-sm"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={14} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search customers..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="h-14 rounded-lg animate-pulse"
              style={{ background: 'var(--hub-hover)' }}
            />
          ))}
        </div>
      )}

      {isError && (
        <div
          className="flex items-center justify-center py-16 rounded-xl text-sm"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text-muted)',
          }}
        >
          Failed to load customers. Please try again.
        </div>
      )}

      {!isLoading && !isError && customers.length === 0 && (
        <div
          className="flex flex-col items-center justify-center py-16 rounded-xl text-sm"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text-muted)',
          }}
        >
          <Users size={32} className="mb-3" style={{ color: 'var(--hub-text-dim)' }} />
          <p>No customers found</p>
        </div>
      )}

      {!isLoading && !isError && customers.length > 0 && (
        <div
          className="rounded-xl overflow-hidden"
          style={{ border: '1px solid var(--hub-border)' }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr style={{ background: 'var(--hub-card)' }}>
                <th
                  className="text-left px-4 py-2.5 text-xs font-medium"
                  style={{
                    color: 'var(--hub-text-muted)',
                    borderBottom: '1px solid var(--hub-border)',
                  }}
                >
                  Name
                </th>
                <th
                  className="text-left px-4 py-2.5 text-xs font-medium"
                  style={{
                    color: 'var(--hub-text-muted)',
                    borderBottom: '1px solid var(--hub-border)',
                  }}
                >
                  Email
                </th>
                <th
                  className="text-left px-4 py-2.5 text-xs font-medium"
                  style={{
                    color: 'var(--hub-text-muted)',
                    borderBottom: '1px solid var(--hub-border)',
                  }}
                >
                  Tickets
                </th>
                <th
                  className="text-left px-4 py-2.5 text-xs font-medium"
                  style={{
                    color: 'var(--hub-text-muted)',
                    borderBottom: '1px solid var(--hub-border)',
                  }}
                >
                  Last Contact
                </th>
              </tr>
            </thead>
            <tbody>
              {customers.map((customer: any) => (
                <tr
                  key={customer.id}
                  className="transition-colors"
                  style={{ borderBottom: '1px solid var(--hub-border)' }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2.5">
                      <div
                        className="h-7 w-7 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
                        style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
                      >
                        {customer.name?.charAt(0)?.toUpperCase() || '?'}
                      </div>
                      <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                        {customer.name}
                      </span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div
                      className="flex items-center gap-1.5 text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <Mail size={11} />
                      {customer.email}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div
                      className="flex items-center gap-1.5 text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <Ticket size={11} />
                      {customer.ticketCount ?? 0}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                    {customer.lastContactAt || 'Never'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default Customers;
