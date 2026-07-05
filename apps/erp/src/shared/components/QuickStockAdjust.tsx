import React, { useState } from 'react';
import { Minus, Plus } from 'lucide-react';
import { useAdjustInventory } from '../../features/inventory/useInventory';
import { toast } from 'sonner';

interface QuickStockAdjustProps {
  productId: string;
  currentStock: number;
  onChange?: (newStock: number) => void;
}

const QuickStockAdjust: React.FC<QuickStockAdjustProps> = ({
  productId,
  currentStock,
  onChange,
}) => {
  const [stock, setStock] = useState(currentStock);
  const adjust = useAdjustInventory();

  const bump = (delta: number) => {
    const prev = stock;
    const next = Math.max(0, stock + delta);
    setStock(next);
    adjust.mutate(
      { productId, change: delta, type: 'adjustment', notes: 'Quick stock adjustment' },
      {
        onSuccess: () => {
          toast.success(`Stock adjusted by ${delta > 0 ? '+' : ''}${delta}`);
          onChange?.(next);
        },
        onError: () => {
          // Roll back the optimistic update so the UI matches the backend.
          setStock(prev);
          toast.error('Failed to adjust stock');
        },
      }
    );
  };

  return (
    <div
      className="inline-flex items-center gap-1 rounded-lg p-0.5"
      style={{ background: 'var(--hub-hover)' }}
    >
      <button
        onClick={() => bump(-1)}
        className="h-6 w-6 rounded flex items-center justify-center"
        style={{ color: 'var(--hub-text-muted)' }}
      >
        <Minus size={12} />
      </button>
      <span
        className="text-xs font-semibold min-w-[32px] text-center"
        style={{ color: 'var(--hub-text)' }}
      >
        {stock}
      </span>
      <button
        onClick={() => bump(1)}
        className="h-6 w-6 rounded flex items-center justify-center"
        style={{ color: 'var(--hub-text-muted)' }}
      >
        <Plus size={12} />
      </button>
    </div>
  );
};

export default QuickStockAdjust;
