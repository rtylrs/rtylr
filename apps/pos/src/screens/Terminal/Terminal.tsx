import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Minus, Trash2, CreditCard, Banknote, Wifi, ShoppingBag } from 'lucide-react';
import { toast } from 'sonner';
import { usePOSStore, TABLE_OPTIONS } from '../../stores/posStore';
import type { PaymentMethod } from '../../stores/posStore';
import { cn } from '../../lib/cn';
import { posApi } from '../../services/posApi';

// ── Mock menu data ─────────────────────────────────────────────

type Category = 'All' | 'Starters' | 'Mains' | 'Drinks' | 'Desserts';

interface MenuItem {
  id: string;
  name: string;
  price: number;
  category: Exclude<Category, 'All'>;
  colorHex: string;
  emoji: string;
}

const MENU_ITEMS: MenuItem[] = [
  { id: 'bruschetta',   name: 'Bruschetta',      price: 12,   category: 'Starters',  colorHex: '#c2522b', emoji: '🍞' },
  { id: 'calamari',     name: 'Calamari Rings',  price: 15,   category: 'Starters',  colorHex: '#e8a045', emoji: '🦑' },
  { id: 'caesar',       name: 'Caesar Salad',    price: 16,   category: 'Starters',  colorHex: '#4a7c59', emoji: '🥗' },
  { id: 'burger',       name: 'Wagyu Burger',    price: 24,   category: 'Mains',     colorHex: '#8b4513', emoji: '🍔' },
  { id: 'risotto',      name: 'Mushroom Risotto',price: 22,   category: 'Mains',     colorHex: '#c5a028', emoji: '🍚' },
  { id: 'salmon',       name: 'Grilled Salmon',  price: 28,   category: 'Mains',     colorHex: '#e87040', emoji: '🐟' },
  { id: 'pasta',        name: 'Truffle Pasta',   price: 20,   category: 'Mains',     colorHex: '#a0825a', emoji: '🍝' },
  { id: 'cola',         name: 'Coca-Cola',       price: 4,    category: 'Drinks',    colorHex: '#8b1a1a', emoji: '🥤' },
  { id: 'water',        name: 'Sparkling Water', price: 3.5,  category: 'Drinks',    colorHex: '#3a7aba', emoji: '💧' },
  { id: 'wine',         name: 'House Red Wine',  price: 9,    category: 'Drinks',    colorHex: '#6b1f3a', emoji: '🍷' },
  { id: 'tiramisu',     name: 'Tiramisu',        price: 11,   category: 'Desserts',  colorHex: '#7a5c3a', emoji: '🍰' },
  { id: 'lava-cake',    name: 'Lava Cake',       price: 13,   category: 'Desserts',  colorHex: '#3d2014', emoji: '🍫' },
];

const CATEGORIES: Category[] = ['All', 'Starters', 'Mains', 'Drinks', 'Desserts'];

const PAYMENT_OPTIONS: { method: PaymentMethod; label: string; icon: React.ComponentType<{ size?: number }> }[] = [
  { method: 'cash',   label: 'Cash',   icon: Banknote   },
  { method: 'card',   label: 'Card',   icon: CreditCard },
  { method: 'online', label: 'Online', icon: Wifi       },
];

// ── Menu item card ─────────────────────────────────────────────

interface MenuCardProps {
  item: MenuItem;
  qtyInCart: number;
  onAdd: () => void;
}

const MenuCard: React.FC<MenuCardProps> = ({ item, qtyInCart, onAdd }) => {
  const [hovered, setHovered] = useState(false);

  return (
    <motion.button
      whileTap={{ scale: 0.96 }}
      onClick={onAdd}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      className="relative flex flex-col rounded-xl overflow-hidden text-left transition-all"
      style={{
        background: 'var(--hub-card)',
        border: `1px solid ${hovered ? 'rgba(69,131,196,0.35)' : 'var(--hub-border)'}`,
        boxShadow: hovered ? '0 0 0 1px rgba(69,131,196,0.2)' : 'none',
      }}
    >
      {/* Color swatch / image placeholder */}
      <div
        className="relative flex items-center justify-center"
        style={{
          background: item.colorHex + '22',
          height: '80px',
          borderBottom: '1px solid var(--hub-border)',
        }}
      >
        <span style={{ fontSize: '32px', lineHeight: 1 }}>{item.emoji}</span>

        {/* +1 hint on hover */}
        <AnimatePresence>
          {hovered && (
            <motion.div
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8 }}
              transition={{ duration: 0.1 }}
              className="absolute inset-0 flex items-center justify-center rounded-t-xl"
              style={{ background: 'rgba(69,131,196,0.18)' }}
            >
              <span
                className="text-xl font-bold"
                style={{ color: '#4583C4' }}
              >
                +1
              </span>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Cart qty badge */}
        {qtyInCart > 0 && (
          <motion.span
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            className="absolute top-2 right-2 h-5 w-5 rounded-full text-[10px] font-bold flex items-center justify-center"
            style={{ background: '#4583C4', color: '#fff' }}
          >
            {qtyInCart}
          </motion.span>
        )}
      </div>

      {/* Info */}
      <div className="px-3 py-2.5">
        <p className="text-xs font-semibold leading-tight mb-0.5 line-clamp-1" style={{ color: 'var(--hub-text)' }}>
          {item.name}
        </p>
        <p className="text-xs font-bold" style={{ color: '#4583C4' }}>
          ${item.price.toFixed(2)}
        </p>
      </div>
    </motion.button>
  );
};

// ── Cart item row ──────────────────────────────────────────────

interface CartRowProps {
  id: string;
  name: string;
  price: number;
  qty: number;
  onIncrement: () => void;
  onDecrement: () => void;
  onRemove: () => void;
}

const CartRow: React.FC<CartRowProps> = ({ name, price, qty, onIncrement, onDecrement, onRemove }) => (
  <motion.div
    layout
    initial={{ opacity: 0, x: 20 }}
    animate={{ opacity: 1, x: 0 }}
    exit={{ opacity: 0, x: 20 }}
    transition={{ duration: 0.15 }}
    className="flex items-center gap-2 py-2.5"
    style={{ borderBottom: '1px solid var(--hub-border)' }}
  >
    <div className="flex-1 min-w-0">
      <p className="text-xs font-semibold truncate" style={{ color: 'var(--hub-text)' }}>{name}</p>
      <p className="text-[11px]" style={{ color: 'var(--hub-text-muted)' }}>${price.toFixed(2)} each</p>
    </div>

    {/* Qty controls */}
    <div className="flex items-center gap-1 shrink-0">
      <button
        onClick={onDecrement}
        className="h-6 w-6 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--hub-hover)]"
        style={{ border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
      >
        <Minus size={10} />
      </button>
      <span
        className="text-xs font-bold w-5 text-center tabular-nums"
        style={{ color: 'var(--hub-text)' }}
      >
        {qty}
      </span>
      <button
        onClick={onIncrement}
        className="h-6 w-6 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--hub-hover)]"
        style={{ border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
      >
        <Plus size={10} />
      </button>
    </div>

    {/* Item total */}
    <span
      className="text-xs font-bold tabular-nums w-14 text-right shrink-0"
      style={{ color: 'var(--hub-text)' }}
    >
      ${(price * qty).toFixed(2)}
    </span>

    <button
      onClick={onRemove}
      className="h-6 w-6 rounded-lg flex items-center justify-center transition-colors hover:bg-[rgba(239,68,68,0.1)] shrink-0"
      style={{ color: 'rgba(239,68,68,0.6)' }}
    >
      <Trash2 size={11} />
    </button>
  </motion.div>
);

// ── Terminal screen ────────────────────────────────────────────

const Terminal: React.FC = () => {
  const [activeCategory, setActiveCategory] = useState<Category>('All');

  const {
    activeTableId,
    cartItems,
    addItem,
    removeItem,
    incrementQty,
    decrementQty,
    clearCart,
    paymentMethod,
    setPaymentMethod,
    orderNotes,
    setOrderNotes,
    chargeOrder,
  } = usePOSStore();

  const activeTable = TABLE_OPTIONS.find((t) => t.id === activeTableId);

  const filtered =
    activeCategory === 'All'
      ? MENU_ITEMS
      : MENU_ITEMS.filter((m) => m.category === activeCategory);

  const subtotal = cartItems.reduce((s, i) => s + i.price * i.qty, 0);
  const tax = subtotal * 0.1;
  const total = subtotal + tax;

  const handleCharge = () => {
    if (cartItems.length === 0) {
      toast.error('No items in order');
      return;
    }
    chargeOrder();
    toast.success(`Order charged — $${total.toFixed(2)} via ${paymentMethod}`);
  };

  return (
    <div className="flex h-full overflow-hidden">
      {/* ── Left: menu area ──────────────────────────────────── */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* Category tabs */}
        <div
          className="flex items-center gap-1 px-4 py-2.5 shrink-0"
          style={{ borderBottom: '1px solid var(--hub-border)', background: 'var(--hub-surface)' }}
        >
          {CATEGORIES.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
              style={{
                background: activeCategory === cat ? '#4583C4' : 'var(--hub-card)',
                color: activeCategory === cat ? '#fff' : 'var(--hub-text-muted)',
                border: `1px solid ${activeCategory === cat ? '#4583C4' : 'var(--hub-border)'}`,
              }}
            >
              {cat}
            </button>
          ))}
        </div>

        {/* Menu grid */}
        <div className="flex-1 overflow-y-auto p-4">
          <motion.div
            layout
            className="grid grid-cols-3 gap-3"
            style={{ gridAutoRows: 'min-content' }}
          >
            <AnimatePresence mode="popLayout">
              {filtered.map((item) => {
                const cartEntry = cartItems.find((c) => c.id === item.id);
                return (
                  <motion.div
                    key={item.id}
                    layout
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.15 }}
                  >
                    <MenuCard
                      item={item}
                      qtyInCart={cartEntry?.qty ?? 0}
                      onAdd={() =>
                        addItem({
                          id: item.id,
                          name: item.name,
                          price: item.price,
                          category: item.category,
                        })
                      }
                    />
                  </motion.div>
                );
              })}
            </AnimatePresence>
          </motion.div>
        </div>
      </div>

      {/* ── Right: order / cart ──────────────────────────────── */}
      <div
        className="w-96 shrink-0 flex flex-col overflow-hidden"
        style={{
          background: '#0e0e0e',
          borderLeft: '1px solid var(--hub-border)',
        }}
      >
        {/* Cart header */}
        <div
          className="px-4 py-3 shrink-0"
          style={{ borderBottom: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
                Current Order
              </h2>
              <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
                {activeTable?.label ?? 'Takeaway'} · {cartItems.length} item{cartItems.length !== 1 ? 's' : ''}
              </p>
            </div>
            <ShoppingBag size={16} style={{ color: 'var(--hub-text-dim)' }} />
          </div>
        </div>

        {/* Cart items */}
        <div className="flex-1 overflow-y-auto px-4">
          {cartItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full gap-2 py-8">
              <ShoppingBag size={28} style={{ color: 'var(--hub-text-dim)' }} />
              <p className="text-xs text-center" style={{ color: 'var(--hub-text-dim)' }}>
                No items yet.
                <br />
                Tap a menu item to add.
              </p>
            </div>
          ) : (
            <AnimatePresence>
              {cartItems.map((item) => (
                <CartRow
                  key={item.id}
                  id={item.id}
                  name={item.name}
                  price={item.price}
                  qty={item.qty}
                  onIncrement={() => incrementQty(item.id)}
                  onDecrement={() => decrementQty(item.id)}
                  onRemove={() => removeItem(item.id)}
                />
              ))}
            </AnimatePresence>
          )}
        </div>

        {/* Totals + payment */}
        <div
          className="px-4 pt-3 pb-4 shrink-0 space-y-3"
          style={{ borderTop: '1px solid var(--hub-border)' }}
        >
          {/* Subtotal / tax / total */}
          <div className="space-y-1.5">
            <div className="flex justify-between text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              <span>Subtotal</span>
              <span className="tabular-nums">${subtotal.toFixed(2)}</span>
            </div>
            <div className="flex justify-between text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              <span>Tax (10%)</span>
              <span className="tabular-nums">${tax.toFixed(2)}</span>
            </div>
            <div
              className="flex justify-between text-sm font-bold pt-1.5"
              style={{ borderTop: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
            >
              <span>Total</span>
              <span className="tabular-nums">${total.toFixed(2)}</span>
            </div>
          </div>

          {/* Payment method */}
          <div className="flex gap-1.5">
            {PAYMENT_OPTIONS.map(({ method, label, icon: Icon }) => (
              <button
                key={method}
                onClick={() => setPaymentMethod(method)}
                className={cn(
                  'flex-1 flex flex-col items-center gap-1 py-2 rounded-xl text-[10px] font-semibold transition-all'
                )}
                style={{
                  background: paymentMethod === method ? 'rgba(69,131,196,0.15)' : 'var(--hub-card)',
                  color: paymentMethod === method ? '#4583C4' : 'var(--hub-text-muted)',
                  border: `1px solid ${paymentMethod === method ? 'rgba(69,131,196,0.4)' : 'var(--hub-border)'}`,
                }}
              >
                <Icon size={14} />
                {label}
              </button>
            ))}
          </div>

          {/* Order notes */}
          <textarea
            value={orderNotes}
            onChange={(e) => setOrderNotes(e.target.value)}
            placeholder="Order notes…"
            rows={2}
            className="w-full resize-none rounded-xl px-3 py-2 text-xs outline-none transition-colors placeholder:text-[var(--hub-text-dim)]"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
            onFocus={(e) => {
              e.currentTarget.style.borderColor = 'var(--accent-border)';
            }}
            onBlur={(e) => {
              e.currentTarget.style.borderColor = 'var(--hub-border)';
            }}
          />

          {/* Charge button */}
          <motion.button
            whileTap={{ scale: 0.98 }}
            onClick={handleCharge}
            className="w-full py-3 rounded-xl text-sm font-bold transition-opacity"
            style={{
              background: cartItems.length === 0 ? 'var(--hub-card)' : '#4583C4',
              color: cartItems.length === 0 ? 'var(--hub-text-dim)' : '#fff',
              cursor: cartItems.length === 0 ? 'not-allowed' : 'pointer',
            }}
          >
            Charge ${total.toFixed(2)}
          </motion.button>

          {/* Clear order */}
          {cartItems.length > 0 && (
            <button
              onClick={clearCart}
              className="w-full py-2 rounded-xl text-xs font-semibold transition-colors hover:bg-[var(--hub-hover)]"
              style={{
                background: 'transparent',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text-muted)',
              }}
            >
              Clear Order
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default Terminal;
