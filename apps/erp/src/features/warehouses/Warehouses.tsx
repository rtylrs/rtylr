import React from 'react';
import { motion } from 'framer-motion';
import { Plus, Warehouse as WarehouseIcon, MapPin, User } from 'lucide-react';
import { useWarehouses } from './useWarehouses';
import PageHeader from '../../shared/components/PageHeader';
import EmptyState from '../../shared/components/EmptyState';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatNumber } from '../../shared/lib/format';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const Warehouses: React.FC = () => {
  const { data: warehouses = [], isLoading } = useWarehouses();

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
      <PageHeader
        title="Warehouses"
        description={`${warehouses.length} active storage locations`}
        actions={
          <button
            onClick={() => toast.success('Warehouse created')}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Warehouse
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-40 w-full rounded-xl" />
          ))}
        </div>
      ) : warehouses.length === 0 ? (
        <EmptyState title="No warehouses" description="Create your first warehouse to track stock locations" />
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {warehouses.map((w) => {
            const pct = Math.round((w.used / w.capacity) * 100);
            return (
              <div
                key={w.id}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div
                      className="h-10 w-10 rounded-lg flex items-center justify-center"
                      style={{ background: 'rgba(245,158,11,0.1)' }}
                    >
                      <WarehouseIcon size={16} style={{ color: AMBER }} />
                    </div>
                    <div>
                      <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                        {w.name}
                      </p>
                      {w.isDefault && (
                        <span
                          className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                          style={{ background: 'rgba(16,185,129,0.1)', color: '#10b981' }}
                        >
                          Default
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <div className="space-y-1.5 mb-4">
                  <div className="flex items-center gap-2 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    <MapPin size={11} /> {w.address}
                  </div>
                  <div className="flex items-center gap-2 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    <User size={11} /> Managed by {w.manager}
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-xs mb-1">
                    <span style={{ color: 'var(--hub-text-muted)' }}>Capacity</span>
                    <span style={{ color: 'var(--hub-text)' }}>
                      {formatNumber(w.used)} / {formatNumber(w.capacity)}
                    </span>
                  </div>
                  <div
                    className="h-2 rounded-full overflow-hidden"
                    style={{ background: 'var(--hub-hover)' }}
                  >
                    <div
                      className="h-full rounded-full"
                      style={{
                        width: `${pct}%`,
                        background: pct > 85 ? '#f87171' : pct > 65 ? AMBER : '#10b981',
                      }}
                    />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default Warehouses;
