import React from 'react';
import type { MenuItem } from '../types/pos';
import { formatCurrency } from '../utils/format';

interface ItemTileProps {
  item: MenuItem;
  onClick: (item: MenuItem) => void;
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
}

const ItemTile: React.FC<ItemTileProps> = ({ item, onClick, size = 'md', disabled }) => {
  const unavailable = disabled || item.available === false;
  const showStock = typeof item.stock === 'number' && Number.isFinite(item.stock);
  const stockLabel =
    showStock &&
    (Number.isInteger(item.stock)
      ? String(item.stock)
      : item.stock.toFixed(3).replace(/\.?0+$/, ''));
  const padding =
    size === 'lg' ? 'p-4 min-h-[7rem]' : size === 'sm' ? 'p-2 min-h-[4rem]' : 'p-3 min-h-[5.5rem]';
  return (
    <button
      onClick={() => !unavailable && onClick(item)}
      disabled={unavailable}
      className={`text-left ${padding} rounded-xl bg-pos-card border border-pos-border hover:border-white/25 transition active:scale-[0.97] flex flex-col justify-between ${
        unavailable ? 'opacity-50 cursor-not-allowed' : ''
      }`}
    >
      {item.image ? (
        <img
          src={item.image}
          alt={item.name}
          loading="lazy"
          className="w-full aspect-square object-contain rounded-lg bg-white p-2 mb-1.5"
        />
      ) : (
        <div className="mb-1.5 flex w-full aspect-square items-center justify-center rounded-lg bg-white/6 text-sm font-semibold text-pos-muted">
          {item.name.slice(0, 1).toUpperCase()}
        </div>
      )}
      <div className="flex-1">
        <p className="text-sm font-semibold text-pos-text leading-tight line-clamp-2">
          {item.name}
        </p>
        {item.sku && (
          <p className="mt-1 truncate text-[10px] font-medium uppercase tracking-wide text-pos-muted/80">
            {item.sku}
          </p>
        )}
        <p className="text-xs text-pos-muted mt-1">{formatCurrency(item.price)}</p>
        {showStock && (
          <p className="mt-1 text-[10px] font-semibold text-pos-muted">{stockLabel} in stock</p>
        )}
      </div>
      {unavailable && (
        <span className="text-[10px] text-destructive font-semibold mt-1">Out of stock</span>
      )}
    </button>
  );
};

export default ItemTile;
