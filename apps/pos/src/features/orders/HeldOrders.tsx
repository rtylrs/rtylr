import React, { useState } from "react";
import { CirclePause } from "lucide-react";
import { toast } from "sonner";
import { useNavigate } from "react-router-dom";
import { createOneKitClient } from "@voxire/api";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { OrderHoldServiceClient } from "@voxire/voxie/pos/orders/holds/v1/client";
import type { Order as ServerHeldOrder } from "@voxire/voxie/pos/orders/v1/types";
import ScreenLayout from "../../layouts/ScreenLayout";
import EmptyState from "../../shared/components/EmptyState";
import ConfirmModal from "../../shared/components/ConfirmModal";
import { useHeldOrdersStore } from "../../shared/stores/heldOrdersStore";
import { useCartStore } from "../../shared/stores/cartStore";
import { formatDate, formatCurrency } from "../../shared/utils/format";
import {
  buildHeldOrderCartProducts,
  buildServerHeldOrderResume,
} from "./heldOrdersWorkflow";
import {
  buildResumeOrderPayload,
  persistResumedOrder,
} from "./orderCompatWorkflow";

const orderHoldClient = createOneKitClient(OrderHoldServiceClient, "pos");

const HeldOrdersScreen: React.FC = () => {
  const orders = useHeldOrdersStore((s) => s.orders);
  const release = useHeldOrdersStore((s) => s.release);
  const remove = useHeldOrdersStore((s) => s.remove);
  const navigate = useNavigate();
  const replaceItemsFromProducts = useCartStore(
    (s) => s.replaceItemsFromProducts,
  );
  const setTableNumber = useCartStore((s) => s.setTableNumber);
  const setServerOrderId = useCartStore((s) => s.setServerOrderId);
  const setCustomer = useCartStore((s) => s.setCustomer);
  const setOrderDiscount = useCartStore((s) => s.setOrderDiscount);
  const setTip = useCartStore((s) => s.setTip);
  const setNotes = useCartStore((s) => s.setNotes);
  const setKitchenNotes = useCartStore((s) => s.setKitchenNotes);
  const setOrderType = useCartStore((s) => s.setOrderType);

  const queryClient = useQueryClient();
  const [discarding, setDiscarding] = useState<string | null>(null);
  const [resuming, setResuming] = useState<number | null>(null);
  const { data: serverOrders = [] } = useQuery<ServerHeldOrder[]>({
    queryKey: ["pos", "held-orders"],
    queryFn: async () => {
      const response = await orderHoldClient.listHeldOrders({});
      return response.orders ?? [];
    },
    refetchInterval: 15_000,
    refetchIntervalInBackground: false,
  });

  const restore = (id: string) => {
    if (useCartStore.getState().items.length > 0) {
      toast.error(
        "Hold or clear the current order before restoring another one",
      );
      return;
    }
    const existingOrder = orders.find((order) => order.id === id);
    if (!existingOrder) return;

    let products: ReturnType<typeof buildHeldOrderCartProducts>;
    try {
      products = buildHeldOrderCartProducts(existingOrder);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to release held order",
      );
      return;
    }

    // Replace the current cart atomically. Adding into an existing cart can
    // merge same-product lines and silently discard held notes or quantities.
    replaceItemsFromProducts(products);
    const order = release(id);
    if (!order) return;
    setTableNumber(order.tableNumber ?? null);
    setOrderDiscount(order.orderDiscount ?? null);
    setTip(order.tipAmount ?? 0);
    setNotes(order.note ?? "");
    setKitchenNotes("");
    setServerOrderId(null);
    toast.success(`Released ${order.label}`);
    navigate("/");
  };

  const resumeServerOrder = async (o: ServerHeldOrder): Promise<void> => {
    if (resuming !== null) return;
    if (useCartStore.getState().items.length > 0) {
      toast.error(
        "Hold or clear the current order before resuming another one",
      );
      return;
    }
    let resume: ReturnType<typeof buildServerHeldOrderResume>;
    try {
      resume = buildServerHeldOrderResume(o);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to resume held order",
      );
      return;
    }

    setResuming(resume.orderId);
    try {
      // The server operation both validates the tenant and atomically moves
      // held -> pending. Do not expose a local cart until that write succeeds.
      await persistResumedOrder(
        buildResumeOrderPayload({
          orderId: resume.orderId,
          items: resume.products,
          locationId: resume.locationId,
          tableId: resume.tableId,
          orderType: resume.orderType,
          customerId: resume.customerId,
          customerName: resume.customerName,
          notes: resume.notes,
          kitchenNotes: resume.kitchenNotes,
        }),
      );
      replaceItemsFromProducts(resume.products);
      setServerOrderId(resume.orderId);
      setTableNumber(resume.tableId != null ? String(resume.tableId) : null);
      setCustomer(resume.customerId ?? null, resume.customerName ?? null);
      setOrderDiscount(
        resume.discountAmount != null
          ? { type: "fixed", value: resume.discountAmount }
          : null,
      );
      setNotes(resume.notes ?? "");
      setKitchenNotes(resume.kitchenNotes ?? "");
      setOrderType(resume.orderType ?? "takeaway");
      await queryClient.invalidateQueries({ queryKey: ["pos", "held-orders"] });
      toast.success(`Resumed order #${o.orderNumber}`);
      navigate("/");
    } catch (error) {
      toast.error(
        error instanceof Error && error.message
          ? error.message
          : "Failed to resume held order",
      );
    } finally {
      setResuming(null);
    }
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
                        <div className="text-sm font-semibold text-pos-text">
                          #{o.orderNumber}
                        </div>
                        <div className="text-[11px] text-pos-muted">
                          {o.customerName ?? "Guest"} ·{" "}
                          {formatCurrency(o.total ?? 0)} ·{" "}
                          {formatDate(o.createdAt ?? "")}
                        </div>
                      </div>
                      <button
                        onClick={() => void resumeServerOrder(o)}
                        disabled={resuming !== null}
                        className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg font-semibold text-sm disabled:opacity-40"
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
                        <div className="text-sm font-semibold text-pos-text">
                          {o.label}
                        </div>
                        <div className="text-[11px] text-pos-muted">
                          {o.items.length} items · {formatDate(o.createdAt)}
                          {o.tableNumber && <> · Table {o.tableNumber}</>}
                        </div>
                        {o.note && (
                          <div className="text-xs text-pos-muted italic mt-1">
                            {o.note}
                          </div>
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
