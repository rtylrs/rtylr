import React from 'react';
import { motion } from 'motion/react';
import { Truck, Plus, MapPin, CheckCircle, XCircle } from 'lucide-react';
import { useShippingZones, useShippingMethods } from './useShipping';
import { formatCurrency } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';

const TEAL = '#14b8a6';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

function methodTypeLabel(type: string): string {
  if (type === 'flat_rate') return 'Flat rate';
  if (type === 'free') return 'Free shipping';
  if (type === 'weight_based') return 'Weight-based';
  if (type === 'price_based') return 'Price-based';
  return type;
}

const Shipping: React.FC = () => {
  const { data: zones, isLoading: zonesLoading, error: zonesError } = useShippingZones();
  const { data: methods, isLoading: methodsLoading, error: methodsError } = useShippingMethods();

  const isLoading = zonesLoading || methodsLoading;
  const hasError = zonesError || methodsError;

  if (hasError) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>
          Failed to load shipping data
        </p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Please check your connection and try again.
        </p>
      </div>
    );
  }

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-5xl"
    >
      <motion.div variants={itemVariants as any} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Truck size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Shipping
          </h2>
        </div>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: TEAL, color: '#fff' }}
        >
          <Plus size={14} />
          Add Zone
        </button>
      </motion.div>

      {/* Shipping zones */}
      <motion.div variants={itemVariants as any}>
        <h3 className="text-sm font-semibold mb-3" style={{ color: 'var(--hub-text)' }}>
          Shipping Zones
        </h3>
        {isLoading ? (
          <div className="space-y-1">
            {Array.from({ length: 3 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : (zones ?? []).length === 0 ? (
          <div
            className="flex flex-col items-center py-12 gap-3 rounded-xl"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <MapPin size={28} style={{ color: 'var(--hub-text-dim)' }} />
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No shipping zones configured
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {(zones ?? []).map((zone) => (
              <div
                key={zone.id}
                className="rounded-xl p-4"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <MapPin size={13} style={{ color: TEAL }} />
                  <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {zone.name}
                  </span>
                </div>
                <div className="flex flex-wrap gap-1.5 mb-3">
                  {zone.regions.map((r) => (
                    <span
                      key={r}
                      className="text-[10px] px-2 py-0.5 rounded-full"
                      style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                    >
                      {r}
                    </span>
                  ))}
                </div>
                {zone.methods.length > 0 && (
                  <div className="space-y-1">
                    {zone.methods.map((m) => (
                      <div
                        key={m.id}
                        className="flex items-center gap-3 px-3 py-2 rounded-lg"
                        style={{ background: 'var(--hub-hover)' }}
                      >
                        <span
                          className="text-xs font-medium flex-1"
                          style={{ color: 'var(--hub-text)' }}
                        >
                          {m.name}
                        </span>
                        <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                          {methodTypeLabel(m.type)}
                        </span>
                        <span
                          className="text-xs font-medium"
                          style={{ color: 'var(--hub-text-muted)' }}
                        >
                          {formatCurrency(m.cost)}
                        </span>
                        <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                          {m.estimatedDays}
                        </span>
                        {m.enabled ? (
                          <CheckCircle size={12} style={{ color: '#10b981' }} />
                        ) : (
                          <XCircle size={12} style={{ color: '#ef4444' }} />
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </motion.div>

      {/* Standalone methods */}
      {(methods ?? []).length > 0 && (
        <motion.div variants={itemVariants as any}>
          <h3 className="text-sm font-semibold mb-3" style={{ color: 'var(--hub-text)' }}>
            Shipping Methods
          </h3>
          <div className="space-y-1">
            <div
              className="grid grid-cols-5 gap-2 px-3 py-2 text-[10px] uppercase tracking-wider font-semibold"
              style={{ color: 'var(--hub-text-dim)' }}
            >
              <span>Method</span>
              <span>Type</span>
              <span>Cost</span>
              <span>Estimated</span>
              <span>Status</span>
            </div>
            {(methods ?? []).map((m) => (
              <div
                key={m.id}
                className="grid grid-cols-5 gap-2 items-center px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {m.name}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {methodTypeLabel(m.type)}
                </span>
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(m.cost)}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {m.estimatedDays}
                </span>
                <span
                  className="text-[10px] font-medium px-2 py-0.5 rounded-full w-fit"
                  style={{
                    background: m.enabled ? 'rgba(16,185,129,0.1)' : 'rgba(239,68,68,0.1)',
                    color: m.enabled ? '#10b981' : '#ef4444',
                  }}
                >
                  {m.enabled ? 'Enabled' : 'Disabled'}
                </span>
              </div>
            ))}
          </div>
        </motion.div>
      )}
    </motion.div>
  );
};

export default Shipping;
