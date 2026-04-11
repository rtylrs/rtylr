import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { ListOrdered, Search, Bell, Clock } from 'lucide-react';
import { useWaitlist, useNotifyWaitlist } from './useWaitlist';
import { toast } from 'sonner';

const INDIGO = '#6366f1';

const SkeletonRow: React.FC = () => (
  <div className="h-12 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

function getStatusColor(status: string): string {
  if (status === 'waiting') return '#f59e0b';
  if (status === 'notified') return '#60a5fa';
  if (status === 'booked') return '#10b981';
  if (status === 'expired') return '#ef4444';
  return 'var(--hub-text-dim)';
}

const Waitlist: React.FC = () => {
  const [search, setSearch] = useState('');
  const { data: waitlist, isLoading, isError } = useWaitlist({
    search: search || undefined,
  });

  const notifyMutation = useNotifyWaitlist();

  const handleNotify = (id: number) => {
    notifyMutation.mutate(id, {
      onSuccess: () => toast.success('Customer notified'),
      onError: () => toast.error('Failed to notify customer'),
    });
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-7xl"
    >
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Waitlist
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Customers waiting for availability
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
            placeholder="Search waitlist..."
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
          Failed to load waitlist. Please try again later.
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
          <span>Service</span>
          <span>Preferred Time</span>
          <span>Added</span>
          <span>Status</span>
          <span className="text-right">Actions</span>
        </div>

        {isLoading && (
          <div className="p-3 space-y-2">
            {Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        )}

        {!isLoading && !isError && (!waitlist || waitlist.length === 0) && (
          <div className="py-12 text-center">
            <ListOrdered size={32} className="mx-auto mb-2" style={{ color: 'var(--hub-text-dim)' }} />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No one on the waitlist
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              Customers will appear here when slots are full
            </p>
          </div>
        )}

        {!isLoading && !isError && waitlist && waitlist.length > 0 && (
          <div>
            {waitlist.map((entry: any) => {
              const statusColor = getStatusColor(entry.status);
              return (
                <div
                  key={entry.id}
                  className="grid grid-cols-6 items-center px-4 py-3 transition-colors"
                  style={{ borderBottom: '1px solid var(--hub-border)' }}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <div>
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{entry.customerName}</p>
                    {entry.email && <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>{entry.email}</p>}
                  </div>
                  <p className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>{entry.serviceName}</p>
                  <div className="flex items-center gap-1">
                    <Clock size={11} style={{ color: 'var(--hub-text-dim)' }} />
                    <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{entry.preferredTime ?? '-'}</span>
                  </div>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{entry.addedAt ?? entry.createdAt}</p>
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize w-fit"
                    style={{ background: `${statusColor}18`, color: statusColor }}
                  >
                    {entry.status}
                  </span>
                  <div className="flex justify-end">
                    {entry.status === 'waiting' && (
                      <button
                        onClick={() => handleNotify(entry.id)}
                        disabled={notifyMutation.isPending}
                        className="flex items-center gap-1 px-2 py-1 rounded-md text-[10px] font-medium transition-colors"
                        style={{ background: 'rgba(99,102,241,0.1)', color: INDIGO }}
                      >
                        <Bell size={10} />
                        Notify
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </motion.div>
  );
};

export default Waitlist;
