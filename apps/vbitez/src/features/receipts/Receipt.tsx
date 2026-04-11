import React, { useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import POSLayout from '../../layouts/POSLayout';
import { useCartStore } from '../../shared/stores/cartStore';

const fmt = (n: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const Receipt: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, clearCart } = useCartStore();
  const orderRef = useRef(`ORD-${Date.now().toString(36).toUpperCase()}`);

  const subtotal = total();
  const tax = subtotal * 0.08;
  const grandTotal = subtotal + tax;

  if (items.length === 0 && subtotal === 0) { navigate('/'); return null; }

  return (
    <POSLayout>
      <div className="h-full flex items-center justify-center p-6">
        <motion.div
          className="w-full max-w-sm bg-pos-card border border-pos-border rounded-2xl p-6 space-y-5"
          initial={{ opacity: 0, y: 24 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.3 }}
        >
          {/* Check */}
          <div className="flex justify-center">
            <motion.div
              className="w-16 h-16 rounded-full bg-success/15 border-2 border-success/30 flex items-center justify-center text-success text-2xl font-bold"
              initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: 'spring', damping: 14, delay: 0.15 }}
            >✓</motion.div>
          </div>

          <div className="text-center">
            <h1 className="text-xl font-bold text-pos-text">Order Placed</h1>
            <p className="text-xs text-pos-muted mt-1">{orderRef.current} · {new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</p>
          </div>

          {/* Items */}
          <div className="space-y-1.5 text-sm">
            {items.map((item) => (
              <div key={item.id} className="flex gap-2">
                <span className="text-pos-muted w-6 shrink-0">{item.quantity}×</span>
                <span className="flex-1 text-pos-text">{item.name}</span>
                <span className="text-pos-muted">{fmt(item.price * item.quantity)}</span>
              </div>
            ))}
          </div>

          <div className="h-px bg-pos-border" />

          <div className="space-y-1.5 text-sm">
            <div className="flex justify-between text-pos-muted"><span>Subtotal</span><span>{fmt(subtotal)}</span></div>
            <div className="flex justify-between text-pos-muted"><span>Tax (8%)</span><span>{fmt(tax)}</span></div>
            <div className="flex justify-between font-bold text-pos-text"><span>Total</span><span>{fmt(grandTotal)}</span></div>
          </div>

          <div className="flex gap-3">
            <button className="flex-1 py-2.5 rounded-xl border border-pos-border text-pos-muted text-sm hover:border-white/20 transition-colors" onClick={() => window.print()}>🖨 Print</button>
            <button className="flex-1 py-2.5 rounded-xl bg-pos-text text-pos-bg text-sm font-semibold hover:opacity-90 transition-opacity" onClick={() => { clearCart(); navigate('/'); }}>New Order</button>
          </div>
        </motion.div>
      </div>
    </POSLayout>
  );
};

export default Receipt;
