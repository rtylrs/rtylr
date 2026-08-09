import React from 'react';

interface StockBadgeProps {
  stock: number;
  lowStock?: number;
  excessThreshold?: number;
  showLabel?: boolean;
}

const AMBER = '#f59e0b';

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
    out: { color: '#f87171', bg: 'rgba(248,113,113,0.12)', label: 'Out of Stock' },
    low: { color: AMBER, bg: 'rgba(245,158,11,0.12)', label: 'Low Stock' },
    good: { color: '#10b981', bg: 'rgba(16,185,129,0.12)', label: 'In Stock' },
    excess: { color: '#60a5fa', bg: 'rgba(96,165,250,0.12)', label: 'Excess' },
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
