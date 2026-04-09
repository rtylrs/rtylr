import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Plus, FileText, Send, CheckCircle } from 'lucide-react';
import { erpApi, PurchaseOrder } from '../../services/erpApi';

const AMBER = '#f59e0b';

// ─── Mock / fallback data ─────────────────────────────────────────────────────

const mockOrders: PurchaseOrder[] = [
  { id: 87,  supplier: 'Fresh Farms Co.',       supplierId: 1, date: '2024-04-01', expectedDate: '2024-04-05', items: 8,  total: '$2,340.00', status: 'Received' },
  { id: 88,  supplier: 'Artisan Bakers Inc.',   supplierId: 3, date: '2024-04-03', expectedDate: '2024-04-06', items: 5,  total: '$890.00',   status: 'Received' },
  { id: 89,  supplier: 'Fresh Farms Co.',       supplierId: 1, date: '2024-04-07', expectedDate: '2024-04-10', items: 12, total: '$1,240.00', status: 'Received' },
  { id: 90,  supplier: 'Global Beverages Ltd.', supplierId: 2, date: '2024-04-08', expectedDate: '2024-04-12', items: 20, total: '$3,560.00', status: 'Sent'     },
  { id: 91,  supplier: 'Artisan Bakers Inc.',   supplierId: 3, date: '2024-04-09', expectedDate: '2024-04-13', items: 4,  total: '$890.00',   status: 'Draft'    },
  { id: 92,  supplier: 'Dairy Direct Ltd.',     supplierId: 4, date: '2024-04-09', expectedDate: '2024-04-14', items: 9,  total: '$680.00',   status: 'Draft'    },
];

// ─── Config ───────────────────────────────────────────────────────────────────

const statusConfig: Record<PurchaseOrder['status'], { color: string; bg: string; icon: React.ComponentType<{ size?: number }> }> = {
  Draft:    { color: 'var(--hub-text-muted)', bg: 'var(--hub-hover)',          icon: FileText    },
  Sent:     { color: AMBER,                   bg: 'rgba(245,158,11,0.1)',       icon: Send        },
  Received: { color: '#10b981',               bg: 'rgba(16,185,129,0.1)',       icon: CheckCircle },
};

// ─── Animation variants ───────────────────────────────────────────────────────

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const columns = ['Order ID', 'Supplier', 'Order Date', 'Expected', 'Items', 'Total', 'Status'];

// ─── Skeleton row ─────────────────────────────────────────────────────────────

const SkeletonTableRow: React.FC = () => (
  <div
    className="grid px-4 py-3 items-center gap-2"
    style={{ gridTemplateColumns: '1.2fr 1.5fr 1fr 1fr 0.6fr 0.9fr 0.9fr' }}
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

// ─── Component ────────────────────────────────────────────────────────────────

const Purchasing: React.FC = () => {
  const [orders, setOrders] = useState<PurchaseOrder[]>(mockOrders);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function fetchOrders() {
      setIsLoading(true);
      try {
        const res = await erpApi.listPurchaseOrders();
        if (!cancelled && res.data?.length) {
          setOrders(res.data);
        }
      } catch {
        // silently fall back to mock data
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    }

    fetchOrders();
    return () => { cancelled = true; };
  }, []);

  const statusCounts = {
    Draft:    orders.filter((o) => o.status === 'Draft').length,
    Sent:     orders.filter((o) => o.status === 'Sent').length,
    Received: orders.filter((o) => o.status === 'Received').length,
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-5 max-w-5xl"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Purchase Orders
        </h2>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-opacity"
          style={{ background: AMBER, color: '#000' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.opacity = '0.9'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.opacity = '1'; }}
        >
          <Plus size={14} />
          New Order
        </button>
      </motion.div>

      {/* Stats row */}
      <motion.div variants={itemVariants} className="grid grid-cols-3 gap-3">
        {(Object.entries(statusCounts) as [PurchaseOrder['status'], number][]).map(([status, count]) => {
          const sc = statusConfig[status];
          const Icon = sc.icon;
          return (
            <div
              key={status}
              className="flex items-center gap-3 px-4 py-3 rounded-xl"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <span
                className="flex items-center justify-center h-8 w-8 rounded-lg shrink-0"
                style={{ background: sc.bg }}
              >
                <Icon size={15} style={{ color: sc.color }} />
              </span>
              <div>
                <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>
                  {isLoading ? '…' : count}
                </p>
                <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {status}
                </p>
              </div>
            </div>
          );
        })}
      </motion.div>

      {/* Table */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        {/* Column headers */}
        <div
          className="grid px-4 py-3"
          style={{
            gridTemplateColumns: '1.2fr 1.5fr 1fr 1fr 0.6fr 0.9fr 0.9fr',
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

        {/* Rows */}
        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading ? (
            Array.from({ length: 6 }).map((_, i) => <SkeletonTableRow key={i} />)
          ) : (
            orders.map((order) => {
              const sc = statusConfig[order.status];
              const orderId = typeof order.id === 'number' ? `PO-${order.id}` : order.id;
              return (
                <div
                  key={order.id}
                  className="grid px-4 py-3 items-center transition-colors cursor-pointer"
                  style={{ gridTemplateColumns: '1.2fr 1.5fr 1fr 1fr 0.6fr 0.9fr 0.9fr' }}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <span className="text-xs font-mono font-medium" style={{ color: 'var(--hub-text)' }}>
                    {orderId}
                  </span>
                  <span className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
                    {order.supplier}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {order.date}
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
                </div>
              );
            })
          )}
        </div>
      </motion.div>
    </motion.div>
  );
};

export default Purchasing;
