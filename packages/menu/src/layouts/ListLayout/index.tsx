import React from 'react';
import classNames from 'classnames';
import MenuHeader, { BusinessInfo } from '../../components/MenuHeader';
import MenuCategory, { MenuCategoryData } from '../../components/MenuCategory';
import MenuCart, { CartItem } from '../../components/MenuCart';
import AiChat from '../../components/AiChat';
import { MenuItemData } from '../../components/MenuItem';
import styles from './ListLayout.module.scss';

interface ListLayoutProps {
  business: BusinessInfo;
  categories: MenuCategoryData[];
  enableCart?: boolean;
  enableAiChat?: boolean;
  onAiMessage?: (message: string, history: any[]) => Promise<string>;
  onItemClick?: (item: MenuItemData) => void;
  className?: string;
}

const ListLayout: React.FC<ListLayoutProps> = ({
  business,
  categories,
  enableCart = true,
  enableAiChat = true,
  onAiMessage,
  onItemClick,
  className,
}) => {
  const [cartItems, setCartItems] = React.useState<CartItem[]>([]);
  const [cartOpen, setCartOpen] = React.useState(false);
  const [activeCategoryId, setActiveCategoryId] = React.useState<string>(
    categories[0]?.id ?? ''
  );

  const handleAddToCart = (item: MenuItemData) => {
    setCartItems((prev) => {
      const existing = prev.find((c) => c.id === item.id);
      if (existing) {
        return prev.map((c) =>
          c.id === item.id ? { ...c, quantity: c.quantity + 1 } : c
        );
      }
      return [...prev, { ...item, quantity: 1 }];
    });
  };

  const handleQuantityChange = (itemId: string, qty: number) => {
    if (qty <= 0) {
      setCartItems((prev) => prev.filter((c) => c.id !== itemId));
    } else {
      setCartItems((prev) =>
        prev.map((c) => (c.id === itemId ? { ...c, quantity: qty } : c))
      );
    }
  };

  const handleCheckout = (items: CartItem[]) => {
    console.log('Checkout:', items);
  };

  const totalItems = cartItems.reduce((sum, c) => sum + c.quantity, 0);
  const categoryNavItems = categories.map((c) => ({ id: c.id, name: c.name }));

  const scrollToCategory = (id: string) => {
    setActiveCategoryId(id);
    const el = document.getElementById(`category-${id}`);
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  return (
    <div className={classNames(styles.layout, className)}>
      <MenuHeader
        business={business}
        categories={categoryNavItems}
        activeCategoryId={activeCategoryId}
        onCategorySelect={scrollToCategory}
      />

      {enableCart && totalItems > 0 && (
        <button className={styles.layout__cartBar} onClick={() => setCartOpen(true)}>
          <span className={styles.layout__cartCount}>{totalItems}</span>
          View Order
        </button>
      )}

      <main className={styles.layout__main}>
        {categories.map((cat) => (
          <MenuCategory
            key={cat.id}
            category={cat}
            layout="list"
            onItemClick={onItemClick}
            onAddToCart={enableCart ? handleAddToCart : undefined}
          />
        ))}
      </main>

      {enableCart && (
        <MenuCart
          items={cartItems}
          isOpen={cartOpen}
          onClose={() => setCartOpen(false)}
          onQuantityChange={handleQuantityChange}
          onCheckout={handleCheckout}
        />
      )}

      {enableAiChat && onAiMessage && (
        <AiChat
          onSendMessage={onAiMessage}
          businessName={business.name}
          buttonLabel="Ask AI"
          placeholder="Ask about dishes, allergens, specials…"
        />
      )}
    </div>
  );
};

export default ListLayout;
