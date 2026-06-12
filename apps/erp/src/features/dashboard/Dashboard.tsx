import React from 'react';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import {
  Package,
  AlertTriangle,
  ShoppingCart,
  Users,
  TrendingDown,
  Clock,
  TrendingUp,
  Activity,
  PackagePlus,
} from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { LowStockItem, PurchaseOrder } from '../../shared/services/erpApi';
import { useLowStock } from '../inventory/useInventory';
import { usePurchaseOrders } from '../purchasing/usePurchaseOrders';
import { useProducts } from '../products/useProducts';
import { useSuppliers } from '../suppliers/useSuppliers';
import { useActivity } from '../../shared/hooks/useActivity';
import { timeAgo, formatCurrency } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';

const AMBER = 'var(--erp-accent, #f59e0b)';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

function getStatusColor(status: PurchaseOrder['status']): string {
  if (status === 'Received') return 'var(--erp-success, #10b981)';
  if (status === 'Sent' || status === 'Approved') return AMBER;
  if (status === 'Pending Approval') return 'var(--erp-purple, #a78bfa)';
  return 'var(--hub-text-dim)';
}

function getStatusBg(status: PurchaseOrder['status']): string {
  if (status === 'Received') return 'rgba(16,185,129,0.09)';
  if (status === 'Sent' || status === 'Approved') return 'rgba(245,158,11,0.09)';
  if (status === 'Pending Approval') return 'rgba(167,139,250,0.09)';
  return 'rgba(0,0,0,0.06)';
}

const activityIconFor = (type: string) => {
  if (type === 'order') return ShoppingCart;
  if (type === 'product') return Package;
  if (type === 'transfer') return TrendingUp;
  if (type === 'supplier') return Users;
  return Activity;
};

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Dashboard: React.FC = () => {
  const { data: lowStockData, isLoading: lowStockLoading } = useLowStock();
  const { data: ordersResult, isLoading: ordersLoading } = usePurchaseOrders();
  const { data: productsResult, isLoading: productsLoading } = useProducts();
  const { data: suppliersResult, isLoading: suppliersLoading } = useSuppliers();
  const productsData = productsResult?.products ?? [];
  const suppliersData = suppliersResult?.suppliers ?? [];
  const { data: activityData = [] } = useActivity(8);

  const isLoading = lowStockLoading || ordersLoading || productsLoading || suppliersLoading;

  // useLowStock returns the backend inventory-report shape; map it to the
  // LowStockItem display shape this widget renders.
  const lowStockItems: LowStockItem[] = (lowStockData ?? []).map((i) => ({
    name: i.productName,
    sku: i.sku,
    stock: i.stockQuantity,
    minStock: i.lowStockThreshold,
    unit: i.unit,
  }));
  const ordersForList: PurchaseOrder[] = ordersResult?.orders ?? [];
  const recentOrders = ordersForList.slice(0, 3);

  const stats = [
    {
      label: 'Total Products',
      value: productsData?.length ?? 0,
      icon: Package,
      change: '',
      changePositive: true,
      to: '/products',
    },
    {
      label: 'Low Stock Items',
      value: lowStockItems.length,
      icon: AlertTriangle,
      change: `${lowStockItems.filter((i) => i.stock <= 5).length} critical`,
      changePositive: false,
      to: '/alerts',
    },
    {
      label: 'Pending Orders',
      value: ordersForList.filter((o) => o.status === 'Sent' || o.status === 'Draft').length,
      icon: ShoppingCart,
      change: `${ordersForList.filter((o) => o.status === 'Pending Approval').length} awaiting approval`,
      changePositive: true,
      to: '/purchasing',
    },
    {
      label: 'Total Suppliers',
      value: suppliersData?.length ?? 0,
      icon: Users,
      change: '',
      changePositive: true,
      to: '/suppliers',
    },
  ];

  const weeklyValue = (productsData || [])
    .slice(0, 5)
    .reduce((s, p) => s + p.salePrice * (p.stock || 0), 0);

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="w-full min-w-0 space-y-6"
    >
      <HelmetContainer title="Dashboard · rtylr" />
      <motion.div
        variants={itemVariants as any}
        className="flex flex-wrap items-start justify-between gap-4"
      >
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Inventory Overview
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {today}
          </p>
        </div>
        <Link
          to="/products"
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: AMBER, color: '#000', textDecoration: 'none' }}
        >
          <PackagePlus size={14} />
          New Product
        </Link>
      </motion.div>

      {/* Stat cards */}
      <motion.div
        variants={containerVariants as any}
        className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
      >
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <motion.div key={stat.label} variants={itemVariants as any}>
              <Link
                to={stat.to}
                className="block rounded-xl p-4 transition-all"
                style={{
                  background: 'var(--hub-card)',
                  border: '1px solid var(--hub-border)',
                  textDecoration: 'none',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.borderColor = 'rgba(245,158,11,0.3)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
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
                  {isLoading ? '…' : stat.value}
                </p>
                {stat.change && (
                  <p
                    className="text-xs mt-1"
                    style={{
                      color: stat.changePositive
                        ? 'var(--erp-success, #10b981)'
                        : 'var(--erp-danger, #f87171)',
                    }}
                  >
                    {stat.change}
                  </p>
                )}
              </Link>
            </motion.div>
          );
        })}
      </motion.div>

      {/* Extra widgets row */}
      <motion.div
        variants={containerVariants as any}
        className="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,1.1fr)_minmax(320px,0.9fr)]"
      >
        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <TrendingUp size={14} style={{ color: 'var(--erp-success, #10b981)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Catalog Value (Top 5)
            </h3>
          </div>
          <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
            {formatCurrency(weeklyValue)}
          </p>
          <p className="text-[11px] mt-1" style={{ color: 'var(--hub-text-dim)' }}>
            top 5 products by price
          </p>
        </motion.div>

        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <Package size={14} style={{ color: AMBER }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Top 5 Items
            </h3>
          </div>
          <div className="space-y-1.5">
            {(productsData || []).slice(0, 5).map((p, i) => (
              <div key={p.id} className="flex items-center gap-2">
                <span className="text-[10px] w-4" style={{ color: 'var(--hub-text-dim)' }}>
                  {i + 1}
                </span>
                <span
                  className="text-xs flex-1 truncate"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  {p.name}
                </span>
              </div>
            ))}
          </div>
        </motion.div>

        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-3">
            <Activity size={14} style={{ color: 'var(--erp-info, #60a5fa)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Recent Activity
            </h3>
          </div>
          <div className="space-y-2">
            {activityData.slice(0, 5).map((a) => {
              const Icon = activityIconFor(a.type);
              return (
                <div key={a.id} className="flex items-start gap-2">
                  <Icon
                    size={10}
                    className="mt-1 shrink-0"
                    style={{ color: 'var(--hub-text-dim)' }}
                  />
                  <div className="flex-1 min-w-0">
                    <p
                      className="text-[11px] font-medium truncate"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      {a.title}
                    </p>
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {timeAgo(a.timestamp)}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        </motion.div>
      </motion.div>

      {/* Two column grid */}
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-4">
            <TrendingDown size={15} style={{ color: 'var(--erp-danger, #f87171)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Low Stock Alert
            </h3>
            <span
              className="ml-auto text-xs px-2 py-0.5 rounded-full font-medium"
              style={{ background: 'rgba(248,113,113,0.1)', color: 'var(--erp-danger, #f87171)' }}
            >
              {isLoading ? '…' : `${lowStockItems.length} items`}
            </span>
          </div>
          <div className="space-y-1">
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={3} />)
              : lowStockItems.slice(0, 5).map((item) => (
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
                      style={{ color: item.stock <= 5 ? 'var(--erp-danger, #f87171)' : AMBER }}
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

        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-4">
            <Clock size={15} style={{ color: AMBER }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Recent Purchase Orders
            </h3>
          </div>
          <div className="space-y-1">
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={4} />)
              : recentOrders.map((order) => {
                  const statusColor = getStatusColor(order.status);
                  const statusBg = getStatusBg(order.status);
                  return (
                    <div
                      key={order.id}
                      className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg"
                      style={{ background: 'var(--hub-hover)' }}
                    >
                      <div>
                        <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                          PO-{order.id}
                        </p>
                        <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                          {order.date}
                        </p>
                      </div>
                      <p
                        className="text-xs truncate pr-2"
                        style={{ color: 'var(--hub-text-muted)' }}
                      >
                        {order.supplier}
                      </p>
                      <p
                        className="text-xs font-semibold text-right"
                        style={{ color: 'var(--hub-text)' }}
                      >
                        {order.total}
                      </p>
                      <div className="flex justify-end">
                        <span
                          className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                          style={{ background: statusBg, color: statusColor }}
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
