import React from 'react';
import { Link } from '@tanstack/react-router';
import { Package } from 'lucide-react';
import { cdnImageUrl } from '@voxire/config';
import type { Product } from '../../features/products/useProducts';
import { formatCurrency } from '../lib/format';
import StockBadge from './StockBadge';

interface ProductCardProps {
  product: Product;
  selected?: boolean;
  onToggle?: (id: string) => void;
  compact?: boolean;
}

const ProductCard: React.FC<ProductCardProps> = ({ product, selected, onToggle, compact }) => {
  const imageUrl = cdnImageUrl((product as Product & { imageUrl?: string }).imageUrl);
  return (
    <div
      className="rounded-xl p-3 transition-colors"
      style={{
        background: 'var(--hub-card)',
        border: selected
          ? '1px solid color-mix(in srgb, var(--erp-accent) 50%, transparent)'
          : '1px solid var(--hub-border)',
      }}
    >
      <div className="flex items-start gap-3">
        {onToggle && (
          <input
            type="checkbox"
            checked={!!selected}
            onChange={() => onToggle(product.id ?? '')}
            className="mt-1 cursor-pointer"
          />
        )}
        <div
          className="flex items-center justify-center h-12 w-12 rounded-lg shrink-0"
          style={{ background: 'var(--hub-hover)' }}
        >
          {imageUrl ? (
            <img
              src={imageUrl}
              alt={product.name}
              className="h-12 w-12 rounded-lg bg-white p-1 object-contain"
            />
          ) : (
            <Package size={18} style={{ color: 'var(--hub-text-dim)' }} />
          )}
        </div>
        <div className="flex-1 min-w-0">
          <Link
            to="/products/$id"
            params={{ id: String(product.id) }}
            className="text-sm font-medium truncate block hover:underline"
            style={{ color: 'var(--hub-text)', textDecoration: 'none' }}
          >
            {product.name}
          </Link>
          <p className="text-[11px] font-mono truncate" style={{ color: 'var(--hub-text-dim)' }}>
            {product.sku}
          </p>
          {!compact && (
            <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
              Category #{product.categoryId}
            </p>
          )}
        </div>
      </div>
      <div className="flex items-center justify-between mt-3">
        <StockBadge stock={product.stockQuantity ?? 0} lowStock={product.lowStockThreshold} />
        <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
          {formatCurrency(product.salePrice ?? 0)}
        </span>
      </div>
    </div>
  );
};

export default ProductCard;
