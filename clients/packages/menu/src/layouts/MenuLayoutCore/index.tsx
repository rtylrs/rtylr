import React, { useState } from "react";
import classNames from "classnames";
import MenuHeader from "../../components/MenuHeader";
import MenuCategory from "../../components/MenuCategory";
import MenuCart, { CartItem } from "../../components/MenuCart";
import AiChat from "../../components/AiChat";
import { MenuItemData } from "../../components/MenuItem";
import { MenuCategoryData } from "../../components/MenuCategory";
import {
  BusinessInfo,
  MenuLayoutConfig,
  DEFAULT_LAYOUT_CONFIG,
} from "../../types";
import { ChatMessage } from "../../components/AiChat";
import AboutPage from "../../pages/AboutPage";
import GalleryPage from "../../pages/GalleryPage";
import ReviewsPage from "../../pages/ReviewsPage";
import OffersPage from "../../pages/OffersPage";
import styles from "./MenuLayoutCore.module.scss";

export interface MenuLayoutCoreProps {
  business: BusinessInfo;
  categories: MenuCategoryData[];
  config?: Partial<MenuLayoutConfig>;
  enableCart?: boolean;
  enableAiChat?: boolean;
  onAiMessage?: (message: string, history: ChatMessage[]) => Promise<string>;
  onItemClick?: (item: MenuItemData) => void;
  onCheckout?: (items: CartItem[]) => void;
  className?: string;
}

const MenuLayoutCore: React.FC<MenuLayoutCoreProps> = ({
  business,
  categories,
  config: configOverride,
  enableCart = true,
  enableAiChat = true,
  onAiMessage,
  onItemClick,
  onCheckout,
  className,
}) => {
  const config: MenuLayoutConfig = {
    ...DEFAULT_LAYOUT_CONFIG,
    ...configOverride,
    header: { ...DEFAULT_LAYOUT_CONFIG.header, ...configOverride?.header },
    nav: { ...DEFAULT_LAYOUT_CONFIG.nav, ...configOverride?.nav },
    items: { ...DEFAULT_LAYOUT_CONFIG.items, ...configOverride?.items },
    pages: { menu: true, ...configOverride?.pages },
  };

  const [activePage, setActivePage] = useState("menu");
  const pageEnabled =
    activePage === "menu" ||
    config.pages[activePage as keyof MenuLayoutConfig["pages"]] === true;
  const renderedPage = pageEnabled ? activePage : "menu";
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [cartOpen, setCartOpen] = useState(false);
  const [activeCategoryId, setActiveCategoryId] = useState<string>(
    categories[0]?.id ?? "",
  );

  const handleAddToCart = (item: MenuItemData) => {
    setCartItems((prev) => {
      const existing = prev.find((c) => c.id === item.id);
      if (existing) {
        return prev.map((c) =>
          c.id === item.id ? { ...c, quantity: c.quantity + 1 } : c,
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
        prev.map((c) => (c.id === itemId ? { ...c, quantity: qty } : c)),
      );
    }
  };

  const scrollToCategory = (id: string) => {
    setActiveCategoryId(id);
    const el = document.getElementById(`category-${id}`);
    if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
  };

  const totalItems = cartItems.reduce((sum, c) => sum + c.quantity, 0);
  const cartEnabled = enableCart && config.items.layout !== "minimal";
  const categoryNavItems = categories.map((c) => ({ id: c.id, name: c.name }));

  return (
    <div className={classNames(styles.layout, className)}>
      <MenuHeader
        business={business}
        headerConfig={config.header}
        navConfig={config.nav}
        pages={config.pages}
        activePage={renderedPage}
        onPageChange={(page) => {
          setActivePage(page);
          window.scrollTo({ top: 0, behavior: "smooth" });
        }}
        categories={categoryNavItems}
        activeCategoryId={activeCategoryId}
        onCategorySelect={scrollToCategory}
      />

      {/* Cart bar */}
      {cartEnabled && totalItems > 0 && renderedPage === "menu" && (
        <button
          className={styles.layout__cartBar}
          onClick={() => setCartOpen(true)}
        >
          <span className={styles.layout__cartCount}>{totalItems}</span>
          View Order
        </button>
      )}

      {/* Page content */}
      <main className={styles.layout__main}>
        {renderedPage === "menu" && (
          <>
            {categories.map((cat) => (
              <MenuCategory
                key={cat.id}
                category={cat}
                displayConfig={config.items}
                onItemClick={onItemClick}
                onAddToCart={cartEnabled ? handleAddToCart : undefined}
              />
            ))}
          </>
        )}

        {renderedPage === "about" && <AboutPage business={business} />}

        {renderedPage === "gallery" && (
          <GalleryPage
            photos={business.photos ?? []}
            businessName={business.name}
          />
        )}

        {renderedPage === "reviews" && (
          <ReviewsPage
            reviews={business.reviews ?? []}
            averageRating={business.averageRating}
            reviewCount={business.reviewCount}
          />
        )}

        {renderedPage === "offers" && (
          <OffersPage offers={business.offers ?? []} />
        )}
      </main>

      {/* Cart drawer */}
      {cartEnabled && (
        <MenuCart
          items={cartItems}
          isOpen={cartOpen}
          onClose={() => setCartOpen(false)}
          onQuantityChange={handleQuantityChange}
          onCheckout={(items) => {
            setCartOpen(false);
            onCheckout?.(items);
          }}
        />
      )}

      {/* AI chat */}
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

export default MenuLayoutCore;
