import React, { useMemo, useState } from 'react';
import { useParams, Link, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { useUserStore } from '@voxire/auth';
import {
  ArrowLeft,
  Edit3,
  Trash2,
  Package,
  PackageX,
  TrendingUp,
  Clock,
  Truck,
} from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { useDeleteProduct, useProductFull } from './useProducts';
import VariantMatrix from './components/VariantMatrix';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import StockBadge from '../../shared/components/StockBadge';
import BarcodeSvg from '../../shared/components/BarcodeSvg';
import QuickStockAdjust from '../../shared/components/QuickStockAdjust';
import { Skeleton } from '../../shared/components/Skeleton';
import QuickAddProductModal from '../../shared/components/QuickAddProductModal';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { formatCurrency } from '../../shared/lib/format';
import { useCategories } from '../categories/useCategories';
import { useBrands } from '../brands/useBrands';
import { canErpPermission } from '../../shared/lib/erpPermissions';
import { pageTitle } from '../../shared/lib/pageTitle';

const AMBER = 'var(--erp-accent)';

// Placeholder chart data - will be replaced with real API data
const stockHistory: number[] = [];
const salesHistory: number[] = [];

const MiniLineChart: React.FC<{ data: number[]; color: string; height?: number }> = ({
  data,
  color,
  height = 60,
}) => {
  // Guard the degenerate cases: empty data makes Math.max/min return
  // ±Infinity, and a single point makes (data.length - 1) a zero divisor → NaN.
  if (data.length === 0) {
    return <svg width="100%" height={height} viewBox="0 0 100 100" preserveAspectRatio="none" />;
  }
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = max - min || 1;
  const denom = data.length > 1 ? data.length - 1 : 1;
  const points = data
    .map((v, i) => {
      const x = (i / denom) * 100;
      const y = 100 - ((v - min) / range) * 100;
      return `${x},${y}`;
    })
    .join(' ');

  return (
    <svg width="100%" height={height} viewBox="0 0 100 100" preserveAspectRatio="none">
      <polyline
        fill="none"
        stroke={color}
        strokeWidth="2"
        vectorEffect="non-scaling-stroke"
        points={points}
      />
      <polyline
        fill={`color-mix(in srgb, ${color} 13%, transparent)`}
        stroke="none"
        points={`0,100 ${points} 100,100`}
      />
    </svg>
  );
};

const ProductDetail: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams({ strict: false });
  const productId = id || undefined;
  const { data: product, isLoading } = useProductFull(productId);
  const { data: categories = [] } = useCategories();
  const { data: brands = [] } = useBrands();
  const deleteProduct = useDeleteProduct();
  const [editing, setEditing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canDeleteProduct = canErpPermission(role, 'erp:products:delete');
  const canAdjustInventory = canErpPermission(role, 'erp:inventory:adjust');
  const categoryNameById = useMemo(() => {
    const map = new Map<string, string>();
    categories.forEach((category) => map.set(String(category.id), category.name ?? ''));
    return map;
  }, [categories]);
  const brandNameById = useMemo(() => {
    const map = new Map<string, string>();
    brands.forEach((brand) => map.set(String(brand.id), brand.name ?? ''));
    return map;
  }, [brands]);

  const handleDelete = (): void => {
    if (!product) return;
    deleteProduct.mutate(product.id ?? '', {
      onSuccess: () => {
        toast.success(t('erp.products.toast.deleted'));
        setConfirmDelete(false);
        navigate({ to: '/inventory' });
      },
      onError: () => {
        toast.error(t('erp.products.toast.deleteFailed'));
      },
    });
  };

  if (isLoading) {
    return (
      <div className="w-full min-w-0 space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!product) {
    return (
      <EmptyState
        icon={PackageX}
        title={t('erp.products.detail.notFound')}
        description={t('erp.products.detail.notFoundDescription')}
        action={{
          label: t('erp.products.detail.backToInventory'),
          onClick: () => navigate({ to: '/inventory' }),
        }}
      />
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t('nav.titles.productDetail'))} />
      <Breadcrumbs
        items={[{ label: 'Inventory', to: '/inventory' }, { label: product.name ?? '' }]}
      />

      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <Link
            to="/inventory"
            className="h-8 w-8 rounded-lg flex items-center justify-center"
            style={{
              background: 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              textDecoration: 'none',
            }}
          >
            <ArrowLeft size={14} />
          </Link>
          <div>
            <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
              {product.name}
            </h2>
            <p className="text-xs font-mono mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
              {product.sku}
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            data-testid="product-edit-button"
            onClick={() => setEditing(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
            style={{
              background: 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <Edit3 size={12} />
            {t('erp.products.detail.edit')}
          </button>
          {canDeleteProduct && (
            <button
              type="button"
              data-testid="product-delete-button"
              onClick={() => setConfirmDelete(true)}
              disabled={deleteProduct.isPending}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
              style={{
                color: 'var(--erp-danger)',
                border: '1px solid color-mix(in srgb, var(--erp-danger) 30%, transparent)',
                opacity: deleteProduct.isPending ? 0.6 : 1,
              }}
            >
              <Trash2 size={12} />
              {t('erp.products.detail.delete')}
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Main details */}
        <div className="lg:col-span-2 space-y-4">
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-start gap-4">
              <div
                className="h-24 w-24 rounded-lg flex items-center justify-center shrink-0"
                style={{ background: 'var(--hub-hover)' }}
              >
                <Package size={32} style={{ color: 'var(--hub-text-dim)' }} />
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <StockBadge
                    stock={product.stockQuantity ?? 0}
                    lowStock={product.lowStockThreshold ?? 0}
                  />
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                    style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                  >
                    {categoryNameById.get(product.categoryId ?? '') ||
                      `Category #${product.categoryId}`}
                  </span>
                  {product.brandId !== '0' && (
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{
                        background: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)',
                        color: AMBER,
                      }}
                    >
                      {brandNameById.get(product.brandId ?? '') || `Brand #${product.brandId}`}
                    </span>
                  )}
                </div>
                {product.description && (
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {product.description}
                  </p>
                )}
              </div>
            </div>
            <div
              className="grid grid-cols-4 gap-4 mt-5 pt-5"
              style={{ borderTop: '1px solid var(--hub-border)' }}
            >
              <div>
                <p
                  className="text-[10px] uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  {t('erp.products.costPrice')}
                </p>
                <p className="text-lg font-bold mt-0.5" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(product.costPrice ?? 0)}
                </p>
              </div>
              <div>
                <p
                  className="text-[10px] uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  {t('erp.products.salePrice')}
                </p>
                <p className="text-lg font-bold mt-0.5" style={{ color: AMBER }}>
                  {formatCurrency(product.salePrice ?? 0)}
                </p>
              </div>
              <div>
                <p
                  className="text-[10px] uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  {t('erp.products.margin')}
                </p>
                <p className="text-lg font-bold mt-0.5" style={{ color: 'var(--erp-success)' }}>
                  {(() => {
                    const salePrice = product.salePrice ?? 0;
                    const costPrice = product.costPrice ?? 0;
                    return salePrice > 0
                      ? `${Math.round(((salePrice - costPrice) / salePrice) * 100)}%`
                      : '—';
                  })()}
                </p>
              </div>
              <div>
                <p
                  className="text-[10px] uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  Stock
                </p>
                <div className="mt-0.5">
                  {canAdjustInventory ? (
                    <QuickStockAdjust
                      productId={product.id ?? ''}
                      currentStock={product.stockQuantity ?? 0}
                    />
                  ) : (
                    <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>
                      {product.stockQuantity ?? 0}
                    </p>
                  )}
                </div>
              </div>
            </div>
          </div>

          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center justify-between mb-3">
              <h3
                className="text-sm font-semibold flex items-center gap-2"
                style={{ color: 'var(--hub-text)' }}
              >
                <TrendingUp size={14} style={{ color: AMBER }} />
                Stock History (12 weeks)
              </h3>
            </div>
            {stockHistory.length > 0 ? (
              <MiniLineChart data={stockHistory} color={AMBER} height={80} />
            ) : (
              <p className="text-xs py-4 text-center" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.products.noStockHistory')}
              </p>
            )}
          </div>

          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center justify-between mb-3">
              <h3
                className="text-sm font-semibold flex items-center gap-2"
                style={{ color: 'var(--hub-text)' }}
              >
                <TrendingUp size={14} style={{ color: 'var(--erp-success)' }} />
                Sales History (7 days)
              </h3>
            </div>
            {salesHistory.length > 0 ? (
              <MiniLineChart data={salesHistory} color="var(--erp-success)" height={80} />
            ) : (
              <p className="text-xs py-4 text-center" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.products.noSalesHistory')}
              </p>
            )}
          </div>

          <VariantMatrix
            productId={product.id ?? ''}
            productSku={product.sku ?? ''}
            baseSalePrice={product.salePrice ?? 0}
          />
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-xs font-semibold mb-3" style={{ color: 'var(--hub-text-muted)' }}>
              {t('erp.products.barcode')}
            </h3>
            <div className="rounded-lg overflow-hidden">
              <BarcodeSvg value={(product.barcode || product.sku) ?? ''} width={240} height={72} />
            </div>
          </div>

          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3
              className="text-xs font-semibold mb-3 flex items-center gap-2"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <Truck size={12} /> SUPPLIER INFO
            </h3>
            <div className="space-y-2 text-xs">
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-dim)' }}>{t('erp.products.leadTime')}</span>
                <span style={{ color: 'var(--hub-text)' }}>{product.leadTimeDays || 7} days</span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-dim)' }}>
                  {t('erp.products.reorderPoint')}
                </span>
                <span style={{ color: 'var(--hub-text)' }}>
                  {t('erp.products.unitsCount', {
                    count: product.reorderQuantity || product.lowStockThreshold,
                  })}
                </span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-dim)' }}>{t('erp.products.unit')}</span>
                <span style={{ color: 'var(--hub-text)' }}>{product.unit || 'pcs'}</span>
              </div>
            </div>
          </div>

          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3
              className="text-xs font-semibold mb-3 flex items-center gap-2"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <Clock size={12} /> ACTIVITY
            </h3>
            <p className="text-xs py-2 text-center" style={{ color: 'var(--hub-text-dim)' }}>
              {t('erp.products.detail.noActivity')}
            </p>
          </div>
        </div>
      </div>

      <QuickAddProductModal open={editing} onClose={() => setEditing(false)} editing={product} />
      <ConfirmModal
        open={confirmDelete}
        title={t('erp.products.detail.delete')}
        danger
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </motion.div>
  );
};

export default ProductDetail;
