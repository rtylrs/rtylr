import React from 'react';
import { motion, type Variants } from 'motion/react';
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
  Tags,
} from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { useProducts } from '../products/useProducts';
import { useActivity } from '../../shared/hooks/useActivity';
import { timeAgo, formatCurrency } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';
import QuickAddProductModal from '../../shared/components/QuickAddProductModal';
import { useDashboardStats } from '../../shared/hooks/queries/useDashboardStats';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';
import { useTranslation } from 'react-i18next';

const AMBER = 'var(--erp-accent)';
const CARD_STYLE: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
};
const CARD_CLASS = 'flex h-full min-h-0 flex-col rounded-lg p-4';
const SCROLL_AREA_CLASS = 'min-h-0 flex-1 overflow-y-auto pr-1';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

function getStatusColor(status: string): string {
  const normalized = status.toLowerCase();
  if (normalized === 'received') return 'var(--erp-success)';
  if (normalized === 'sent' || normalized === 'approved' || normalized === 'partial') return AMBER;
  if (normalized === 'pending_approval') return 'var(--erp-purple)';
  return 'var(--hub-text-dim)';
}

function getStatusBg(status: string): string {
  const normalized = status.toLowerCase();
  if (normalized === 'received') return 'color-mix(in srgb, var(--erp-success) 9%, transparent)';
  if (normalized === 'sent' || normalized === 'approved' || normalized === 'partial')
    return 'color-mix(in srgb, var(--erp-accent) 9%, transparent)';
  if (normalized === 'pending_approval')
    return 'color-mix(in srgb, var(--erp-purple) 9%, transparent)';
  return 'color-mix(in srgb, var(--hub-text) 6%, transparent)';
}

function formatStatus(status: string): string {
  return status
    .replace(/_/g, ' ')
    .split(' ')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

const activityIconFor = (type: string) => {
  if (type === 'order') return ShoppingCart;
  if (type === 'product') return Package;
  if (type === 'transfer') return TrendingUp;
  if (type === 'supplier') return Users;
  return Activity;
};

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const [quickAddOpen, setQuickAddOpen] = React.useState(false);
  const { data: dashboardStats, isLoading: statsLoading } = useDashboardStats();
  const { data: productsResult, isLoading: productsLoading } = useProducts({ pageSize: 5 });
  const productsData = productsResult?.products ?? [];
  const { data: activityData = [] } = useActivity(8);

  const isLoading = statsLoading || productsLoading;
  const lowStockItems = dashboardStats?.lowStockItems ?? [];
  const recentOrders = (dashboardStats?.recentOrders ?? []).slice(0, 3);
  const topCategories = dashboardStats?.topCategories ?? [];
  const criticalLowStockCount = lowStockItems.filter((i) => (i.currentStock ?? 0) <= 5).length;

  const stats = [
    {
      id: 'totalProducts',
      label: t('erp.dashboard.totalProducts'),
      value: dashboardStats?.totalProducts ?? 0,
      icon: Package,
      change: '',
      changePositive: true,
      to: '/inventory',
    },
    {
      id: 'lowStockItems',
      label: t('erp.dashboard.lowStockItems'),
      value: dashboardStats?.lowStockCount ?? 0,
      icon: AlertTriangle,
      change: t('erp.dashboard.criticalCount', { count: criticalLowStockCount }),
      changePositive: false,
      to: '/alerts',
    },
    {
      id: 'pendingOrders',
      label: t('erp.dashboard.pendingOrders'),
      value: dashboardStats?.pendingOrders ?? 0,
      icon: ShoppingCart,
      change: t('erp.dashboard.needsReview'),
      changePositive: true,
      to: '/purchasing',
    },
    {
      id: 'totalSuppliers',
      label: t('erp.dashboard.totalSuppliers'),
      value: dashboardStats?.totalSuppliers ?? 0,
      icon: Users,
      change: '',
      changePositive: true,
      to: '/suppliers',
    },
  ];

  const inventoryValue =
    dashboardStats?.inventoryValue ??
    (productsData || [])
      .slice(0, 5)
      .reduce((s, p) => s + (p.salePrice ?? 0) * (p.stockQuantity || 0), 0);

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="w-full min-w-0 space-y-6"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/'].title))} />
      <motion.div
        variants={itemVariants}
        className="flex flex-wrap items-start justify-between gap-4"
      >
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('erp.dashboard.inventoryOverview')}
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {today}
          </p>
        </div>
        <button
          type="button"
          onClick={() => setQuickAddOpen(true)}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{
            background: AMBER,
            color: 'var(--erp-accent-foreground)',
            textDecoration: 'none',
          }}
        >
          <PackagePlus size={14} />
          {t('erp.dashboard.newProduct')}
        </button>
      </motion.div>

      <motion.div
        variants={containerVariants}
        className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
        style={{ gridAutoRows: '168px' }}
      >
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <motion.div key={stat.id} variants={itemVariants} className="h-full">
              <Link
                to={stat.to}
                className={`${CARD_CLASS} justify-between transition-all`}
                style={{ ...CARD_STYLE, textDecoration: 'none' }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.borderColor =
                    'color-mix(in srgb, var(--erp-accent) 30%, transparent)';
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
                    style={{ background: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)' }}
                  >
                    <Icon size={14} style={{ color: AMBER }} />
                  </span>
                </div>
                <div>
                  <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                    {isLoading ? '…' : stat.value}
                  </p>
                  {stat.change && (
                    <p
                      className="text-xs mt-1"
                      style={{
                        color: stat.changePositive ? 'var(--erp-success)' : 'var(--erp-danger)',
                      }}
                    >
                      {stat.change}
                    </p>
                  )}
                </div>
              </Link>
            </motion.div>
          );
        })}

        <motion.div variants={itemVariants} className={CARD_CLASS} style={CARD_STYLE}>
          <div className="flex items-center gap-2">
            <TrendingUp size={14} style={{ color: 'var(--erp-success)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.dashboard.inventoryValue')}
            </h3>
          </div>
          <div className="mt-auto">
            <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {isLoading ? '…' : formatCurrency(inventoryValue)}
            </p>
            <p className="text-[11px] mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              active tracked stock at cost
            </p>
          </div>
        </motion.div>

        <motion.div variants={itemVariants} className={CARD_CLASS} style={CARD_STYLE}>
          <div className="flex items-center gap-2 mb-3">
            <Package size={14} style={{ color: AMBER }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.dashboard.topItems')}
            </h3>
          </div>
          <div className={`${SCROLL_AREA_CLASS} space-y-1.5`}>
            {productsData.slice(0, 5).map((p, i) => (
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
            {!productsData.length && (
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.dashboard.noProducts')}
              </p>
            )}
          </div>
        </motion.div>

        <motion.div variants={itemVariants} className={CARD_CLASS} style={CARD_STYLE}>
          <div className="flex items-center gap-2 mb-3">
            <Tags size={14} style={{ color: AMBER }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.dashboard.topCategories')}
            </h3>
          </div>
          <div className={`${SCROLL_AREA_CLASS} space-y-1.5`}>
            {topCategories.slice(0, 5).map((category, i) => (
              <div key={category.name || i} className="flex items-center gap-2">
                <span className="text-[10px] w-4" style={{ color: 'var(--hub-text-dim)' }}>
                  {i + 1}
                </span>
                <span
                  className="text-xs flex-1 truncate"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  {category.name}
                </span>
                <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                  {category.productCount}
                </span>
              </div>
            ))}
            {!topCategories.length && (
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.dashboard.noCategories')}
              </p>
            )}
          </div>
        </motion.div>

        <motion.div variants={itemVariants} className={CARD_CLASS} style={CARD_STYLE}>
          <div className="flex items-center gap-2 mb-3">
            <Activity size={14} style={{ color: 'var(--erp-info)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.dashboard.recentActivity')}
            </h3>
          </div>
          <div className={`${SCROLL_AREA_CLASS} space-y-2`}>
            {activityData.slice(0, 5).map((a) => {
              const Icon = activityIconFor(a.entityType ?? '');
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
                      {a.description}
                    </p>
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {timeAgo(a.createdAt ?? '')}
                    </p>
                  </div>
                </div>
              );
            })}
            {!activityData.length && (
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.dashboard.noRecentActivity')}
              </p>
            )}
          </div>
        </motion.div>

        <motion.div variants={itemVariants} className={CARD_CLASS} style={CARD_STYLE}>
          <div className="flex items-center gap-2 mb-3">
            <TrendingDown size={15} style={{ color: 'var(--erp-danger)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.dashboard.lowStockAlert')}
            </h3>
            <span
              className="ml-auto text-xs px-2 py-0.5 rounded-full font-medium"
              style={{
                background: 'color-mix(in srgb, var(--erp-danger) 10%, transparent)',
                color: 'var(--erp-danger)',
              }}
            >
              {isLoading ? '…' : `${dashboardStats?.lowStockCount ?? 0} items`}
            </span>
          </div>
          <div className={`${SCROLL_AREA_CLASS} space-y-1`}>
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
                      style={{
                        color: (item.currentStock ?? 0) <= 5 ? 'var(--erp-danger)' : AMBER,
                      }}
                    >
                      {item.currentStock}
                    </p>
                    <p className="text-xs text-right" style={{ color: 'var(--hub-text-muted)' }}>
                      {item.reorderPoint}
                    </p>
                  </div>
                ))}
            {!isLoading && !lowStockItems.length && (
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.dashboard.noLowStockItems')}
              </p>
            )}
          </div>
        </motion.div>

        <motion.div variants={itemVariants} className={CARD_CLASS} style={CARD_STYLE}>
          <div className="flex items-center gap-2 mb-3">
            <Clock size={15} style={{ color: AMBER }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.dashboard.recentPurchaseOrders')}
            </h3>
          </div>
          <div className={`${SCROLL_AREA_CLASS} space-y-1`}>
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={4} />)
              : recentOrders.map((order) => {
                  const statusColor = getStatusColor(order.status ?? '');
                  const statusBg = getStatusBg(order.status ?? '');
                  return (
                    <div
                      key={order.id}
                      className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg"
                      style={{ background: 'var(--hub-hover)' }}
                    >
                      <div>
                        <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                          {order.poNumber}
                        </p>
                        <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                          {new Date(order.createdAt ?? '').toLocaleDateString('en-US')}
                        </p>
                      </div>
                      <p
                        className="text-xs truncate pr-2"
                        style={{ color: 'var(--hub-text-muted)' }}
                      >
                        {order.supplierName}
                      </p>
                      <p
                        className="text-xs font-semibold text-right"
                        style={{ color: 'var(--hub-text)' }}
                      >
                        {formatCurrency(order.total ?? 0)}
                      </p>
                      <div className="flex justify-end">
                        <span
                          className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                          style={{ background: statusBg, color: statusColor }}
                        >
                          {formatStatus(order.status ?? '')}
                        </span>
                      </div>
                    </div>
                  );
                })}
            {!isLoading && !recentOrders.length && (
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.dashboard.noPurchaseOrders')}
              </p>
            )}
          </div>
        </motion.div>
      </motion.div>
      <QuickAddProductModal open={quickAddOpen} onClose={() => setQuickAddOpen(false)} />
    </motion.div>
  );
};

export default Dashboard;
