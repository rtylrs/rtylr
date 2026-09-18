import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { useTranslation } from 'react-i18next';

import { formatMoney } from '../../../../lib/formatMoney';

interface MenuItem {
  id: string;
  nameIndex: number;
  price: number;
  emoji: string;
}

interface CartItem extends MenuItem {
  qty: number;
}

// nameIndex points at home.features.demos.pos.items[i], in the same order.
const MENU: MenuItem[] = [
  { id: 'esp', nameIndex: 0, price: 3.5, emoji: '☕' },
  { id: 'lat', nameIndex: 1, price: 5.0, emoji: '🥛' },
  { id: 'cap', nameIndex: 2, price: 5.0, emoji: '☕' },
  { id: 'cro', nameIndex: 3, price: 3.5, emoji: '🥐' },
  { id: 'oj', nameIndex: 4, price: 4.5, emoji: '🍊' },
  { id: 'avo', nameIndex: 5, price: 12.0, emoji: '🥑' },
  { id: 'sal', nameIndex: 6, price: 14.0, emoji: '🥗' },
  { id: 'cake', nameIndex: 7, price: 7.5, emoji: '🍰' },
];

const SEQ = ['esp', 'cro', 'lat', 'esp', 'avo'];

export const POSDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const [cart, setCart] = useState<CartItem[]>([]);
  const [flash, setFlash] = useState<string | null>(null);
  const seqRef = useRef(0);

  const addItem = (id: string): void => {
    const item = MENU.find((m) => m.id === id);
    if (!item) return;
    setCart((prev) => {
      const ex = prev.find((c) => c.id === id);
      if (ex) return prev.map((c) => (c.id === id ? { ...c, qty: c.qty + 1 } : c));
      return [...prev, { ...item, qty: 1 }];
    });
    setFlash(id);
    setTimeout(() => setFlash(null), 280);
  };

  useEffect(() => {
    if (!isActive) return;
    const timer = setInterval(() => {
      addItem(SEQ[seqRef.current % SEQ.length]);
      seqRef.current += 1;
    }, 2000);
    return () => clearInterval(timer);
  }, [isActive]);

  const total = cart.reduce((s, c) => s + c.price * c.qty, 0);
  const money = (n: number): string => formatMoney(n, i18n.language);
  const itemNames = t('home.features.demos.pos.items', { returnObjects: true }) as string[];

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        display: 'flex',
        background: 'var(--bg-surface-a)',
        fontFamily: 'inherit',
        fontSize: 13,
        overflow: 'hidden',
      }}
    >
      {/* Menu grid */}
      <div
        style={{
          flex: 1,
          borderInlineEnd: '1px solid var(--hairline)',
          padding: '20px 16px',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            fontSize: 10,
            fontWeight: 600,
            letterSpacing: '0.12em',
            color: 'rgba(19,19,19,0.5)',
            textTransform: 'uppercase',
            marginBottom: 14,
          }}
        >
          {t('home.features.demos.pos.menuLabel')}
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8 }}>
          {MENU.map((item) => (
            <button
              key={item.id}
              onClick={() => addItem(item.id)}
              style={{
                background: flash === item.id ? 'var(--bg-surface-a)' : 'var(--bg-surface-b)',
                border: `1px solid ${flash === item.id ? 'var(--ink)' : 'var(--hairline)'}`,
                borderRadius: 8,
                padding: '10px 12px',
                cursor: 'pointer',
                textAlign: 'start',
                transition: 'background 0.15s ease, border-color 0.15s ease',
                color: 'var(--ink)',
              }}
            >
              <div style={{ marginBottom: 3, fontSize: 16 }}>{item.emoji}</div>
              <div style={{ fontWeight: 500, fontSize: 12, marginBottom: 2 }}>
                {itemNames[item.nameIndex]}
              </div>
              <div style={{ fontSize: 11, color: 'rgba(19,19,19,0.55)' }}>
                {money(item.price)}
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Cart */}
      <div
        style={{
          width: 185,
          padding: '20px 14px',
          display: 'flex',
          flexDirection: 'column',
          background: 'var(--bg-surface-b)',
        }}
      >
        <div
          style={{
            fontSize: 10,
            fontWeight: 600,
            letterSpacing: '0.12em',
            color: 'rgba(19,19,19,0.5)',
            textTransform: 'uppercase',
            marginBottom: 14,
          }}
        >
          {t('home.features.demos.pos.orderLabel')}
        </div>

        <div
          style={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            gap: 8,
            overflowY: 'hidden',
          }}
        >
          {cart.length === 0 ? (
            <div style={{ color: 'rgba(19,19,19,0.35)', fontSize: 12 }}>{t('home.features.demos.pos.addItems')}</div>
          ) : (
            <AnimatePresence>
              {cart.map((item) => (
                <motion.div
                  key={item.id}
                  initial={{ opacity: 0, x: 10 }}
                  animate={{ opacity: 1, x: 0 }}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    fontSize: 12,
                  }}
                >
                  <span style={{ color: 'rgba(19,19,19,0.7)' }}>
                    <span dir="ltr">{item.qty}×</span> {itemNames[item.nameIndex]}
                  </span>
                  <span style={{ color: 'var(--ink)', fontWeight: 600 }}>
                    {money(item.price * item.qty)}
                  </span>
                </motion.div>
              ))}
            </AnimatePresence>
          )}
        </div>

        <div style={{ borderTop: '1px solid var(--hairline)', paddingTop: 14, marginTop: 12 }}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: 14,
            }}
          >
            <span style={{ color: 'rgba(19,19,19,0.55)', fontSize: 12 }}>{t('home.features.demos.pos.total')}</span>
            <span style={{ color: 'var(--ink)', fontWeight: 700, fontSize: 16 }}>
              {money(total)}
            </span>
          </div>
          <div
            style={{
              background: cart.length ? 'var(--ink)' : 'var(--hairline)',
              color: cart.length ? '#ffffff' : 'rgba(19,19,19,0.35)',
              borderRadius: 8,
              padding: '11px',
              textAlign: 'center',
              fontWeight: 700,
              fontSize: 13,
              transition: 'all 0.25s ease',
              cursor: cart.length ? 'pointer' : 'default',
            }}
          >
            {cart.length
              ? t('home.features.demos.pos.chargeAmount', {
                  // Plain decimal, not formatMoney(): chargeAmount's own
                  // translated string already places the $ (before the
                  // number in English/Arabic, after it in French), so the
                  // interpolated value must not carry a second one.
                  amount: total.toLocaleString(i18n.language, {
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2,
                  }),
                })
              : t('home.features.demos.pos.charge')}
          </div>
        </div>
      </div>
    </div>
  );
};
