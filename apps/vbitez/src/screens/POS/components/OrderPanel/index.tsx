import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useCartStore } from '../../../../stores/cartStore';

interface OrderPanelProps { tableNumber?: string; }

const fmt = (n: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const OrderPanel: React.FC<OrderPanelProps> = ({ tableNumber }) => {
  const navigate = useNavigate();
  const { items, updateQuantity, removeItem, clearCart, total } = useCartStore();
  const subtotal = total();
  const tax = subtotal * 0.08;
  const grandTotal = subtotal + tax;

  return (
    <div className="w-80 shrink-0 flex flex-col bg-pos-card border-l border-pos-border h-full">
      {/* Header */}
      <div className="flex items-start justify-between px-4 py-3.5 border-b border-pos-border shrink-0">
        <div>
          <h2 className="text-sm font-semibold text-pos-text">Current Order</h2>
          {tableNumber && <span className="text-xs text-pos-muted">Table {tableNumber}</span>}
        </div>
        {items.length > 0 && (
          <button className="text-xs text-pos-muted hover:text-pos-text transition-colors" onClick={clearCart}>Clear</button>
        )}
      </div>

      {/* Items */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        <AnimatePresence>
          {items.length === 0 ? (
            <motion.div className="flex flex-col items-center justify-center h-48 gap-2 text-pos-muted" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
              <span className="text-3xl">🧾</span>
              <p className="text-sm font-medium">No items added yet</p>
              <span className="text-xs">Tap a product to add it</span>
            </motion.div>
          ) : items.map((item) => (
            <motion.div
              key={item.id}
              className="flex items-center justify-between gap-2 p-2.5 rounded-lg bg-white/5"
              initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }} layout
            >
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-pos-text truncate">{item.name}</p>
                <p className="text-xs text-pos-muted">{fmt(item.price * item.quantity)}</p>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                <button className="w-6 h-6 flex items-center justify-center rounded bg-white/8 text-pos-text hover:bg-white/14 transition-colors text-sm" onClick={() => updateQuantity(item.id, item.quantity - 1)}>−</button>
                <span className="w-6 text-center text-xs font-semibold text-pos-text">{item.quantity}</span>
                <button className="w-6 h-6 flex items-center justify-center rounded bg-white/8 text-pos-text hover:bg-white/14 transition-colors text-sm" onClick={() => updateQuantity(item.id, item.quantity + 1)}>+</button>
                <button className="w-5 h-5 flex items-center justify-center rounded text-pos-muted hover:text-destructive transition-colors text-xs ml-1" onClick={() => removeItem(item.id)}>✕</button>
              </div>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      {/* Footer */}
      {items.length > 0 && (
        <div className="border-t border-pos-border p-4 space-y-3 shrink-0">
          <div className="space-y-1.5 text-sm">
            <div className="flex justify-between text-pos-muted"><span>Subtotal</span><span>{fmt(subtotal)}</span></div>
            <div className="flex justify-between text-pos-muted"><span>Tax (8%)</span><span>{fmt(tax)}</span></div>
            <div className="flex justify-between font-bold text-pos-text pt-1 border-t border-pos-border"><span>Total</span><span>{fmt(grandTotal)}</span></div>
          </div>
          <button
            className="w-full py-3 rounded-xl bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 transition-opacity active:scale-[0.98]"
            onClick={() => navigate('/payment')}
          >
            Charge {fmt(grandTotal)}
          </button>
        </div>
      )}
    </div>
  );
};

export default OrderPanel;
