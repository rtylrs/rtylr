import React from "react";
import { AlertTriangle, ShoppingBag } from "lucide-react";
import { toast } from "sonner";
import ScreenLayout from "../../layouts/ScreenLayout";
import EmptyState from "../../shared/components/EmptyState";
import Skeleton from "../../shared/components/Skeleton";
import { useCombos } from "./useCombos";
import { useCartStore } from "../../shared/stores/cartStore";
import { formatCurrency } from "../../shared/utils/format";
import { buildComboCartProduct } from "../../shared/utils/cart/cartProduct";

const CombosScreen: React.FC = () => {
  const { combos, loading, error } = useCombos();
  const addItem = useCartStore((s) => s.addItem);

  const addCombo = (combo: (typeof combos)[number]) => {
    try {
      const product = buildComboCartProduct(combo);
      // The cart line key has no product-kind field. Use the otherwise-unused
      // variant slot to keep a combo distinct from a menu item with the same ID.
      addItem({ ...product, variantName: "combo" });
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to add combo",
      );
      return;
    }
    toast.success(`Added combo: ${combo.name}`);
  };

  return (
    <ScreenLayout title="Combos" subtitle="Bundle builder">
      <div className="p-5">
        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState
            icon={AlertTriangle}
            title="Failed to load"
            message={error}
          />
        ) : combos.length === 0 ? (
          <EmptyState
            icon={ShoppingBag}
            title="No combos yet"
            message="Create combos to bundle items."
          />
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {combos.map((c) => (
              <div
                key={c.id}
                className="rounded-xl bg-pos-card border border-pos-border p-4 flex flex-col gap-3"
              >
                <div>
                  <h3 className="text-base font-bold text-pos-text">
                    {c.name}
                  </h3>
                  {c.description && (
                    <p className="text-xs text-pos-muted mt-1">
                      {c.description}
                    </p>
                  )}
                  <p className="text-lg font-bold text-pos-text mt-2">
                    {formatCurrency(c.price)}
                  </p>
                </div>
                <div className="space-y-1">
                  {c.items.map((it, idx) => (
                    <div key={idx} className="text-xs text-pos-muted">
                      {it.qty}× {it.name ?? it.itemId}
                    </div>
                  ))}
                </div>
                <button
                  onClick={() => addCombo(c)}
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
