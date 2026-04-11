import React from 'react';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import { useHeldOrdersStore } from '../../shared/stores/heldOrdersStore';
import { useCartStore } from '../../shared/stores/cartStore';
import { formatDate } from '../../shared/utils/format';

const HeldOrdersScreen: React.FC = () => {
  const orders = useHeldOrdersStore((s) => s.orders);
  const release = useHeldOrdersStore((s) => s.release);
  const remove = useHeldOrdersStore((s) => s.remove);
  const navigate = useNavigate();
  const addItem = useCartStore((s) => s.addItem);
  const setTable = useCartStore((s) => s.setTableNumber);

  const restore = (id: string) => {
    const order = release(id);
    if (!order) return;
    order.items.forEach((it) => addItem({ id: it.id, name: it.name, price: it.price }));
    setTable(order.tableNumber ?? null);
    toast.success(`Released ${order.label}`);
    navigate('/');
  };

  return (
    <ScreenLayout title="Held Orders">
      <div className="p-5 max-w-3xl">
        {orders.length === 0 ? (
          <EmptyState icon="⏸" title="No held orders" />
        ) : (
          <ul className="space-y-2">
            {orders.map((o) => (
              <li
                key={o.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between"
              >
                <div>
                  <div className="text-sm font-semibold text-pos-text">{o.label}</div>
                  <div className="text-[11px] text-pos-muted">
                    {o.items.length} items · {formatDate(o.createdAt)}
                    {o.tableNumber && <> · Table {o.tableNumber}</>}
                  </div>
                  {o.note && <div className="text-xs text-pos-muted italic mt-1">{o.note}</div>}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => restore(o.id)}
                    className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg font-semibold text-sm"
                  >
                    Release
                  </button>
                  <button
                    onClick={() => remove(o.id)}
                    className="h-10 px-4 rounded-lg bg-destructive text-white font-semibold text-sm"
                  >
                    Discard
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </ScreenLayout>
  );
};

export default HeldOrdersScreen;
