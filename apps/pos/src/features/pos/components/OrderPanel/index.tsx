import React, { useState } from 'react';
import { ReceiptText, X } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import { useCartStore } from '../../../../shared/stores/cartStore';
import { useHeldOrdersStore } from '../../../../shared/stores/heldOrdersStore';
import { useSettingsStore } from '../../../../shared/stores/settingsStore';
import { useTerminalContextStore } from '../../../../shared/stores/terminalContextStore';
import ConfirmModal from '../../../../shared/components/ConfirmModal';
import QuickNoteModal from '../../../../shared/components/QuickNoteModal';
import OfflineBadge from '../../../../shared/components/OfflineBadge';
import { posService } from '../../../../shared/services/api/posService';
import { formatCurrency } from '../../../../shared/utils/format';
import { buildKitchenOrderPayload, extractCreatedOrderId } from './kitchenWorkflow';

interface OrderPanelProps {
  tableNumber?: string;
}

const OrderPanel: React.FC<OrderPanelProps> = ({ tableNumber }) => {
  const navigate = useNavigate();
  const {
    items,
    updateQuantity,
    removeItem,
    clearCart,
    updateItemNotes,
    total,
    serverOrderId,
    setServerOrderId,
  } = useCartStore();
  const hold = useHeldOrdersStore((s) => s.hold);
  const taxRate = useSettingsStore((s) => s.taxRate);
  const locationId = useTerminalContextStore((s) => s.locationId);
  const customerId = useCartStore((s) => s.customerId);
  const customerName = useCartStore((s) => s.customerName);

  const subtotal = total();
  const tax = subtotal * taxRate;
  const grandTotal = subtotal + tax;

  const [confirmClear, setConfirmClear] = useState(false);
  const [notingItemId, setNotingItemId] = useState<string | null>(null);
  const [sendingToKitchen, setSendingToKitchen] = useState(false);

  const onHold = () => {
    if (items.length === 0) return;
    hold({
      label: tableNumber ? `Table ${tableNumber}` : `Cart ${new Date().toLocaleTimeString()}`,
      items: items.map((i) => ({
        id: i.id,
        name: i.name,
        price: i.price,
        quantity: i.quantity,
        notes: i.notes,
      })),
      tableNumber,
    });
    clearCart();
    toast.success('Order held');
  };

  const notingItem = notingItemId ? items.find((i) => i.id === notingItemId) : null;

  const handleSendToKitchen = async (): Promise<void> => {
    if (items.length === 0) return;
    setSendingToKitchen(true);
    try {
      let orderId = serverOrderId;
      if (!orderId) {
        const created = await unwrap(
          posService.createOrder(
            buildKitchenOrderPayload({
              items,
              locationId,
              customerId,
              customerName,
            })
          )
        );
        orderId = extractCreatedOrderId(created);
        setServerOrderId(orderId);
      }
      await unwrap(posService.sendToKitchen(orderId));
      toast.success('Sent to kitchen');
    } catch {
      toast.error('Failed to send to kitchen');
    } finally {
      setSendingToKitchen(false);
    }
  };

  return (
    <div className="w-80 shrink-0 flex flex-col bg-pos-card border-l border-pos-border h-full">
      <div className="flex items-start justify-between px-4 py-3.5 border-b border-pos-border shrink-0 gap-2">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold text-pos-text">Current Order</h2>
          {tableNumber && <span className="text-xs text-pos-muted">Table {tableNumber}</span>}
          <div className="mt-1">
            <OfflineBadge />
          </div>
        </div>
        {items.length > 0 && (
          <button
            className="text-xs text-pos-muted hover:text-pos-text transition-colors"
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
              <ReceiptText className="h-8 w-8" aria-hidden="true" strokeWidth={1.8} />
              <p className="text-sm font-medium">No items added yet</p>
              <span className="text-xs">Tap a product to add it</span>
            </motion.div>
          ) : (
            items.map((item) => (
              <motion.div
                key={item.id}
                className="flex flex-col gap-1 p-2.5 rounded-lg bg-white/5"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                layout
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-pos-text truncate">{item.name}</p>
                    <p className="text-xs text-pos-muted">
                      {formatCurrency(item.price * item.quantity)}
                    </p>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <button
                      className="w-7 h-7 flex items-center justify-center rounded bg-white/10 text-pos-text text-sm"
                      onClick={() => updateQuantity(item.id, item.quantity - 1)}
                    >
                      −
                    </button>
                    <span className="w-6 text-center text-xs font-semibold text-pos-text">
                      {item.quantity}
                    </span>
                    <button
                      className="w-7 h-7 flex items-center justify-center rounded bg-white/10 text-pos-text text-sm"
                      onClick={() => updateQuantity(item.id, item.quantity + 1)}
                    >
                      +
                    </button>
                    <button
                      className="w-6 h-6 flex items-center justify-center rounded text-pos-muted hover:text-destructive transition-colors text-xs ml-1"
                      onClick={() => removeItem(item.id)}
                    >
                      <X className="h-3.5 w-3.5" aria-hidden="true" />
                    </button>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <button
                    onClick={() => setNotingItemId(item.id)}
                    className="text-[10px] text-pos-muted hover:text-pos-text"
                  >
                    {item.notes ? `Note: ${item.notes}` : '+ Add note'}
                  </button>
                </div>
              </motion.div>
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
              <span>Tax ({Math.round(taxRate * 100)}%)</span>
              <span>{formatCurrency(tax)}</span>
            </div>
            <div className="flex justify-between font-bold text-pos-text pt-1 border-t border-pos-border">
              <span>Total</span>
              <span>{formatCurrency(grandTotal)}</span>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <button
              className="h-12 rounded-xl bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20 transition"
              onClick={onHold}
            >
              Hold
            </button>
            <button
              className="h-12 rounded-xl bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20 transition"
              onClick={() => navigate('/split-bill/current')}
            >
              Split
            </button>
          </div>
          <button
            className="w-full h-11 rounded-xl bg-orange-500/20 text-orange-300 font-semibold text-sm hover:bg-orange-500/30 transition disabled:opacity-40"
            onClick={handleSendToKitchen}
            disabled={sendingToKitchen}
          >
            {sendingToKitchen ? 'Sending…' : 'Send to Kitchen'}
          </button>
          <button
            className="w-full h-14 rounded-xl bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 transition-opacity active:scale-[0.98]"
            onClick={() => navigate('/payment')}
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
          clearCart();
          setConfirmClear(false);
        }}
        onCancel={() => setConfirmClear(false)}
      />

      <QuickNoteModal
        open={!!notingItem}
        initialValue={notingItem?.notes ?? ''}
        title={`Note for ${notingItem?.name ?? ''}`}
        onCancel={() => setNotingItemId(null)}
        onSave={(val) => {
          if (notingItemId) updateItemNotes(notingItemId, val);
          setNotingItemId(null);
        }}
      />
    </div>
  );
};

export default OrderPanel;
