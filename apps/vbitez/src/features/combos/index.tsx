import React from 'react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useCombos } from './useCombos';
import { useCartStore } from '../../shared/stores/cartStore';
import { formatCurrency } from '../../shared/utils/format';

const CombosScreen: React.FC = () => {
  const { combos, loading, error } = useCombos();
  const addItem = useCartStore((s) => s.addItem);

  const addCombo = (id: string, name: string, price: number) => {
    addItem({ id: `combo_${id}`, name, price });
    toast.success(`Added combo: ${name}`);
  };

  return (
    <ScreenLayout title="Combos" subtitle="Bundle builder">
      <div className="p-5">
        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState icon="⚠" title="Failed to load" message={error} />
        ) : combos.length === 0 ? (
          <EmptyState icon="🍔" title="No combos yet" message="Create combos to bundle items." />
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {combos.map((c) => (
              <div
                key={c.id}
                className="rounded-xl bg-pos-card border border-pos-border p-4 flex flex-col gap-3"
              >
                <div>
                  <h3 className="text-base font-bold text-pos-text">{c.name}</h3>
                  {c.description && (
                    <p className="text-xs text-pos-muted mt-1">{c.description}</p>
                  )}
                  <p className="text-lg font-bold text-pos-text mt-2">{formatCurrency(c.price)}</p>
                </div>
                <div className="space-y-1">
                  {c.items.map((it, idx) => (
                    <div key={idx} className="text-xs text-pos-muted">
                      {it.qty}× {it.name ?? it.itemId}
                    </div>
                  ))}
                </div>
                <button
                  onClick={() => addCombo(c.id, c.name, c.price)}
                  className="mt-auto h-11 rounded-lg bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 transition"
                >
                  Add to Order
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default CombosScreen;
