import React, { useState } from 'react';
import { CirclePause } from 'lucide-react';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import { createOneKitClient } from '@voxire/api';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { OrderHoldServiceClient } from '@voxire/voxie/pos/orders/holds/v1/client';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import type { Order as ServerHeldOrder } from '@voxire/voxie/pos/orders/v1/types';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { useHeldOrdersStore } from '../../shared/stores/heldOrdersStore';
import { useCartStore } from '../../shared/stores/cartStore';
import { formatDate, formatCurrency } from '../../shared/utils/format';
import {
  buildHeldOrderCartProducts,
  buildServerHeldOrderResume,
} from './heldOrdersWorkflow';

const orderHoldClient = createOneKitClient(OrderHoldServiceClient, 'pos');
const orderClient = createOneKitClient(OrderServiceClient, 'pos');

const HeldOrdersScreen: React.FC = () => {
  const orders = useHeldOrdersStore((s) => s.orders);
  const release = useHeldOrdersStore((s) => s.release);
  const remove = useHeldOrdersStore((s) => s.remove);
  const navigate = useNavigate();
  const addItem = useCartStore((s) => s.addItem);
  const setTable = useCartStore((s) => s.setTableNumber);
  const setServerOrderId = useCartStore((s) => s.setServerOrderId);

  const queryClient = useQueryClient();
  const [discarding, setDiscarding] = useState<string | null>(null);
  const { data: serverOrders = [] } = useQuery<ServerHeldOrder[]>({
    queryKey: ['pos', 'held-orders'],
    queryFn: async () => {
      const response = await orderHoldClient.listHeldOrders({});
      return response.orders ?? [];
    },
    refetchInterval: 15_000,
    refetchIntervalInBackground: false,
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

    // Add items BEFORE releasing the hold: if an add fails mid-loop the hold
    // survives and can be retried, instead of vanishing with a partial cart.
    try {
      products.forEach(addItem);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to restore held items');
      return;
    }
    const order = release(id);
    if (!order) return;
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
    // Transition the order out of 'held' so it leaves the held list and
    // cannot be resumed again on this or another terminal.
    void orderClient
      .updateOrderStatus({ orderId: String(resume.orderId), body: { status: 'pending' } })
      .then(() => {
        void queryClient.invalidateQueries({ queryKey: ['pos', 'held-orders'] });
      })
      .catch(() => {
        toast.warning('Resumed locally - the server hold could not be released');
      });
    toast.success(`Resumed order #${o.orderNumber}`);
    navigate('/');
  };

  const isEmpty = orders.length === 0 && serverOrders.length === 0;

  return (
    <ScreenLayout title="Held Orders">
      <div className="p-5 max-w-3xl space-y-4">
        {isEmpty ? (
          <EmptyState icon={CirclePause} title="No held orders" />
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
                          {o.customerName ?? 'Guest'} · {formatCurrency(o.total ?? 0)} ·{' '}
                          {formatDate(o.createdAt ?? '')}
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
                          onClick={() => setDiscarding(o.id)}
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
      <ConfirmModal
        open={discarding !== null}
        title="Discard held order?"
        message="The held items will be permanently removed from this terminal."
        confirmText="Discard"
        danger
        onCancel={() => setDiscarding(null)}
        onConfirm={() => {
          if (discarding) remove(discarding);
          setDiscarding(null);
        }}
      />
    </ScreenLayout>
  );
};

export default HeldOrdersScreen;
