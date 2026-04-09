import React, { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import POSLayout from '../../layouts/POSLayout';
import { useCartStore } from '../../stores/cartStore';
import styles from './Receipt.module.scss';

const Receipt: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, clearCart } = useCartStore();
  const orderRef = useRef(`ORD-${Date.now().toString(36).toUpperCase()}`);

  const subtotal = total();
  const tax = subtotal * 0.08;
  const grandTotal = subtotal + tax;

  const formatPrice = (p: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(p);

  const formatTime = () =>
    new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });

  const handleNewOrder = () => {
    clearCart();
    navigate('/');
  };

  if (items.length === 0 && subtotal === 0) {
    // If cart was already cleared (e.g. refresh), go home
    navigate('/');
    return null;
  }

  return (
    <POSLayout>
      <div className={styles.page}>
        <motion.div
          className={styles.receipt}
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          {/* Success check */}
          <div className={styles.receipt__check}>
            <motion.div
              className={styles.receipt__checkCircle}
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ type: 'spring', damping: 14, delay: 0.15 }}
            >
              ✓
            </motion.div>
          </div>

          <h1 className={styles.receipt__title}>Order Placed</h1>
          <p className={styles.receipt__meta}>
            {orderRef.current} · {formatTime()}
          </p>

          {/* Line items */}
          <div className={styles.receipt__items}>
            {items.map((item) => (
              <div key={item.id} className={styles.receipt__row}>
                <span className={styles.receipt__qty}>{item.quantity}×</span>
                <span className={styles.receipt__name}>{item.name}</span>
                <span className={styles.receipt__price}>
                  {formatPrice(item.price * item.quantity)}
                </span>
              </div>
            ))}
          </div>

          <div className={styles.receipt__divider} />

          <div className={styles.receipt__totals}>
            <div className={styles.receipt__totalRow}>
              <span>Subtotal</span>
              <span>{formatPrice(subtotal)}</span>
            </div>
            <div className={styles.receipt__totalRow}>
              <span>Tax (8%)</span>
              <span>{formatPrice(tax)}</span>
            </div>
            <div className={`${styles.receipt__totalRow} ${styles['receipt__totalRow--grand']}`}>
              <span>Total</span>
              <span>{formatPrice(grandTotal)}</span>
            </div>
          </div>

          <div className={styles.receipt__actions}>
            <button className={styles.receipt__printBtn} onClick={() => window.print()}>
              🖨 Print Receipt
            </button>
            <button className={styles.receipt__newBtn} onClick={handleNewOrder}>
              New Order
            </button>
          </div>
        </motion.div>
      </div>
    </POSLayout>
  );
};

export default Receipt;
