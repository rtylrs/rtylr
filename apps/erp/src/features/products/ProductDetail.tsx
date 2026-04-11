import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { motion } from 'motion/react';
import { ArrowLeft, Edit3, Trash2, Package, TrendingUp, Clock, Tag, Truck } from 'lucide-react';
import { useProduct } from './useProducts';
import { useVariants } from './useVariants';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import StockBadge from '../../shared/components/StockBadge';
import BarcodeSvg from '../../shared/components/BarcodeSvg';
import QuickStockAdjust from '../../shared/components/QuickStockAdjust';
import EmptyState from '../../shared/components/EmptyState';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatCurrency } from '../../shared/lib/format';

const AMBER = '#f59e0b';

// Placeholder chart data - will be replaced with real API data
const stockHistory: number[] = [];
const salesHistory: number[] = [];

const MiniLineChart: React.FC<{ data: number[]; color: string; height?: number }> = ({
  data,
  color,
  height = 60,
}) => {
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = max - min || 1;
  const points = data
    .map((v, i) => {
      const x = (i / (data.length - 1)) * 100;
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
      <polyline fill={`${color}22`} stroke="none" points={`0,100 ${points} 100,100`} />
    </svg>
  );
};

const ProductDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const productId = id ? parseInt(id) : undefined;
  const { data: product, isLoading } = useProduct(productId);
  const { data: variants = [] } = useVariants(productId);

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-5xl">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!product) {
    return (
      <EmptyState
        title="Product not found"
        description="The product you are looking for does not exist or has been deleted."
        action={
          <Link
            to="/products"
            className="px-3 py-2 rounded-lg text-xs font-semibold"
            style={{ background: AMBER, color: '#000', textDecoration: 'none' }}
          >
            Back to Products
          </Link>
        }
      />
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
      <Breadcrumbs items={[{ label: 'Products', to: '/products' }, { label: product.name }]} />

      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <Link
            to="/products"
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
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
            style={{
              background: 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <Edit3 size={12} />
            Edit
          </button>
          <button
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium"
            style={{ color: '#f87171', border: '1px solid rgba(248,113,113,0.3)' }}
          >
            <Trash2 size={12} />
            Delete
          </button>
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
                  <StockBadge stock={product.stock} lowStock={product.lowStock} />
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                    style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                  >
                    {product.category}
                  </span>
                  {product.brand && (
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{ background: 'rgba(245,158,11,0.1)', color: AMBER }}
                    >
                      {product.brand}
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
                  Cost Price
                </p>
                <p className="text-lg font-bold mt-0.5" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(product.costPrice)}
                </p>
              </div>
              <div>
                <p
                  className="text-[10px] uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  Sale Price
                </p>
                <p className="text-lg font-bold mt-0.5" style={{ color: AMBER }}>
                  {formatCurrency(product.salePrice)}
                </p>
              </div>
              <div>
                <p
                  className="text-[10px] uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  Margin
                </p>
                <p className="text-lg font-bold mt-0.5" style={{ color: '#10b981' }}>
                  {Math.round(((product.salePrice - product.costPrice) / product.salePrice) * 100)}%
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
                  <QuickStockAdjust productId={product.id} currentStock={product.stock} />
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
                No stock history data yet.
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
                <TrendingUp size={14} style={{ color: '#10b981' }} />
                Sales History (7 days)
              </h3>
            </div>
            {salesHistory.length > 0 ? (
              <MiniLineChart data={salesHistory} color="#10b981" height={80} />
            ) : (
              <p className="text-xs py-4 text-center" style={{ color: 'var(--hub-text-dim)' }}>
                No sales history data yet.
              </p>
            )}
          </div>

          {variants.length > 0 && (
            <div
              className="rounded-xl p-5"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <h3
                className="text-sm font-semibold mb-3 flex items-center gap-2"
                style={{ color: 'var(--hub-text)' }}
              >
                <Tag size={14} style={{ color: AMBER }} /> Variants
              </h3>
              <div className="space-y-2">
                {variants.map((v) => (
                  <div
                    key={v.id}
                    className="flex items-center justify-between p-2 rounded-lg"
                    style={{ background: 'var(--hub-hover)' }}
                  >
                    <div>
                      <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                        {v.name}
                      </p>
                      <p className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                        {v.sku}
                      </p>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                        {v.stock} in stock
                      </span>
                      <span className="text-sm font-semibold" style={{ color: AMBER }}>
                        {formatCurrency(v.price)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-xs font-semibold mb-3" style={{ color: 'var(--hub-text-muted)' }}>
              BARCODE
            </h3>
            <div className="rounded-lg overflow-hidden">
              <BarcodeSvg value={product.barcode || product.sku} width={240} height={72} />
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
                <span style={{ color: 'var(--hub-text-dim)' }}>Lead Time</span>
                <span style={{ color: 'var(--hub-text)' }}>{product.leadTimeDays || 7} days</span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-dim)' }}>Reorder Point</span>
                <span style={{ color: 'var(--hub-text)' }}>
                  {product.reorderPoint || product.lowStock} units
                </span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-dim)' }}>Unit</span>
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
            <div className="space-y-2 text-xs">
              {[
                { text: 'Stock adjusted +5', when: '2h ago' },
                { text: 'Price updated', when: '1d ago' },
                { text: 'Product created', when: '5d ago' },
              ].map((a, i) => (
                <div key={i} className="flex justify-between">
                  <span style={{ color: 'var(--hub-text)' }}>{a.text}</span>
                  <span style={{ color: 'var(--hub-text-dim)' }}>{a.when}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </motion.div>
  );
};

export default ProductDetail;
