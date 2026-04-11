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
  const padding = size === 'lg' ? 'p-4 min-h-[7rem]' : size === 'sm' ? 'p-2 min-h-[4rem]' : 'p-3 min-h-[5.5rem]';
  return (
    <button
      onClick={() => !unavailable && onClick(item)}
      disabled={unavailable}
      className={`text-left ${padding} rounded-xl bg-pos-card border border-pos-border hover:border-white/25 transition active:scale-[0.97] flex flex-col justify-between ${
        unavailable ? 'opacity-50 cursor-not-allowed' : ''
      }`}
    >
      {item.image && (
        <img
          src={item.image}
          alt={item.name}
          loading="lazy"
          className="w-full aspect-square object-cover rounded-lg mb-1.5"
        />
      )}
      <div className="flex-1">
        <p className="text-sm font-semibold text-pos-text leading-tight line-clamp-2">
          {item.name}
        </p>
        <p className="text-xs text-pos-muted mt-1">{formatCurrency(item.price)}</p>
      </div>
      {unavailable && (
        <span className="text-[10px] text-destructive font-semibold mt-1">Out of stock</span>
      )}
    </button>
  );
};

export default ItemTile;
