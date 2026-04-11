import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Monitor, Search, Filter } from 'lucide-react';
import { useResources } from './useResources';

const INDIGO = '#6366f1';

const TYPE_OPTIONS = ['all', 'room', 'station', 'equipment'];

const SkeletonRow: React.FC = () => (
  <div className="h-12 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

function getTypeColor(type: string): string {
  if (type === 'room') return '#10b981';
  if (type === 'station') return '#f59e0b';
  if (type === 'equipment') return '#60a5fa';
  return 'var(--hub-text-dim)';
}

const Resources: React.FC = () => {
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');

  const { data: resources, isLoading, isError } = useResources({
    search: search || undefined,
    type: typeFilter !== 'all' ? typeFilter : undefined,
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
            Resources
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Manage rooms, stations, and equipment
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
            placeholder="Search resources..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-xs"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <div className="flex items-center gap-1">
          <Filter size={12} style={{ color: 'var(--hub-text-dim)' }} />
          {TYPE_OPTIONS.map((t) => (
            <button
              key={t}
              onClick={() => setTypeFilter(t)}
              className="px-2 py-1 rounded-md text-[11px] font-medium capitalize"
              style={{
                background: typeFilter === t ? 'rgba(99,102,241,0.12)' : 'transparent',
                color: typeFilter === t ? INDIGO : 'var(--hub-text-muted)',
              }}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      {isError && (
        <div
          className="rounded-xl p-4 text-sm"
          style={{ background: 'rgba(239,68,68,0.08)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.2)' }}
        >
          Failed to load resources. Please try again later.
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
          <span>Name</span>
          <span>Type</span>
          <span>Capacity</span>
          <span>Location</span>
          <span className="text-right">Status</span>
        </div>

        {isLoading && (
          <div className="p-3 space-y-2">
            {Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        )}

        {!isLoading && !isError && (!resources || resources.length === 0) && (
          <div className="py-12 text-center">
            <Monitor size={32} className="mx-auto mb-2" style={{ color: 'var(--hub-text-dim)' }} />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No resources found
            </p>
            <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              Add rooms, stations, or equipment to manage availability
            </p>
          </div>
        )}

        {!isLoading && !isError && resources && resources.length > 0 && (
          <div>
            {resources.map((res: any) => {
              const typeColor = getTypeColor(res.type);
              return (
                <div
                  key={res.id}
                  className="grid grid-cols-5 items-center px-4 py-3 transition-colors"
                  style={{ borderBottom: '1px solid var(--hub-border)' }}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{res.name}</p>
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize w-fit"
                    style={{ background: `${typeColor}18`, color: typeColor }}
                  >
                    {res.type}
                  </span>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{res.capacity ?? '-'}</p>
                  <p className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>{res.location ?? '-'}</p>
                  <div className="flex justify-end">
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{
                        background: res.available !== false ? 'rgba(16,185,129,0.1)' : 'rgba(239,68,68,0.1)',
                        color: res.available !== false ? '#10b981' : '#ef4444',
                      }}
                    >
                      {res.available !== false ? 'Available' : 'Unavailable'}
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

export default Resources;
