import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Briefcase, Search, Clock, DollarSign, Users } from 'lucide-react';
import { useServices } from './useServices';

const INDIGO = '#6366f1';

const SkeletonRow: React.FC = () => (
  <div className="h-12 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

const Services: React.FC = () => {
  const [search, setSearch] = useState('');

  const {
    data: services,
    isLoading,
    isError,
  } = useServices({
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
            Services
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Service catalog with pricing and staff assignments
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
            placeholder="Search services..."
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
          style={{
            background: 'rgba(239,68,68,0.08)',
            color: '#ef4444',
            border: '1px solid rgba(239,68,68,0.2)',
          }}
        >
          Failed to load services. Please try again later.
        </div>
      )}

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid grid-cols-5 px-4 py-2 text-[10px] font-semibold uppercase tracking-wider"
          style={{ color: 'var(--hub-text-dim)', borderBottom: '1px solid var(--hub-border)' }}
        >
          <span>Service Name</span>
          <span>Duration</span>
          <span>Price</span>
          <span>Assigned Staff</span>
          <span className="text-right">Status</span>
        </div>

        {isLoading && (
          <div className="p-3 space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} />
            ))}
          </div>
        )}

        {!isLoading && !isError && (!services || services.length === 0) && (
          <div className="py-12 text-center">
            <Briefcase
              size={32}
              className="mx-auto mb-2"
              style={{ color: 'var(--hub-text-dim)' }}
            />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No services found
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              Add services to start accepting bookings
            </p>
          </div>
        )}

        {!isLoading && !isError && services && services.length > 0 && (
          <div>
            {services.map((svc: any) => (
              <div
                key={svc.id}
                className="grid grid-cols-5 items-center px-4 py-3 transition-colors"
                style={{ borderBottom: '1px solid var(--hub-border)' }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'transparent';
                }}
              >
                <div>
                  <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {svc.name}
                  </p>
                  {svc.category && (
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {svc.category}
                    </p>
                  )}
                </div>
                <div className="flex items-center gap-1">
                  <Clock size={11} style={{ color: 'var(--hub-text-dim)' }} />
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {svc.duration} min
                  </span>
                </div>
                <div className="flex items-center gap-1">
                  <DollarSign size={11} style={{ color: 'var(--hub-text-dim)' }} />
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {svc.price}
                  </span>
                </div>
                <div className="flex items-center gap-1">
                  <Users size={11} style={{ color: 'var(--hub-text-dim)' }} />
                  <span className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>
                    {Array.isArray(svc.staff) ? svc.staff.join(', ') : svc.staffCount ?? 0}
                  </span>
                </div>
                <div className="flex justify-end">
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                    style={{
                      background:
                        svc.active !== false ? 'rgba(16,185,129,0.1)' : 'rgba(239,68,68,0.1)',
                      color: svc.active !== false ? '#10b981' : '#ef4444',
                    }}
                  >
                    {svc.active !== false ? 'Active' : 'Inactive'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </motion.div>
  );
};

export default Services;
