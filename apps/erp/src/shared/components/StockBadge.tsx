import React from 'react';

interface StockBadgeProps {
  stock: number;
  lowStock?: number;
  excessThreshold?: number;
  showLabel?: boolean;
}

const AMBER = 'var(--erp-accent)';

export function getStockLevel(
  stock: number,
  lowStock = 10,
  excessThreshold = 500
): 'out' | 'low' | 'good' | 'excess' {
  if (stock <= 0) return 'out';
  if (stock < lowStock) return 'low';
  if (stock > excessThreshold) return 'excess';
  return 'good';
}

const StockBadge: React.FC<StockBadgeProps> = ({
  stock,
  lowStock = 10,
  excessThreshold = 500,
  showLabel = true,
}) => {
  const level = getStockLevel(stock, lowStock, excessThreshold);
  const cfg = {
    out: {
      color: 'var(--erp-danger)',
      bg: 'color-mix(in srgb, var(--erp-danger) 12%, transparent)',
      label: 'Out of Stock',
    },
    low: {
      color: AMBER,
      bg: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
      label: 'Low Stock',
    },
    good: {
      color: 'var(--erp-success)',
      bg: 'color-mix(in srgb, var(--erp-success) 12%, transparent)',
      label: 'In Stock',
    },
    excess: {
      color: 'var(--erp-info)',
      bg: 'color-mix(in srgb, var(--erp-info) 12%, transparent)',
      label: 'Excess',
    },
  }[level];

  return (
    <span
      className="inline-flex items-center gap-1.5 text-[10px] font-medium px-2 py-0.5 rounded-full"
      style={{ background: cfg.bg, color: cfg.color }}
    >
      <span className="h-1.5 w-1.5 rounded-full" style={{ background: cfg.color }} />
      {showLabel ? cfg.label : stock}
    </span>
  );
};

export default StockBadge;
