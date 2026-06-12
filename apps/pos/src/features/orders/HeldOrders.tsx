import React from 'react';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import { useHeldOrdersStore } from '../../shared/stores/heldOrdersStore';
import { useCartStore } from '../../shared/stores/cartStore';
import { formatDate, formatCurrency } from '../../shared/utils/format';
import { posService } from '../../shared/services/api/posService';
import {
  buildHeldOrderCartProducts,
  buildServerHeldOrderResume,
  normalizeServerHeldOrders,
  type ServerHeldOrder,
} from './heldOrdersWorkflow';

const HeldOrdersScreen: React.FC = () => {
  const orders = useHeldOrdersStore((s) => s.orders);
  const release = useHeldOrdersStore((s) => s.release);
  const remove = useHeldOrdersStore((s) => s.remove);
  const navigate = useNavigate();
  const addItem = useCartStore((s) => s.addItem);
  const setTable = useCartStore((s) => s.setTableNumber);
  const setServerOrderId = useCartStore((s) => s.setServerOrderId);

  const { data: serverOrders = [] } = useQuery<ServerHeldOrder[]>({
    queryKey: ['pos', 'held-orders'],
    queryFn: async () => normalizeServerHeldOrders(await unwrap(posService.getHeldOrders())),
    refetchInterval: 15_000,
  });

  const restore = (id: string) => {
    const existingOrder = orders.find((order) => order.id === id);
    if (!existingOrder) return;

    let products: ReturnType<typeof buildHeldOrderCartProducts>;
    try {
      products = buildHeldOrderCartProducts(existingOrder);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to release held order');
      return;
    }

    const order = release(id);
    if (!order) return;
    products.forEach(addItem);
    setTable(order.tableNumber ?? null);
    toast.success(`Released ${order.label}`);
    navigate('/');
  };

  const resumeServerOrder = (o: ServerHeldOrder): void => {
    let resume: ReturnType<typeof buildServerHeldOrderResume>;
    try {
      resume = buildServerHeldOrderResume(o);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to resume held order');
      return;
    }

    setServerOrderId(resume.orderId);
    resume.products.forEach(addItem);
    toast.success(`Resumed order #${o.orderNumber}`);
    navigate('/');
  };

  const isEmpty = orders.length === 0 && serverOrders.length === 0;

  return (
    <ScreenLayout title="Held Orders">
      <div className="p-5 max-w-3xl space-y-4">
        {isEmpty ? (
          <EmptyState icon="⏸" title="No held orders" />
        ) : (
          <>
            {serverOrders.length > 0 && (
              <div>
                <p className="text-xs font-semibold uppercase tracking-widest text-pos-muted mb-2">
                  Server
                </p>
                <ul className="space-y-2">
                  {serverOrders.map((o) => (
                    <li
                      key={`sv-${o.id}`}
                      className="p-4 rounded-xl bg-pos-card border border-pos-border flex items-center justify-between"
                    >
                      <div>
                        <div className="text-sm font-semibold text-pos-text">#{o.orderNumber}</div>
                        <div className="text-[11px] text-pos-muted">
                          {o.customerName ?? 'Guest'} · {formatCurrency(o.total)} ·{' '}
                          {formatDate(o.createdAt)}
                        </div>
                      </div>
                      <button
                        onClick={() => resumeServerOrder(o)}
                        className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg font-semibold text-sm"
                      >
                        Resume
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {orders.length > 0 && (
              <div>
                <p className="text-xs font-semibold uppercase tracking-widest text-pos-muted mb-2">
                  Local
                </p>
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
                        {o.note && (
                          <div className="text-xs text-pos-muted italic mt-1">{o.note}</div>
                        )}
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
              </div>
            )}
          </>
        )}
      </div>
    </ScreenLayout>
  );
};

export default HeldOrdersScreen;
