import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, FileText, Send, CheckCircle, Clock as ClockIcon, ThumbsUp } from 'lucide-react';
import { PurchaseOrder } from '../../shared/services/erpApi';
import { usePurchaseOrders } from './usePurchaseOrders';
import PrintPOButton from '../../shared/components/PrintPOButton';
import ExportButton from '../../shared/components/ExportButton';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const statusConfig: Record<
  PurchaseOrder['status'],
  { color: string; bg: string; icon: React.ComponentType<{ size?: number }> }
> = {
  Draft: { color: 'var(--hub-text-muted)', bg: 'var(--hub-hover)', icon: FileText },
  'Pending Approval': { color: '#a78bfa', bg: 'rgba(167,139,250,0.12)', icon: ClockIcon },
  Approved: { color: '#60a5fa', bg: 'rgba(96,165,250,0.12)', icon: ThumbsUp },
  Sent: { color: AMBER, bg: 'rgba(245,158,11,0.1)', icon: Send },
  Received: { color: '#10b981', bg: 'rgba(16,185,129,0.1)', icon: CheckCircle },
};

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const columns = ['Order', 'Supplier', 'Expected', 'Items', 'Total', 'Status', 'Actions'];

const SkeletonTableRow: React.FC = () => (
  <div
    className="grid px-4 py-3 items-center gap-2"
    style={{ gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr' }}
  >
    {Array.from({ length: 7 }).map((_, i) => (
      <div
        key={i}
        className="h-3 rounded animate-pulse"
        style={{ background: 'var(--hub-border)', opacity: 0.6 }}
      />
    ))}
  </div>
);

const Purchasing: React.FC = () => {
  const { data: orders = [], isLoading } = usePurchaseOrders();
  const [filter, setFilter] = useState<PurchaseOrder['status'] | 'All'>('All');

  const filteredOrders = filter === 'All' ? orders : orders.filter((o) => o.status === filter);

  const statusCounts: Record<string, number> = {
    Draft: orders.filter((o) => o.status === 'Draft').length,
    'Pending Approval': orders.filter((o) => o.status === 'Pending Approval').length,
    Sent: orders.filter((o) => o.status === 'Sent').length,
    Received: orders.filter((o) => o.status === 'Received').length,
  };

  const handleApprove = (id: number) => {
    toast.success(`PO-${id} approved`);
  };

  const handleSend = (id: number) => {
    toast.success(`PO-${id} sent to supplier`);
  };

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-5 max-w-6xl"
    >
      <motion.div variants={itemVariants as any} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Purchase Orders
        </h2>
        <div className="flex items-center gap-2">
          <ExportButton
            data={orders.map((o) => ({
              id: o.id,
              supplier: o.supplier,
              date: o.date,
              expectedDate: o.expectedDate,
              items: o.items,
              total: o.total,
              status: o.status,
            }))}
            filename="purchase-orders.csv"
          />
          <button
            onClick={() => toast.success('New purchase order draft created')}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} />
            New Order
          </button>
        </div>
      </motion.div>

      <motion.div variants={itemVariants as any} className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        {(Object.entries(statusCounts) as [PurchaseOrder['status'], number][]).map(
          ([status, count]) => {
            const sc = statusConfig[status];
            if (!sc) return null;
            const Icon = sc.icon;
            const active = filter === status;
            return (
              <button
                key={status}
                onClick={() => setFilter(active ? 'All' : status)}
                className="flex items-center gap-3 px-4 py-3 rounded-xl transition-all text-left"
                style={{
                  background: 'var(--hub-card)',
                  border: active ? '1px solid rgba(245,158,11,0.5)' : '1px solid var(--hub-border)',
                }}
              >
                <span
                  className="flex items-center justify-center h-8 w-8 rounded-lg shrink-0"
                  style={{ background: sc.bg }}
                >
                  <Icon size={15} {...({ style: { color: sc.color } } as any)} />
                </span>
                <div>
                  <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>
                    {isLoading ? '…' : count}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {status}
                  </p>
                </div>
              </button>
            );
          }
        )}
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3"
          style={{
            gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          {columns.map((col) => (
            <span
              key={col}
              className="text-[10px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--hub-text-dim)' }}
            >
              {col}
            </span>
          ))}
        </div>

        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading
            ? Array.from({ length: 6 }).map((_, i) => <SkeletonTableRow key={i} />)
            : filteredOrders.map((order) => {
                const sc = statusConfig[order.status] || statusConfig['Draft'];
                return (
                  <div
                    key={order.id}
                    className="grid px-4 py-3 items-center transition-colors"
                    style={{ gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <span
                      className="text-xs font-mono font-medium"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      PO-{order.id}
                    </span>
                    <span
                      className="text-sm truncate pr-2"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      {order.supplier}
                    </span>
                    <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                      {order.expectedDate}
                    </span>
                    <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {order.items}
                    </span>
                    <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                      {order.total}
                    </span>
                    <span>
                      <span
                        className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                        style={{ background: sc.bg, color: sc.color }}
                      >
                        {order.status}
                      </span>
                    </span>
                    <div className="flex items-center gap-1">
                      {order.status === 'Pending Approval' && (
                        <button
                          onClick={() => handleApprove(order.id)}
                          className="text-[10px] font-semibold px-2 py-1 rounded-md"
                          style={{ background: 'rgba(96,165,250,0.15)', color: '#60a5fa' }}
                        >
                          Approve
                        </button>
                      )}
                      {order.status === 'Draft' && (
                        <button
                          onClick={() => handleSend(order.id)}
                          className="text-[10px] font-semibold px-2 py-1 rounded-md"
                          style={{ background: 'rgba(245,158,11,0.15)', color: AMBER }}
                        >
                          Send
                        </button>
                      )}
                      <PrintPOButton order={order} />
                    </div>
                  </div>
                );
              })}
        </div>
      </motion.div>
    </motion.div>
  );
};

export default Purchasing;
