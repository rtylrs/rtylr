import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import POSLayout from '../../layouts/POSLayout';
import { useCartStore } from '../../stores/cartStore';
import styles from './Payment.module.scss';

type PaymentMethod = 'card' | 'cash' | 'qr';

const Payment: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, clearCart } = useCartStore();
  const [method, setMethod] = useState<PaymentMethod>('card');
  const [processing, setProcessing] = useState(false);
  const [cashAmount, setCashAmount] = useState('');

  const subtotal = total();
  const tax = subtotal * 0.08;
  const grandTotal = subtotal + tax;

  const formatPrice = (p: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(p);

  const cashChange =
    cashAmount && parseFloat(cashAmount) >= grandTotal
      ? parseFloat(cashAmount) - grandTotal
      : null;

  const handleCharge = async () => {
    setProcessing(true);
    // Simulate processing — backend will handle this
    await new Promise((r) => setTimeout(r, 1500));
    setProcessing(false);
    navigate('/receipt');
  };

  if (items.length === 0) {
    navigate('/');
    return null;
  }

  return (
    <POSLayout>
      <div className={styles.page}>
        <div className={styles.card}>
          <button className={styles.backBtn} onClick={() => navigate('/')}>
            ← Back to POS
          </button>

          <h1 className={styles.title}>Payment</h1>

          {/* Order Summary */}
          <div className={styles.summary}>
            {items.map((item) => (
              <div key={item.id} className={styles.summary__row}>
                <span className={styles.summary__name}>
                  {item.quantity}× {item.name}
                </span>
                <span className={styles.summary__price}>
                  {formatPrice(item.price * item.quantity)}
                </span>
              </div>
            ))}
            <div className={styles.summary__divider} />
            <div className={styles.summary__row}>
              <span>Subtotal</span>
              <span>{formatPrice(subtotal)}</span>
            </div>
            <div className={styles.summary__row}>
              <span>Tax (8%)</span>
              <span>{formatPrice(tax)}</span>
            </div>
            <div className={`${styles.summary__row} ${styles['summary__row--total']}`}>
              <span>Total</span>
              <span>{formatPrice(grandTotal)}</span>
            </div>
          </div>

          {/* Payment method selector */}
          <div className={styles.methods}>
            {(['card', 'cash', 'qr'] as PaymentMethod[]).map((m) => (
              <button
                key={m}
                className={`${styles.methods__btn} ${method === m ? styles['methods__btn--active'] : ''}`}
                onClick={() => setMethod(m)}
              >
                {m === 'card' ? '💳 Card' : m === 'cash' ? '💵 Cash' : '📱 QR Pay'}
              </button>
            ))}
          </div>

          {method === 'card' && (
            <div className={styles.methodContent}>
              <p className={styles.methodContent__hint}>
                Present card to terminal or tap to pay.
              </p>
            </div>
          )}

          {method === 'cash' && (
            <div className={styles.methodContent}>
              <label className={styles.methodContent__label}>Amount received</label>
              <input
                className={styles.methodContent__input}
                type="number"
                step="0.01"
                min={grandTotal}
                value={cashAmount}
                onChange={(e) => setCashAmount(e.target.value)}
                placeholder={formatPrice(grandTotal)}
              />
              {cashChange !== null && cashChange >= 0 && (
                <motion.div
                  className={styles.methodContent__change}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                >
                  Change: <strong>{formatPrice(cashChange)}</strong>
                </motion.div>
              )}
            </div>
          )}

          {method === 'qr' && (
            <div className={styles.methodContent}>
              <div className={styles.methodContent__qr}>
                <div className={styles.methodContent__qrPlaceholder}>
                  QR
                  <span>Scan to pay</span>
                </div>
              </div>
            </div>
          )}

          <button
            className={styles.chargeBtn}
            onClick={handleCharge}
            disabled={processing || (method === 'cash' && (!cashAmount || parseFloat(cashAmount) < grandTotal))}
          >
            {processing ? (
              <span className={styles.chargeBtn__loading}>Processing…</span>
            ) : (
              `Confirm Payment · ${formatPrice(grandTotal)}`
            )}
          </button>
        </div>
      </div>
    </POSLayout>
  );
};

export default Payment;
