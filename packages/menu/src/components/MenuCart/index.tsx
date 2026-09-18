import React from 'react';
import classNames from 'classnames';
import { motion, AnimatePresence } from 'motion/react';
import { MenuItemData } from '../MenuItem';
import styles from './MenuCart.module.scss';

export interface CartItem extends MenuItemData {
  quantity: number;
}

interface MenuCartProps {
  items: CartItem[];
  isOpen: boolean;
  onClose: () => void;
  onQuantityChange: (itemId: string, quantity: number) => void;
  onCheckout: (items: CartItem[]) => void;
  currency?: string;
  className?: string;
}

const MenuCart: React.FC<MenuCartProps> = ({
  items,
  isOpen,
  onClose,
  onQuantityChange,
  onCheckout,
  currency = 'USD',
  className,
}) => {
  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(price);

  const subtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            className={styles.overlay}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.aside
            role="dialog"
            aria-modal="true"
            aria-label="Your Order"
            className={classNames(styles.cart, className)}
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 28, stiffness: 300 }}
          >
            <div className={styles.cart__header}>
              <h2 className={styles.cart__title}>Your Order</h2>
              <button className={styles.cart__close} onClick={onClose} aria-label="Close cart">
                ✕
              </button>
            </div>

            <div className={styles.cart__body}>
              {items.length === 0 ? (
                <div className={styles.cart__empty}>
                  <span className={styles.cart__emptyIcon}>🛒</span>
                  <p>Your cart is empty</p>
                </div>
              ) : (
                <ul className={styles.cart__list}>
                  {items.map((item) => (
                    <li key={item.id} className={styles.cart__item}>
                      <div className={styles.cart__itemInfo}>
                        <span className={styles.cart__itemName}>{item.name}</span>
                        <span className={styles.cart__itemPrice}>{formatPrice(item.price)}</span>
                      </div>
                      <div className={styles.cart__itemControls}>
                        <button
                          className={styles.cart__qtyBtn}
                          onClick={() => onQuantityChange(item.id, item.quantity - 1)}
                          aria-label="Decrease quantity"
                        >
                          −
                        </button>
                        <span className={styles.cart__qty}>{item.quantity}</span>
                        <button
                          className={styles.cart__qtyBtn}
                          onClick={() => onQuantityChange(item.id, item.quantity + 1)}
                          aria-label="Increase quantity"
                        >
                          +
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            {items.length > 0 && (
              <div className={styles.cart__footer}>
                <div className={styles.cart__subtotal}>
                  <span>Subtotal</span>
                  <span>{formatPrice(subtotal)}</span>
                </div>
                <button className={styles.cart__checkoutBtn} onClick={() => onCheckout(items)}>
                  Place Order · {formatPrice(subtotal)}
                </button>
              </div>
            )}
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
};

export default MenuCart;
