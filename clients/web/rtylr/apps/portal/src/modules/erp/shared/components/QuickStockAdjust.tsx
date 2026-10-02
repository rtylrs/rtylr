import React, { useState } from 'react';
import { Minus, Plus } from 'lucide-react';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
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
      {/* The adjust endpoint only accepts change > 0, so decreases stay
          disabled until the contract supports signed adjustments. */}
      <button
        type="button"
        disabled
        aria-label="Decrease stock"
        title={t('erp.inventory.decreaseUnsupported')}
        className="h-11 w-11 rounded flex items-center justify-center"
        style={{ color: 'var(--hub-text-muted)', opacity: 0.4 }}
      >
        <Minus size={16} />
      </button>
      <span
        className="text-xs font-semibold min-w-[32px] text-center"
        style={{ color: 'var(--hub-text)' }}
      >
        {stock}
      </span>
      <button
        type="button"
        onClick={() => bump(1)}
        aria-label="Increase stock"
        className="h-11 w-11 rounded flex items-center justify-center"
        style={{ color: 'var(--hub-text-muted)' }}
      >
        <Plus size={16} />
      </button>
    </div>
  );
};

export default QuickStockAdjust;
