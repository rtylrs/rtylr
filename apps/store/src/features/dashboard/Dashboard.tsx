import React from 'react';
import { motion } from 'motion/react';
import {
  Package,
  ShoppingCart,
  DollarSign,
  Eye,
  TrendingUp,
  ArrowRight,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { useDashboardStats } from './useDashboard';
import { useOrders } from '../orders/useOrders';
import { useProducts } from '../products/useProducts';
import { formatCurrency, formatNumber } from '../../shared/lib/format';
import { SkeletonCard, SkeletonRow } from '../../shared/components/Skeleton';

const TEAL = '#14b8a6';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

function getOrderStatusColor(status: string): string {
  if (status === 'delivered') return '#10b981';
  if (status === 'shipped') return '#3b82f6';
  if (status === 'confirmed') return TEAL;
  if (status === 'pending') return '#f59e0b';
  if (status === 'cancelled' || status === 'refunded') return '#ef4444';
  return 'var(--hub-text-dim)';
}

const Dashboard: React.FC = () => {
  const { data: stats, isLoading: statsLoading, error: statsError } = useDashboardStats();
  const { data: recentOrders, isLoading: ordersLoading, error: ordersError } = useOrders({ limit: 5 });
  const { data: products, isLoading: productsLoading } = useProducts({ limit: 5 });

  const isLoading = statsLoading || ordersLoading || productsLoading;
  const hasError = statsError || ordersError;

  const statCards = [
    {
      label: 'Total Products',
      value: stats?.totalProducts ?? 0,
      icon: Package,
      to: '/products',
    },
    {
      label: 'Orders Today',
      value: stats?.ordersToday ?? 0,
      icon: ShoppingCart,
      to: '/orders',
    },
    {
      label: 'Revenue Today',
      value: formatCurrency(stats?.revenueToday ?? 0),
      icon: DollarSign,
      to: '/orders',
      isFormatted: true,
    },
    {
      label: 'Active Visitors',
      value: stats?.activeVisitors ?? 0,
      icon: Eye,
      to: '/storefront',
    },
  ];

  if (hasError) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>
          Failed to load dashboard data
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
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants} className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Store Overview
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {today}
          </p>
        </div>
      </motion.div>

      {/* Stat cards */}
      <motion.div variants={containerVariants} className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {isLoading
          ? Array.from({ length: 4 }).map((_, i) => (
              <motion.div key={i} variants={itemVariants}>
                <SkeletonCard />
              </motion.div>
            ))
          : statCards.map((stat) => {
              const Icon = stat.icon;
              return (
                <motion.div key={stat.label} variants={itemVariants}>
                  <Link
                    to={stat.to}
                    className="block rounded-xl p-4 transition-all"
                    style={{
                      background: 'var(--hub-card)',
                      border: '1px solid var(--hub-border)',
                      textDecoration: 'none',
                    }}
                    onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'rgba(20,184,166,0.3)'; }}
                    onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)'; }}
                  >
                    <div className="flex items-center justify-between mb-3">
                      <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                        {stat.label}
                      </span>
                      <span
                        className="flex items-center justify-center h-7 w-7 rounded-lg"
                        style={{ background: 'rgba(20,184,166,0.1)' }}
                      >
                        <Icon size={14} style={{ color: TEAL }} />
                      </span>
                    </div>
                    <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                      {stat.isFormatted ? stat.value : formatNumber(stat.value as number)}
                    </p>
                  </Link>
                </motion.div>
              );
            })}
      </motion.div>

      {/* Recent orders + top products */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <ShoppingCart size={15} style={{ color: TEAL }} />
              <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                Recent Orders
              </h3>
            </div>
            <Link
              to="/orders"
              className="flex items-center gap-1 text-xs"
              style={{ color: TEAL, textDecoration: 'none' }}
            >
              View all <ArrowRight size={11} />
            </Link>
          </div>
          <div className="space-y-1">
            {ordersLoading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={4} />)
              : (recentOrders ?? []).length === 0
                ? (
                    <p className="text-xs py-6 text-center" style={{ color: 'var(--hub-text-dim)' }}>
                      No orders yet
                    </p>
                  )
                : (recentOrders ?? []).slice(0, 5).map((order) => {
                    const statusColor = getOrderStatusColor(order.status);
                    return (
                      <div
                        key={order.id}
                        className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg"
                        style={{ background: 'var(--hub-hover)' }}
                      >
                        <div>
                          <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                            {order.orderNumber}
                          </p>
                        </div>
                        <p className="text-xs truncate pr-2" style={{ color: 'var(--hub-text-muted)' }}>
                          {order.customerName}
                        </p>
                        <p className="text-xs font-semibold text-right" style={{ color: 'var(--hub-text)' }}>
                          {formatCurrency(order.total)}
                        </p>
                        <div className="flex justify-end">
                          <span
                            className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize"
                            style={{ background: `${statusColor}18`, color: statusColor }}
                          >
                            {order.status}
                          </span>
                        </div>
                      </div>
                    );
                  })}
          </div>
        </motion.div>

        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <TrendingUp size={15} style={{ color: '#10b981' }} />
              <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                Top Products
              </h3>
            </div>
            <Link
              to="/products"
              className="flex items-center gap-1 text-xs"
              style={{ color: TEAL, textDecoration: 'none' }}
            >
              View all <ArrowRight size={11} />
            </Link>
          </div>
          <div className="space-y-1.5">
            {productsLoading
              ? Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={3} />)
              : (products ?? []).length === 0
                ? (
                    <p className="text-xs py-6 text-center" style={{ color: 'var(--hub-text-dim)' }}>
                      No products yet
                    </p>
                  )
                : (products ?? []).map((p, i) => (
                    <div
                      key={p.id}
                      className="flex items-center gap-3 px-3 py-2 rounded-lg"
                      style={{ background: 'var(--hub-hover)' }}
                    >
                      <span className="text-[10px] w-4 shrink-0" style={{ color: 'var(--hub-text-dim)' }}>
                        {i + 1}
                      </span>
                      <span className="text-xs flex-1 truncate" style={{ color: 'var(--hub-text)' }}>
                        {p.name}
                      </span>
                      <span className="text-xs font-medium shrink-0" style={{ color: 'var(--hub-text-muted)' }}>
                        {formatCurrency(p.price)}
                      </span>
                    </div>
                  ))}
          </div>
        </motion.div>
      </div>
    </motion.div>
  );
};

export default Dashboard;
