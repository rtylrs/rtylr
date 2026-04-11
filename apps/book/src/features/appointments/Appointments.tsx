import React, { useState } from 'react';
import { motion } from 'motion/react';
import { ClipboardList, Search, Filter, Calendar } from 'lucide-react';
import { useAppointments } from './useAppointments';

const INDIGO = '#6366f1';

const STATUS_OPTIONS = ['all', 'confirmed', 'pending', 'cancelled', 'completed', 'no-show'];

function getStatusColor(status: string): string {
  if (status === 'confirmed') return '#10b981';
  if (status === 'pending') return '#f59e0b';
  if (status === 'cancelled') return '#ef4444';
  if (status === 'completed') return INDIGO;
  if (status === 'no-show') return '#a78bfa';
  return 'var(--hub-text-dim)';
}

const SkeletonRow: React.FC = () => (
  <div className="h-12 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

const Appointments: React.FC = () => {
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  const { data: appointments, isLoading, isError } = useAppointments({
    search: search || undefined,
    status: statusFilter !== 'all' ? statusFilter : undefined,
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
            Appointments
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Manage all bookings and appointments
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
            placeholder="Search appointments..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-xs"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <div className="flex items-center gap-1">
          <Filter size={12} style={{ color: 'var(--hub-text-dim)' }} />
          {STATUS_OPTIONS.map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className="px-2 py-1 rounded-md text-[11px] font-medium capitalize"
              style={{
                background: statusFilter === s ? 'rgba(99,102,241,0.12)' : 'transparent',
                color: statusFilter === s ? INDIGO : 'var(--hub-text-muted)',
              }}
            >
              {s}
            </button>
          ))}
        </div>
      </div>

      {isError && (
        <div
          className="rounded-xl p-4 text-sm"
          style={{ background: 'rgba(239,68,68,0.08)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.2)' }}
        >
          Failed to load appointments. Please try again later.
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
          <span>Date / Time</span>
          <span>Customer</span>
          <span>Service</span>
          <span>Staff</span>
          <span>Duration</span>
          <span className="text-right">Status</span>
        </div>

        {isLoading && (
          <div className="p-3 space-y-2">
            {Array.from({ length: 6 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        )}

        {!isLoading && !isError && (!appointments || appointments.length === 0) && (
          <div className="py-12 text-center">
            <Calendar size={32} className="mx-auto mb-2" style={{ color: 'var(--hub-text-dim)' }} />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No appointments found
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              {search || statusFilter !== 'all' ? 'Try adjusting your filters' : 'Create your first appointment to get started'}
            </p>
          </div>
        )}

        {!isLoading && !isError && appointments && appointments.length > 0 && (
          <div>
            {appointments.map((appt: any) => {
              const statusColor = getStatusColor(appt.status);
              return (
                <div
                  key={appt.id}
                  className="grid grid-cols-6 items-center px-4 py-3 transition-colors"
                  style={{ borderBottom: '1px solid var(--hub-border)' }}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <div>
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{appt.date}</p>
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>{appt.time}</p>
                  </div>
                  <p className="text-xs truncate" style={{ color: 'var(--hub-text)' }}>{appt.customerName}</p>
                  <p className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>{appt.serviceName}</p>
                  <p className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>{appt.staffName}</p>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{appt.duration} min</p>
                  <div className="flex justify-end">
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize"
                      style={{ background: `${statusColor}18`, color: statusColor }}
                    >
                      {appt.status}
                    </span>
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

export default Appointments;
