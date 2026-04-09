import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useCartStore } from '../../../../stores/cartStore';
import styles from './OrderPanel.module.scss';

interface OrderPanelProps {
  tableNumber?: string;
}

const OrderPanel: React.FC<OrderPanelProps> = ({ tableNumber }) => {
  const navigate = useNavigate();
  const { items, updateQuantity, removeItem, clearCart, total } = useCartStore();

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);

  const subtotal = total();
  const tax = subtotal * 0.08;
  const grandTotal = subtotal + tax;

  return (
    <div className={styles.panel}>
      <div className={styles.panel__header}>
        <div>
          <h2 className={styles.panel__title}>Current Order</h2>
          {tableNumber && (
            <span className={styles.panel__table}>Table {tableNumber}</span>
          )}
        </div>
        {items.length > 0 && (
          <button className={styles.panel__clearBtn} onClick={clearCart}>
            Clear
          </button>
        )}
      </div>

      <div className={styles.panel__items}>
        <AnimatePresence>
          {items.length === 0 ? (
            <motion.div
              className={styles.panel__empty}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
            >
              <span className={styles.panel__emptyIcon}>🧾</span>
              <p>No items added yet</p>
              <span>Tap a product to add it</span>
            </motion.div>
          ) : (
            items.map((item) => (
              <motion.div
                key={item.id}
                className={styles.panel__item}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                layout
              >
                <div className={styles.panel__itemInfo}>
                  <span className={styles.panel__itemName}>{item.name}</span>
                  <span className={styles.panel__itemPrice}>
                    {formatPrice(item.price * item.quantity)}
                  </span>
                </div>
                <div className={styles.panel__itemControls}>
                  <button
                    className={styles.panel__qtyBtn}
                    onClick={() => updateQuantity(item.id, item.quantity - 1)}
                    aria-label="Decrease"
                  >
                    −
                  </button>
                  <span className={styles.panel__qty}>{item.quantity}</span>
                  <button
                    className={styles.panel__qtyBtn}
                    onClick={() => updateQuantity(item.id, item.quantity + 1)}
                    aria-label="Increase"
                  >
                    +
                  </button>
                  <button
                    className={styles.panel__removeBtn}
                    onClick={() => removeItem(item.id)}
                    aria-label="Remove item"
                  >
                    ✕
                  </button>
                </div>
              </motion.div>
            ))
          )}
        </AnimatePresence>
      </div>

      {items.length > 0 && (
        <div className={styles.panel__footer}>
          <div className={styles.panel__totals}>
            <div className={styles.panel__totalRow}>
              <span>Subtotal</span>
              <span>{formatPrice(subtotal)}</span>
            </div>
            <div className={styles.panel__totalRow}>
              <span>Tax (8%)</span>
              <span>{formatPrice(tax)}</span>
            </div>
            <div className={`${styles.panel__totalRow} ${styles['panel__totalRow--grand']}`}>
              <span>Total</span>
              <span>{formatPrice(grandTotal)}</span>
            </div>
          </div>

          <button
            className={styles.panel__chargeBtn}
            onClick={() => navigate('/payment')}
          >
            Charge {formatPrice(grandTotal)}
          </button>
        </div>
      )}
    </div>
  );
};

export default OrderPanel;
