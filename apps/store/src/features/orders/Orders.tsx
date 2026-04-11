import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { ShoppingCart, Search, Package } from 'lucide-react';
import { useDebounce } from '@voxire/utils';
import { useOrders, OrderFilters } from './useOrders';
import { formatCurrency, formatDateTime } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';

const TEAL = '#14b8a6';

function getStatusColor(status: string): string {
  if (status === 'delivered') return '#10b981';
  if (status === 'shipped') return '#3b82f6';
  if (status === 'confirmed') return TEAL;
  if (status === 'pending') return '#f59e0b';
  if (status === 'cancelled' || status === 'refunded') return '#ef4444';
  return 'var(--hub-text-dim)';
}

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.03 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.18, ease: 'easeOut' } },
};

const statusOptions = ['pending', 'confirmed', 'shipped', 'delivered', 'cancelled', 'refunded'] as const;

const Orders: React.FC = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const [statusFilter, setStatusFilter] = useState<OrderFilters['status']>();

  const { data: orders, isLoading, error } = useOrders({
    search: debouncedSearch || undefined,
    status: statusFilter,
  });

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>Failed to load orders</p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>Please check your connection and try again.</p>
      </div>
    );
  }

  return (
    <motion.div variants={containerVariants} initial="hidden" animate="show" className="space-y-4 max-w-7xl">
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ShoppingCart size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Orders</h2>
          {!isLoading && (
            <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}>
              {(orders ?? []).length}
            </span>
          )}
        </div>
      </motion.div>

      {/* Filters */}
      <motion.div variants={itemVariants} className="flex items-center gap-3">
        <div
          className="flex items-center gap-2 flex-1 max-w-sm px-3 py-2 rounded-lg"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={13} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search orders..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent text-sm outline-none"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <select
          value={statusFilter ?? ''}
          onChange={(e) => setStatusFilter((e.target.value || undefined) as OrderFilters['status'])}
          className="px-3 py-2 rounded-lg text-sm"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
        >
          <option value="">All statuses</option>
          {statusOptions.map((s) => (
            <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
          ))}
        </select>
      </motion.div>

      {/* Order list */}
      {isLoading ? (
        <div className="space-y-1">
          {Array.from({ length: 8 }).map((_, i) => <SkeletonRow key={i} cols={6} />)}
        </div>
      ) : (orders ?? []).length === 0 ? (
        <motion.div variants={itemVariants} className="flex flex-col items-center py-16 gap-3">
          <Package size={32} style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>No orders found</p>
        </motion.div>
      ) : (
        <motion.div variants={containerVariants} className="space-y-1">
          <div className="grid grid-cols-6 gap-2 px-3 py-2 text-[10px] uppercase tracking-wider font-semibold"
            style={{ color: 'var(--hub-text-dim)' }}>
            <span>Order</span>
            <span>Customer</span>
            <span>Items</span>
            <span>Total</span>
            <span>Date</span>
            <span>Status</span>
          </div>
          {(orders ?? []).map((order) => {
            const statusColor = getStatusColor(order.status);
            return (
              <motion.div
                key={order.id}
                variants={itemVariants}
                className="grid grid-cols-6 gap-2 items-center px-3 py-2.5 rounded-lg cursor-pointer"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {order.orderNumber}
                </span>
                <div className="min-w-0">
                  <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                    {order.customerName}
                  </p>
                  <p className="text-[10px] truncate" style={{ color: 'var(--hub-text-dim)' }}>
                    {order.customerEmail}
                  </p>
                </div>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {order.itemCount} items
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(order.total)}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {formatDateTime(order.createdAt)}
                </span>
                <span
                  className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize w-fit"
                  style={{ background: `${statusColor}18`, color: statusColor }}
                >
                  {order.status}
                </span>
              </motion.div>
            );
          })}
        </motion.div>
      )}
    </motion.div>
  );
};

export default Orders;
