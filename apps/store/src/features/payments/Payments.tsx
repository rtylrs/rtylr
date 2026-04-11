import React from 'react';
import { motion } from 'motion/react';
import { CreditCard, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import { usePaymentMethods } from './usePayments';
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

function typeLabel(type: string): string {
  if (type === 'credit_card') return 'Credit Card';
  if (type === 'apple_pay') return 'Apple Pay';
  if (type === 'google_pay') return 'Google Pay';
  if (type === 'cod') return 'Cash on Delivery';
  if (type === 'bank_transfer') return 'Bank Transfer';
  return type.charAt(0).toUpperCase() + type.slice(1);
}

const Payments: React.FC = () => {
  const { data: methods, isLoading, error } = usePaymentMethods();

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>
          Failed to load payment methods
        </p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Please check your connection and try again.
        </p>
      </div>
    );
  }

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-4 max-w-5xl"
    >
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CreditCard size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Payment Methods
          </h2>
        </div>
      </motion.div>

      {isLoading ? (
        <div className="space-y-1">
          {Array.from({ length: 4 }).map((_, i) => (
            <SkeletonRow key={i} cols={5} />
          ))}
        </div>
      ) : (methods ?? []).length === 0 ? (
        <motion.div variants={itemVariants} className="flex flex-col items-center py-16 gap-3">
          <CreditCard size={32} style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
            No payment methods configured
          </p>
          <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
            Configure payment providers to start accepting orders.
          </p>
        </motion.div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {(methods ?? []).map((method) => (
            <motion.div
              key={method.id}
              variants={itemVariants}
              className="rounded-xl p-4 transition-all"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div
                    className="h-9 w-9 rounded-lg flex items-center justify-center"
                    style={{
                      background: method.enabled ? 'rgba(20,184,166,0.08)' : 'var(--hub-hover)',
                    }}
                  >
                    <CreditCard
                      size={16}
                      style={{ color: method.enabled ? TEAL : 'var(--hub-text-dim)' }}
                    />
                  </div>
                  <div>
                    <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                      {method.name}
                    </p>
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-muted)' }}>
                      {method.provider}
                    </p>
                  </div>
                </div>
                {method.enabled ? (
                  <CheckCircle size={16} style={{ color: '#10b981' }} />
                ) : (
                  <XCircle size={16} style={{ color: '#ef4444' }} />
                )}
              </div>
              <div className="flex items-center gap-3">
                <span
                  className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                  style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                >
                  {typeLabel(method.type)}
                </span>
                {method.testMode && (
                  <span
                    className="flex items-center gap-1 text-[10px] font-medium px-2 py-0.5 rounded-full"
                    style={{ background: 'rgba(245,158,11,0.1)', color: '#f59e0b' }}
                  >
                    <AlertTriangle size={9} />
                    Test Mode
                  </span>
                )}
                <span
                  className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                  style={{
                    background: method.enabled ? 'rgba(16,185,129,0.1)' : 'rgba(239,68,68,0.1)',
                    color: method.enabled ? '#10b981' : '#ef4444',
                  }}
                >
                  {method.enabled ? 'Active' : 'Inactive'}
                </span>
              </div>
            </motion.div>
          ))}
        </motion.div>
      )}
    </motion.div>
  );
};

export default Payments;
