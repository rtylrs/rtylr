import React from 'react';
import { motion } from 'framer-motion';
import {
  Package,
  AlertTriangle,
  ShoppingCart,
  Users,
  TrendingDown,
  Clock,
} from 'lucide-react';
import { LowStockItem, PurchaseOrder } from '../../services/erpApi';
import { useLowStock } from '../../hooks/useInventory';
import { usePurchaseOrders } from '../../hooks/usePurchaseOrders';
import { useProducts } from '../../hooks/useProducts';
import { useSuppliers } from '../../hooks/useSuppliers';

const AMBER = '#f59e0b';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

// ─── Mock / fallback data ─────────────────────────────────────────────────────

const mockLowStockItems: LowStockItem[] = [
  { name: 'Arabica Coffee Beans', sku: 'BEV-001', stock: 3, minStock: 20, unit: 'kg' },
  { name: 'Whole Milk (2L)', sku: 'BEV-012', stock: 8, minStock: 50, unit: 'units' },
  { name: 'Sourdough Bread', sku: 'FOOD-034', stock: 2, minStock: 15, unit: 'loaves' },
];

const mockRecentOrders: PurchaseOrder[] = [
  { id: 89, supplier: 'Fresh Farms Co.', supplierId: 1, date: '2024-04-07', expectedDate: '2024-04-10', items: 12, total: '$1,240.00', status: 'Received' },
  { id: 90, supplier: 'Global Beverages Ltd.', supplierId: 2, date: '2024-04-08', expectedDate: '2024-04-12', items: 20, total: '$3,560.00', status: 'Sent' },
  { id: 91, supplier: 'Artisan Bakers Inc.', supplierId: 3, date: '2024-04-09', expectedDate: '2024-04-13', items: 4, total: '$890.00', status: 'Draft' },
];

const mockStats = {
  totalProducts: 248,
  lowStockCount: 14,
  pendingOrders: 7,
  totalSuppliers: 32,
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getStatusColor(status: PurchaseOrder['status']): string {
  if (status === 'Received') return '#10b981';
  if (status === 'Sent') return AMBER;
  return 'var(--hub-text-dim)';
}

// ─── Animation variants ───────────────────────────────────────────────────────

const containerVariants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: 0.07,
    },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

// ─── Skeleton ─────────────────────────────────────────────────────────────────

const SkeletonRow: React.FC<{ cols: number }> = ({ cols }) => (
  <div
    className="grid px-3 py-2.5 rounded-lg gap-2"
    style={{ gridTemplateColumns: `repeat(${cols}, 1fr)`, background: 'var(--hub-hover)' }}
  >
    {Array.from({ length: cols }).map((_, i) => (
      <div
        key={i}
        className="h-3 rounded animate-pulse"
        style={{ background: 'var(--hub-border)', opacity: 0.6 }}
      />
    ))}
  </div>
);

// ─── Component ────────────────────────────────────────────────────────────────

const Dashboard: React.FC = () => {
  const { data: lowStockData, isLoading: lowStockLoading } = useLowStock();
  const { data: ordersData, isLoading: ordersLoading } = usePurchaseOrders();
  const { data: productsData, isLoading: productsLoading } = useProducts();
  const { data: suppliersData, isLoading: suppliersLoading } = useSuppliers();

  const isLoading = lowStockLoading || ordersLoading || productsLoading || suppliersLoading;

  const lowStockItems: LowStockItem[] = lowStockData && lowStockData.length ? lowStockData : mockLowStockItems;
  const recentOrders: PurchaseOrder[] = ordersData && ordersData.length ? ordersData.slice(0, 3) : mockRecentOrders;

  const statsData = {
    totalProducts: productsData && productsData.length ? productsData.length : mockStats.totalProducts,
    lowStockCount:
      lowStockData && lowStockData.length
        ? lowStockData.length
        : productsData && productsData.length
        ? productsData.filter((p: any) => p.status === 'Low Stock').length
        : mockStats.lowStockCount,
    pendingOrders:
      ordersData && ordersData.length
        ? ordersData.filter((o) => o.status === 'Sent' || o.status === 'Draft').length
        : mockStats.pendingOrders,
    totalSuppliers: suppliersData && suppliersData.length ? suppliersData.length : mockStats.totalSuppliers,
  };

  const stats = [
    {
      label: 'Total Products',
      value: isLoading ? '…' : String(statsData.totalProducts),
      icon: Package,
      change: '+12 this month',
      changePositive: true,
    },
    {
      label: 'Low Stock Items',
      value: isLoading ? '…' : String(statsData.lowStockCount),
      icon: AlertTriangle,
      change: '3 critical',
      changePositive: false,
    },
    {
      label: 'Pending Orders',
      value: isLoading ? '…' : String(statsData.pendingOrders),
      icon: ShoppingCart,
      change: '2 awaiting approval',
      changePositive: true,
    },
    {
      label: 'Total Suppliers',
      value: isLoading ? '…' : String(statsData.totalSuppliers),
      icon: Users,
      change: '+2 this month',
      changePositive: true,
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-6xl"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Inventory Overview
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {today}
          </p>
        </div>
      </motion.div>

      {/* Stat cards */}
      <motion.div variants={containerVariants} className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <motion.div
              key={stat.label}
              variants={itemVariants}
              className="rounded-xl p-4"
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
              }}
            >
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                  {stat.label}
                </span>
                <span
                  className="flex items-center justify-center h-7 w-7 rounded-lg"
                  style={{ background: 'rgba(245,158,11,0.1)' }}
                >
                  <Icon size={14} style={{ color: AMBER }} />
                </span>
              </div>
              <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                {stat.value}
              </p>
              <p
                className="text-xs mt-1"
                style={{ color: stat.changePositive ? '#10b981' : '#f87171' }}
              >
                {stat.change}
              </p>
            </motion.div>
          );
        })}
      </motion.div>

      {/* Two column grid */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Low Stock Alert */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
          }}
        >
          <div className="flex items-center gap-2 mb-4">
            <TrendingDown size={15} style={{ color: '#f87171' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Low Stock Alert
            </h3>
            <span
              className="ml-auto text-xs px-2 py-0.5 rounded-full font-medium"
              style={{ background: 'rgba(248,113,113,0.1)', color: '#f87171' }}
            >
              {isLoading ? '…' : `${lowStockItems.length} items`}
            </span>
          </div>

          {/* Column headers */}
          <div
            className="grid grid-cols-3 px-3 py-1.5 rounded-lg mb-1 text-[10px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--hub-text-dim)' }}
          >
            <span>Product</span>
            <span className="text-center">Stock</span>
            <span className="text-right">Min. Required</span>
          </div>

          <div className="space-y-1">
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={3} />)
              : lowStockItems.map((item) => (
                <div
                  key={item.sku}
                  className="grid grid-cols-3 items-center px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <div>
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {item.name}
                    </p>
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {item.sku}
                    </p>
                  </div>
                  <p
                    className="text-sm font-bold text-center"
                    style={{ color: item.stock <= 5 ? '#f87171' : AMBER }}
                  >
                    {item.stock} {item.unit}
                  </p>
                  <p className="text-xs text-right" style={{ color: 'var(--hub-text-muted)' }}>
                    {item.minStock} {item.unit}
                  </p>
                </div>
              ))}
          </div>
        </motion.div>

        {/* Recent Purchase Orders */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
          }}
        >
          <div className="flex items-center gap-2 mb-4">
            <Clock size={15} style={{ color: AMBER }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Recent Purchase Orders
            </h3>
          </div>

          {/* Column headers */}
          <div
            className="grid grid-cols-4 px-3 py-1.5 rounded-lg mb-1 text-[10px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--hub-text-dim)' }}
          >
            <span>Order</span>
            <span>Supplier</span>
            <span className="text-right">Total</span>
            <span className="text-right">Status</span>
          </div>

          <div className="space-y-1">
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={4} />)
              : recentOrders.map((order) => {
                const statusColor = getStatusColor(order.status);
                return (
                  <div
                    key={order.id}
                    className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg"
                    style={{ background: 'var(--hub-hover)' }}
                  >
                    <div>
                      <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                        {typeof order.id === 'number' ? `PO-${order.id}` : order.id}
                      </p>
                      <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                        {order.date}
                      </p>
                    </div>
                    <p className="text-xs truncate pr-2" style={{ color: 'var(--hub-text-muted)' }}>
                      {order.supplier}
                    </p>
                    <p className="text-xs font-semibold text-right" style={{ color: 'var(--hub-text)' }}>
                      {order.total}
                    </p>
                    <div className="flex justify-end">
                      <span
                        className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                        style={{
                          background: `${statusColor}18`,
                          color: statusColor,
                        }}
                      >
                        {order.status}
                      </span>
                    </div>
                  </div>
                );
              })}
          </div>
        </motion.div>
      </div>
    </motion.div>
  );
};

export default Dashboard;
