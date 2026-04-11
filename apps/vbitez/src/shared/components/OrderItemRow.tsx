import React from 'react';
import type { OrderLineItem } from '../types/pos';
import { formatCurrency } from '../utils/format';
import QuantityStepper from './QuantityStepper';

interface OrderItemRowProps {
  item: OrderLineItem;
  onQtyChange: (qty: number) => void;
  onRemove: () => void;
  onNoteClick?: () => void;
  readonly?: boolean;
}

const OrderItemRow: React.FC<OrderItemRowProps> = ({
  item,
  onQtyChange,
  onRemove,
  onNoteClick,
  readonly,
}) => {
  const lineTotal = item.price * item.quantity;
  return (
    <div className="flex items-start gap-3 p-3 rounded-lg bg-white/5 border border-pos-border">
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-pos-text truncate">{item.name}</p>
        {item.modifiers && item.modifiers.length > 0 && (
          <p className="text-[10px] text-pos-muted">
            {item.modifiers.map((m) => m.name).join(', ')}
          </p>
        )}
        {item.notes && <p className="text-[10px] italic text-pos-muted">Note: {item.notes}</p>}
        <p className="text-xs text-pos-muted mt-0.5">{formatCurrency(lineTotal)}</p>
      </div>
      <div className="flex flex-col items-end gap-1.5">
        {!readonly && <QuantityStepper value={item.quantity} onChange={onQtyChange} />}
        {!readonly && onNoteClick && (
          <button onClick={onNoteClick} className="text-[10px] text-pos-muted hover:text-pos-text">
            Note
          </button>
        )}
        {!readonly && (
          <button
            onClick={onRemove}
            className="text-[10px] text-destructive hover:underline"
            aria-label="Remove item"
          >
            Remove
          </button>
        )}
      </div>
    </div>
  );
};

export default OrderItemRow;
