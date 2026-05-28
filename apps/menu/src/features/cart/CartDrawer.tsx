import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { AnimatePresence, motion } from 'motion/react';
import { useCartStore } from '../../shared/stores/cartStore';

interface CartDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  currency?: string;
}

const CartDrawer = ({ isOpen, onClose, currency = 'USD' }: CartDrawerProps): React.JSX.Element => {
  const navigate = useNavigate();
  const { items, removeItem, updateQuantity, total } = useCartStore();

  const formatPrice = (price: number): string =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(price);

  const handleProceed = (): void => {
    onClose();
    void navigate({ to: '/checkout' });
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            style={{
              position: 'fixed',
              inset: 0,
              background: 'rgba(0,0,0,0.4)',
              zIndex: 40,
            }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />

          {/* Drawer - slides from right on desktop, slides from bottom on mobile */}
          <motion.aside
            aria-label="Cart"
            style={{
              position: 'fixed',
              zIndex: 50,
              background: 'var(--menu-bg, #ffffff)',
              display: 'flex',
              flexDirection: 'column',
              boxShadow: '-4px 0 24px rgba(0,0,0,0.12)',
            }}
            className="bottom-0 left-0 right-0 rounded-t-2xl max-h-[85vh] sm:bottom-0 sm:top-0 sm:left-auto sm:right-0 sm:w-96 sm:rounded-none sm:max-h-full"
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ type: 'spring', damping: 28, stiffness: 300 }}
          >
            {/* Header */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '16px 20px',
                borderBottom: '1px solid rgba(0,0,0,0.08)',
                flexShrink: 0,
              }}
            >
              <h2 style={{ margin: 0, fontSize: '18px', fontWeight: 700 }}>Your Order</h2>
              <button
                onClick={onClose}
                aria-label="Close cart"
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: '50%',
                  border: 'none',
                  background: 'rgba(0,0,0,0.06)',
                  cursor: 'pointer',
                  fontSize: '16px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                X
              </button>
            </div>

            {/* Body */}
            <div style={{ flex: 1, overflowY: 'auto', padding: '12px 20px' }}>
              {items.length === 0 ? (
                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: 12,
                    padding: '48px 0',
                    color: 'rgba(0,0,0,0.4)',
                  }}
                >
                  <p style={{ margin: 0, fontSize: '15px' }}>Your cart is empty</p>
                </div>
              ) : (
                <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
                  {items.map((item) => (
                    <li
                      key={item.menuItemId}
                      style={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 8,
                        padding: '12px 0',
                        borderBottom: '1px solid rgba(0,0,0,0.06)',
                      }}
                    >
                      <div
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                        }}
                      >
                        <span style={{ fontWeight: 600, fontSize: '15px' }}>{item.name}</span>
                        <span style={{ fontSize: '14px', color: 'rgba(0,0,0,0.6)' }}>
                          {formatPrice(item.unitPrice)}
                        </span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <button
                          aria-label="Decrease quantity"
                          onClick={() => updateQuantity(item.menuItemId, item.quantity - 1)}
                          style={{
                            width: 32,
                            height: 32,
                            borderRadius: '50%',
                            border: '1px solid rgba(0,0,0,0.15)',
                            background: 'transparent',
                            cursor: 'pointer',
                            fontSize: '18px',
                            lineHeight: 1,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                          }}
                        >
                          -
                        </button>
                        <span style={{ minWidth: 24, textAlign: 'center', fontWeight: 600 }}>
                          {item.quantity}
                        </span>
                        <button
                          aria-label="Increase quantity"
                          onClick={() => updateQuantity(item.menuItemId, item.quantity + 1)}
                          style={{
                            width: 32,
                            height: 32,
                            borderRadius: '50%',
                            border: 'none',
                            background: 'var(--menu-accent, #03aed2)',
                            color: 'var(--menu-accent-text, #ffffff)',
                            cursor: 'pointer',
                            fontSize: '18px',
                            lineHeight: 1,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                          }}
                        >
                          +
                        </button>
                        <button
                          aria-label={`Remove ${item.name}`}
                          onClick={() => removeItem(item.menuItemId)}
                          style={{
                            marginLeft: 'auto',
                            background: 'transparent',
                            border: 'none',
                            cursor: 'pointer',
                            color: 'rgba(0,0,0,0.35)',
                            fontSize: '13px',
                            padding: '4px 8px',
                          }}
                        >
                          Remove
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            {/* Footer */}
            {items.length > 0 && (
              <div
                style={{
                  padding: '16px 20px',
                  borderTop: '1px solid rgba(0,0,0,0.08)',
                  flexShrink: 0,
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 12,
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '15px' }}>
                  <span>Subtotal</span>
                  <span style={{ fontWeight: 700 }}>{formatPrice(total())}</span>
                </div>
                <button
                  onClick={handleProceed}
                  style={{
                    background: 'var(--menu-accent, #03aed2)',
                    color: 'var(--menu-accent-text, #ffffff)',
                    border: 'none',
                    borderRadius: 12,
                    padding: '14px 20px',
                    fontSize: '16px',
                    fontWeight: 700,
                    cursor: 'pointer',
                    width: '100%',
                    minHeight: 44,
                  }}
                >
                  Proceed to Order
                </button>
              </div>
            )}
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
};

export default CartDrawer;

// Floating cart trigger button - rendered outside the drawer, always visible
interface CartButtonProps {
  onClick: () => void;
  currency?: string;
}

export const CartButton = ({
  onClick,
  currency = 'USD',
}: CartButtonProps): React.JSX.Element | null => {
  const { itemCount, total } = useCartStore();
  const count = itemCount();

  const formatPrice = (price: number): string =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(price);

  if (count === 0) return null;

  return (
    <button
      onClick={onClick}
      aria-label={`View cart, ${count} items`}
      style={{
        position: 'fixed',
        bottom: 24,
        right: 20,
        zIndex: 30,
        background: 'var(--menu-accent, #03aed2)',
        color: 'var(--menu-accent-text, #ffffff)',
        border: 'none',
        borderRadius: 999,
        padding: '12px 20px',
        fontSize: '15px',
        fontWeight: 700,
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        boxShadow: '0 4px 16px rgba(0,0,0,0.2)',
        minHeight: 48,
      }}
    >
      <span
        style={{
          background: 'rgba(255,255,255,0.3)',
          borderRadius: '50%',
          width: 24,
          height: 24,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '13px',
          fontWeight: 800,
        }}
      >
        {count}
      </span>
      <span>View Order · {formatPrice(total())}</span>
    </button>
  );
};
