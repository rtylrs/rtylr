import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Ticket, Search, Plus, Circle } from 'lucide-react';
import { useDebounce } from '@voxire/utils';
import { useTickets } from '../../hooks/useTickets';

const statusColors: Record<string, string> = {
  open: '#3b82f6',
  pending: '#f59e0b',
  resolved: '#10b981',
  closed: '#6b7280',
};

const Tickets: React.FC = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const { data, isLoading, isError } = useTickets({
    status: statusFilter || undefined,
    search: debouncedSearch || undefined,
  });

  const tickets: any[] = data ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Tickets
          </h2>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Track and manage support tickets
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
            placeholder="Search tickets..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <div className="flex items-center gap-1">
          {['', 'open', 'pending', 'resolved', 'closed'].map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className="h-8 px-3 rounded-lg text-xs font-medium transition-colors"
              style={{
                background: statusFilter === s ? 'var(--hub-active)' : 'transparent',
                color: statusFilter === s ? 'var(--hub-text)' : 'var(--hub-text-muted)',
                border: statusFilter === s ? '1px solid var(--hub-border)' : '1px solid transparent',
              }}
            >
              {s === '' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
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
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
        >
          Failed to load tickets. Please try again.
        </div>
      )}

      {!isLoading && !isError && tickets.length === 0 && (
        <div
          className="flex flex-col items-center justify-center py-16 rounded-xl text-sm"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
        >
          <Ticket size={32} className="mb-3" style={{ color: 'var(--hub-text-dim)' }} />
          <p>No tickets found</p>
        </div>
      )}

      {!isLoading && !isError && tickets.length > 0 && (
        <div
          className="rounded-xl overflow-hidden"
          style={{ border: '1px solid var(--hub-border)' }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr style={{ background: 'var(--hub-card)' }}>
                <th className="text-left px-4 py-2.5 text-xs font-medium" style={{ color: 'var(--hub-text-muted)', borderBottom: '1px solid var(--hub-border)' }}>ID</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium" style={{ color: 'var(--hub-text-muted)', borderBottom: '1px solid var(--hub-border)' }}>Subject</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium" style={{ color: 'var(--hub-text-muted)', borderBottom: '1px solid var(--hub-border)' }}>Customer</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium" style={{ color: 'var(--hub-text-muted)', borderBottom: '1px solid var(--hub-border)' }}>Status</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium" style={{ color: 'var(--hub-text-muted)', borderBottom: '1px solid var(--hub-border)' }}>Assignee</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium" style={{ color: 'var(--hub-text-muted)', borderBottom: '1px solid var(--hub-border)' }}>Updated</th>
              </tr>
            </thead>
            <tbody>
              {tickets.map((ticket: any) => (
                <tr
                  key={ticket.id}
                  className="transition-colors cursor-pointer"
                  style={{ borderBottom: '1px solid var(--hub-border)' }}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <td className="px-4 py-3">
                    <Link to={`/tickets/${ticket.id}`} className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)', textDecoration: 'none' }}>
                      #{ticket.id}
                    </Link>
                  </td>
                  <td className="px-4 py-3">
                    <Link to={`/tickets/${ticket.id}`} className="text-sm font-medium" style={{ color: 'var(--hub-text)', textDecoration: 'none' }}>
                      {ticket.subject}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {ticket.customerName}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className="inline-flex items-center gap-1.5 text-[11px] font-medium px-2 py-0.5 rounded-full"
                      style={{
                        background: `${statusColors[ticket.status] || '#6b7280'}15`,
                        color: statusColors[ticket.status] || '#6b7280',
                      }}
                    >
                      <Circle size={5} fill="currentColor" stroke="none" />
                      {ticket.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {ticket.assigneeName || 'Unassigned'}
                  </td>
                  <td className="px-4 py-3 text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                    {ticket.updatedAt}
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

export default Tickets;
