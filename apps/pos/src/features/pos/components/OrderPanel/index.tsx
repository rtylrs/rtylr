import React, { useCallback, useMemo, useState } from "react";
import { CheckCircle2, ReceiptText, X } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "motion/react";
import { toast } from "sonner";
import { createOneKitClient } from "@voxire/api";
import { OrderServiceClient } from "@voxire/voxie/pos/orders/v1/client";
import { useCartStore } from "../../../../shared/stores/cartStore";
import { useHeldOrdersStore } from "../../../../shared/stores/heldOrdersStore";
import { useSettingsStore } from "../../../../shared/stores/settingsStore";
import { useTerminalContextStore } from "../../../../shared/stores/terminalContextStore";
import ConfirmModal from "../../../../shared/components/ConfirmModal";
import QuickNoteModal from "../../../../shared/components/QuickNoteModal";
import OfflineBadge from "../../../../shared/components/OfflineBadge";
import { formatCurrency } from "../../../../shared/utils/format";
import {
  buildOrderSubmitPayload,
  extractCreatedOrderId,
} from "./orderSubmitWorkflow";
import {
  buildOrderCompatPayload,
  buildResumeOrderPayload,
  createOrderCompat,
  requiresCompatOrder,
  persistResumedOrder,
} from "../../../orders/orderCompatWorkflow";
import { resolveOpenSessionId } from "../../../close-session/closeSessionWorkflow";

const orderClient = createOneKitClient(OrderServiceClient, "pos");

interface OrderPanelProps {
  tableNumber?: string;
}

interface OrderLineItemProps {
  id: string;
  name: string;
  quantity: number;
  price: number;
  notes: string;
  onUpdateQuantity: (id: string, quantity: number) => void;
  onRemove: (id: string) => void;
  onAddNote: (id: string) => void;
  disabled?: boolean;
}

// Memoized so editing one line (quantity/notes) doesn't re-render every row.
const OrderLineItem = React.memo<OrderLineItemProps>(function OrderLineItem({
  id,
  name,
  quantity,
  price,
  notes,
  onUpdateQuantity,
  onRemove,
  onAddNote,
  disabled = false,
}) {
  return (
    <motion.div
      className="flex flex-col gap-1 p-2.5 rounded-lg bg-white/5"
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      layout
    >
      <div className="flex items-center justify-between gap-2">
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-pos-text truncate">{name}</p>
          <p className="text-xs text-pos-muted">
            {formatCurrency(price * quantity)}
          </p>
        </div>
        <div className="flex items-center gap-1 shrink-0">
          <button
            disabled={disabled}
            className="w-7 h-7 flex items-center justify-center rounded bg-white/10 text-pos-text text-sm disabled:opacity-40"
            onClick={() => onUpdateQuantity(id, quantity - 1)}
          >
            −
          </button>
          <span className="w-6 text-center text-xs font-semibold text-pos-text">
            {quantity}
          </span>
          <button
            disabled={disabled}
            className="w-7 h-7 flex items-center justify-center rounded bg-white/10 text-pos-text text-sm disabled:opacity-40"
            onClick={() => onUpdateQuantity(id, quantity + 1)}
          >
            +
          </button>
          <button
            disabled={disabled}
            className="w-6 h-6 flex items-center justify-center rounded text-pos-muted hover:text-destructive transition-colors text-xs ml-1 disabled:opacity-40"
            onClick={() => onRemove(id)}
          >
            <X className="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </div>
      </div>
      <div className="flex items-center justify-between">
        <button
          disabled={disabled}
          onClick={() => onAddNote(id)}
          className="text-[10px] text-pos-muted hover:text-pos-text disabled:opacity-40"
        >
          {notes ? `Note: ${notes}` : "+ Add note"}
        </button>
      </div>
    </motion.div>
  );
});

const OrderPanel: React.FC<OrderPanelProps> = ({ tableNumber }) => {
  const navigate = useNavigate();
  // Slice selectors: the panel must not re-render on unrelated cart writes
  // (branchId/kitchenNotes/persist hydration).
  const items = useCartStore((s) => s.items);
  const updateQuantity = useCartStore((s) => s.updateQuantity);
  const removeItem = useCartStore((s) => s.removeItem);
  const clearCart = useCartStore((s) => s.clearCart);
  const updateItemNotes = useCartStore((s) => s.updateItemNotes);
  const subtotalFn = useCartStore((s) => s.subtotal);
  const taxTotalFn = useCartStore((s) => s.taxTotal);
  const totalFn = useCartStore((s) => s.total);
  const orderDiscount = useCartStore((s) => s.orderDiscount);
  const tipAmount = useCartStore((s) => s.tipAmount);
  const serverOrderId = useCartStore((s) => s.serverOrderId);
  const tableId = useCartStore((s) => s.tableId);
  const orderType = useCartStore((s) => s.orderType);
  const notes = useCartStore((s) => s.notes);
  const kitchenNotes = useCartStore((s) => s.kitchenNotes);
  const customerId = useCartStore((s) => s.customerId);
  const customerName = useCartStore((s) => s.customerName);
  const hold = useHeldOrdersStore((s) => s.hold);
  const tablesEnabled = useSettingsStore((s) => s.features.tables);
  const locationId = useTerminalContextStore((s) => s.locationId);

  const subtotal = useMemo(
    () => subtotalFn(),
    [subtotalFn, items, orderDiscount],
  );
  const tax = useMemo(() => taxTotalFn(), [taxTotalFn, items, orderDiscount]);
  const grandTotal = useMemo(
    () => totalFn(),
    [totalFn, items, orderDiscount, tipAmount],
  );
  const activeTableId = tablesEnabled ? tableId : undefined;
  const activeTableNumber = tablesEnabled ? tableNumber : undefined;

  const [confirmClear, setConfirmClear] = useState(false);
  const [notingItemId, setNotingItemId] = useState<string | null>(null);
  const [submittingOrder, setSubmittingOrder] = useState(false);

  const handleUpdateQuantity = useCallback(
    (id: string, quantity: number) => {
      if (!submittingOrder) updateQuantity(id, quantity);
    },
    [submittingOrder, updateQuantity],
  );
  const handleRemove = useCallback(
    (id: string) => {
      if (!submittingOrder) removeItem(id);
    },
    [removeItem, submittingOrder],
  );
  const handleAddNote = useCallback(
    (id: string) => {
      if (!submittingOrder) setNotingItemId(id);
    },
    [submittingOrder],
  );

  const onHold = () => {
    if (items.length === 0 || submittingOrder) return;
    if (serverOrderId) {
      toast.error("Resumed orders cannot be held from the order panel");
      return;
    }
    hold({
      label: activeTableNumber
        ? `Table ${activeTableNumber}`
        : `Cart ${new Date().toLocaleTimeString()}`,
      note: notes,
      orderDiscount,
      tipAmount,
      items: items.map((i) => ({
        id: i.lineId,
        productId: i.comboId ?? i.productId,
        name: i.name,
        price: i.unitPrice,
        quantity: i.quantity,
        sku: i.sku,
        barcode: i.barcode,
        notes: i.notes,
        taxRate: i.taxRate,
        modifiers: i.modifiers,
        comboId: i.comboId,
        variantName: i.variantName,
        discountPct: i.discountPct,
        discountAmount: i.discountAmount,
      })),
      tableNumber: activeTableNumber,
    });
    clearCart();
    toast.success("Order held");
  };

  const notingItem = notingItemId
    ? items.find((i) => i.id === notingItemId)
    : null;

  const handleCompleteOrder = async (): Promise<void> => {
    if (items.length === 0 || submittingOrder) return;

    const submittedItems = items;
    const submittedNotes = notes;
    const submittedKitchenNotes = kitchenNotes;
    const submittedOrderDiscount = orderDiscount;
    const submittedTipAmount = tipAmount;
    const submittedCustomerId = customerId;
    const submittedCustomerName = customerName;
    const submittedTableId = tableId;
    const submittedOrderType = orderType;
    setSubmittingOrder(true);
    try {
      if (serverOrderId) {
        await persistResumedOrder(
          buildResumeOrderPayload({
            orderId: serverOrderId,
            items: submittedItems,
            locationId,
            tableId: activeTableId,
            customerId: submittedCustomerId,
            customerName: submittedCustomerName,
            orderType: submittedOrderType,
            notes: submittedNotes,
            kitchenNotes: submittedKitchenNotes,
          }),
        );
      } else {
        const sessionId = await resolveOpenSessionId();
        if (requiresCompatOrder(submittedItems)) {
          const created = await createOrderCompat(
            buildOrderCompatPayload({
              items: submittedItems,
              sessionId,
              locationId,
              tableId: activeTableId,
              customerId: submittedCustomerId,
              customerName: submittedCustomerName,
              orderType: submittedOrderType,
              notes: submittedNotes,
              kitchenNotes: submittedKitchenNotes,
            }),
          );
          extractCreatedOrderId({ orderId: String(created.orderId ?? "") });
        } else {
          const created = await orderClient.createOrder(
            buildOrderSubmitPayload({
              items: submittedItems,
              sessionId,
              locationId,
              tableId: activeTableId,
              customerId: submittedCustomerId,
              customerName: submittedCustomerName,
              orderType: submittedOrderType,
              notes: submittedNotes,
              kitchenNotes: submittedKitchenNotes,
            }),
          );
          extractCreatedOrderId(created);
        }
      }

      const current = useCartStore.getState();
      const cartChanged =
        current.items !== submittedItems ||
        current.notes !== submittedNotes ||
        current.kitchenNotes !== submittedKitchenNotes ||
        current.orderDiscount !== submittedOrderDiscount ||
        current.tipAmount !== submittedTipAmount ||
        current.customerId !== submittedCustomerId ||
        current.customerName !== submittedCustomerName ||
        current.tableId !== submittedTableId ||
        current.orderType !== submittedOrderType;
      if (cartChanged) {
        toast.success("Order saved; newer cart changes were kept");
      } else {
        clearCart();
        toast.success("Order saved to Orders");
      }
      navigate("/orders");
    } catch (error) {
      toast.error(
        error instanceof Error && error.message
          ? error.message
          : "Failed to complete order",
      );
    } finally {
      setSubmittingOrder(false);
    }
  };

  return (
    <div className="w-80 shrink-0 flex flex-col bg-pos-card border-l border-pos-border h-full">
      <div className="flex items-start justify-between px-4 py-3.5 border-b border-pos-border shrink-0 gap-2">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold text-pos-text">Current Order</h2>
          {activeTableNumber && (
            <span className="text-xs text-pos-muted">
              Table {activeTableNumber}
            </span>
          )}
          <div className="mt-1">
            <OfflineBadge />
          </div>
        </div>
        {items.length > 0 && (
          <button
            disabled={submittingOrder}
            className="text-xs text-pos-muted hover:text-pos-text transition-colors disabled:opacity-40"
            onClick={() => setConfirmClear(true)}
          >
            Clear
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        <AnimatePresence>
          {items.length === 0 ? (
            <motion.div
              className="flex flex-col items-center justify-center h-48 gap-2 text-pos-muted"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
            >
              <ReceiptText
                className="h-8 w-8"
                aria-hidden="true"
                strokeWidth={1.8}
              />
              <p className="text-sm font-medium">No items added yet</p>
              <span className="text-xs">Tap a product to add it</span>
            </motion.div>
          ) : (
            items.map((item) => (
              <OrderLineItem
                key={item.id}
                id={item.id}
                name={item.name}
                quantity={item.quantity}
                price={item.price}
                notes={item.notes ?? ""}
                onUpdateQuantity={handleUpdateQuantity}
                onRemove={handleRemove}
                onAddNote={handleAddNote}
                disabled={submittingOrder}
              />
            ))
          )}
        </AnimatePresence>
      </div>

      {items.length > 0 && (
        <div className="border-t border-pos-border p-4 space-y-3 shrink-0">
          <div className="space-y-1.5 text-sm">
            <div className="flex justify-between text-pos-muted">
              <span>Subtotal</span>
              <span>{formatCurrency(subtotal)}</span>
            </div>
            <div className="flex justify-between text-pos-muted">
              <span>Tax</span>
              <span>{formatCurrency(tax)}</span>
            </div>
            <div className="flex justify-between font-bold text-pos-text pt-1 border-t border-pos-border">
              <span>Total</span>
              <span>{formatCurrency(grandTotal)}</span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <button
              disabled={submittingOrder}
              className="h-12 rounded-xl bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20 transition disabled:opacity-40"
              onClick={onHold}
            >
              Hold
            </button>
            <button
              disabled={submittingOrder}
              className="h-12 rounded-xl bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20 transition disabled:opacity-40"
              onClick={() => navigate("/split-bill/current")}
            >
              Split Pay
            </button>
          </div>
          <button
            className="w-full h-11 rounded-xl bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20 transition disabled:opacity-40 inline-flex items-center justify-center gap-2"
            onClick={handleCompleteOrder}
            disabled={submittingOrder}
            title={
              serverOrderId ? "Save changes to this resumed order" : undefined
            }
          >
            {!submittingOrder && (
              <CheckCircle2 className="h-4 w-4" aria-hidden="true" />
            )}
            {submittingOrder ? "Saving…" : "Complete Order"}
          </button>
          <button
            disabled={submittingOrder}
            className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 transition-opacity active:scale-[0.98] disabled:opacity-40"
            onClick={() => navigate("/payment")}
          >
            Charge {formatCurrency(grandTotal)}
          </button>
        </div>
      )}

      <ConfirmModal
        open={confirmClear}
        title="Clear order?"
        message="All items will be removed from the current order."
        danger
        confirmText="Clear"
        onConfirm={() => {
          if (!submittingOrder) clearCart();
          setConfirmClear(false);
        }}
        onCancel={() => setConfirmClear(false)}
      />

      <QuickNoteModal
        open={!!notingItem}
        initialValue={notingItem?.notes ?? ""}
        title={`Note for ${notingItem?.name ?? ""}`}
        onCancel={() => setNotingItemId(null)}
        onSave={(val) => {
          if (notingItemId && !submittingOrder)
            updateItemNotes(notingItemId, val);
          setNotingItemId(null);
        }}
      />
    </div>
  );
};

export default OrderPanel;
