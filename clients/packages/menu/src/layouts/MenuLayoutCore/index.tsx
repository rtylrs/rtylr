import React, { useMemo, useState } from "react";
import classNames from "classnames";
import { cdnImageUrl } from "@voxire/config";
import MenuHeader from "../../components/MenuHeader";
import MenuCategory, { MenuCategoryData } from "../../components/MenuCategory";
import MenuCart, { CartItem } from "../../components/MenuCart";
import AiChat, { ChatMessage } from "../../components/AiChat";
import { MenuItemData } from "../../components/MenuItem";
import {
  BusinessInfo,
  DEFAULT_FEATURES_CONFIG,
  DEFAULT_LAYOUT_CONFIG,
  MenuBannerData,
  MenuCurrencyOption,
  MenuFeaturesConfig,
  MenuLayoutConfig,
  MenuStampCardSummary,
  TableActionType,
} from "../../types";
import { buildMenuThemeStyle, formatConvertedPrice, isRtlLocale } from "../../theme";
import AboutPage from "../../pages/AboutPage";
import GalleryPage from "../../pages/GalleryPage";
import ReviewsPage from "../../pages/ReviewsPage";
import OffersPage from "../../pages/OffersPage";
import styles from "./MenuLayoutCore.module.scss";

export interface MenuLayoutCoreProps {
  business: BusinessInfo;
  categories: MenuCategoryData[];
  config?: Partial<MenuLayoutConfig>;
  banners?: MenuBannerData[];
  currencies?: MenuCurrencyOption[];
  activeCurrencyCode?: string;
  onCurrencyChange?: (code: string) => void;
  availableLocales?: string[];
  activeLocale?: string;
  onLocaleChange?: (locale: string) => void;
  stampCards?: MenuStampCardSummary[];
  tableLabel?: string;
  onTableAction?: (action: TableActionType) => Promise<void> | void;
  enableCart?: boolean;
  enableAiChat?: boolean;
  onAiMessage?: (message: string, history: ChatMessage[]) => Promise<string>;
  onItemClick?: (item: MenuItemData) => void;
  onAddItem?: (item: MenuItemData) => void;
  onCheckout?: (items: CartItem[]) => void;
  className?: string;
}

type ActiveSheet =
  | null
  | "search"
  | "preferences"
  | "filters"
  | "service"
  | "loyalty";

const MenuLayoutCore: React.FC<MenuLayoutCoreProps> = ({
  business,
  categories,
  config: configOverride,
  banners = [],
  currencies = [],
  activeCurrencyCode,
  onCurrencyChange,
  availableLocales = [],
  activeLocale,
  onLocaleChange,
  stampCards = [],
  tableLabel,
  onTableAction,
  enableCart = true,
  enableAiChat = true,
  onAiMessage,
  onItemClick,
  onAddItem,
  onCheckout,
  className,
}) => {
  const features: MenuFeaturesConfig = {
    ...DEFAULT_FEATURES_CONFIG,
    ...configOverride?.features,
    fulfillmentModes: {
      ...DEFAULT_FEATURES_CONFIG.fulfillmentModes,
      ...configOverride?.features?.fulfillmentModes,
    },
    tableActions: {
      ...DEFAULT_FEATURES_CONFIG.tableActions,
      ...configOverride?.features?.tableActions,
    },
  };

  const config: MenuLayoutConfig = {
    ...DEFAULT_LAYOUT_CONFIG,
    ...configOverride,
    header: { ...DEFAULT_LAYOUT_CONFIG.header, ...configOverride?.header },
    nav: { ...DEFAULT_LAYOUT_CONFIG.nav, ...configOverride?.nav },
    items: { ...DEFAULT_LAYOUT_CONFIG.items, ...configOverride?.items },
    pages: { menu: true, ...configOverride?.pages },
    typography: { ...DEFAULT_LAYOUT_CONFIG.typography!, ...configOverride?.typography },
    features,
  };

  const [activePage, setActivePage] = useState("menu");
  const pageEnabled =
    activePage === "menu" ||
    config.pages[activePage as keyof MenuLayoutConfig["pages"]] === true;
  const renderedPage = pageEnabled ? activePage : "menu";
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [cartOpen, setCartOpen] = useState(false);
  const [activeSheet, setActiveSheet] = useState<ActiveSheet>(null);
  const [detailItem, setDetailItem] = useState<MenuItemData | null>(null);
  const [detailQty, setDetailQty] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedDietary, setSelectedDietary] = useState<string | null>(null);
  const [internalCurrency, setInternalCurrency] = useState<string>(
    activeCurrencyCode ?? currencies.find((c) => c.isDefault)?.code ?? currencies[0]?.code ?? "",
  );
  const [tableActionNotice, setTableActionNotice] = useState<string | null>(null);
  const [activeCategoryId, setActiveCategoryId] = useState<string>(
    categories[0]?.id ?? "",
  );

  const effectiveCurrencyCode = activeCurrencyCode ?? internalCurrency;
  const activeCurrency = useMemo(
    () =>
      currencies.find((c) => c.code === effectiveCurrencyCode) ??
      currencies.find((c) => c.isDefault) ??
      currencies[0],
    [currencies, effectiveCurrencyCode],
  );

  const dietaryOptions = useMemo(() => {
    const tags = new Set<string>();
    for (const cat of categories) {
      for (const item of cat.items) {
        for (const tag of item.dietaryTags ?? []) {
          if (tag.trim()) tags.add(tag.trim());
        }
      }
    }
    return Array.from(tags);
  }, [categories]);

  const decoratedCategories = useMemo<MenuCategoryData[]>(() => {
    const q = searchQuery.trim().toLowerCase();
    return categories
      .map((cat) => {
        const items = cat.items
          .filter((item) => {
            if (
              q &&
              !item.name.toLowerCase().includes(q) &&
              !(item.description ?? "").toLowerCase().includes(q)
            ) {
              return false;
            }
            if (selectedDietary && !(item.dietaryTags ?? []).includes(selectedDietary)) {
              return false;
            }
            return true;
          })
          .map((item) => ({
            ...item,
            formattedPrice: activeCurrency
              ? formatConvertedPrice(item.price, activeCurrency, item.currency)
              : item.formattedPrice,
          }));
        return { ...cat, items };
      })
      .filter((cat) => cat.items.length > 0);
  }, [categories, searchQuery, selectedDietary, activeCurrency]);

  const handleAddToCart = (item: MenuItemData, qty = 1) => {
    onAddItem?.(item);
    setCartItems((prev) => {
      const existing = prev.find((c) => c.id === item.id);
      if (existing) {
        return prev.map((c) =>
          c.id === item.id ? { ...c, quantity: c.quantity + qty } : c,
        );
      }
      return [...prev, { ...item, quantity: qty }];
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

  const triggerTableAction = async (action: TableActionType) => {
    if (!onTableAction) return;
    await onTableAction(action);
    const labels: Record<TableActionType, string> = {
      call_waiter: "Staff notified — someone is on the way",
      request_bill: "Bill requested for your table",
      split_bill: "Split bill requested",
    };
    setTableActionNotice(labels[action]);
    setActiveSheet(null);
    setTimeout(() => setTableActionNotice(null), 3500);
  };

  const totalItems = cartItems.reduce((sum, c) => sum + c.quantity, 0);
  const totalBasePrice = cartItems.reduce((sum, c) => sum + c.price * c.quantity, 0);
  const cartEnabled = enableCart && features.enableOrdering && config.items.layout !== "minimal";
  const categoryNavItems = decoratedCategories.map((c) => ({ id: c.id, name: c.name }));
  const themeStyle = useMemo(() => buildMenuThemeStyle(config), [config]);
  const rtl = isRtlLocale(activeLocale);

  const showPrefTrigger =
    (features.enableLanguageSwitcher && availableLocales.length > 1) ||
    (features.enableCurrencySwitcher && currencies.length > 1);
  const showFilterTrigger = features.enableDietaryFilters && dietaryOptions.length > 0;
  const showTableActions =
    Boolean(onTableAction) &&
    (features.tableActions.callWaiter ||
      features.tableActions.requestBill ||
      features.tableActions.splitBill);
  const showLoyalty = features.enableLoyaltyStamps && stampCards.length > 0;
  const showBottomDock =
    renderedPage === "menu" && (showTableActions || showLoyalty || (cartEnabled && totalItems > 0));

  const topStartSlot = tableLabel ? (
    <span className={styles.glassPill}>
      <span className={styles.glassDot} />
      {tableLabel}
    </span>
  ) : null;

  const topEndSlot = (
    <>
      {features.enableSearch && renderedPage === "menu" && (
        <button
          type="button"
          className={styles.glassIconBtn}
          aria-label="Search menu"
          onClick={() => setActiveSheet("search")}
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
          </svg>
        </button>
      )}
      {showFilterTrigger && renderedPage === "menu" && (
        <button
          type="button"
          className={classNames(styles.glassIconBtn, {
            [styles["glassIconBtn--active"]]: Boolean(selectedDietary),
          })}
          aria-label="Dietary filters"
          onClick={() => setActiveSheet("filters")}
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="4" x2="20" y1="6" y2="6" />
            <line x1="7" x2="17" y1="12" y2="12" />
            <line x1="10" x2="14" y1="18" y2="18" />
          </svg>
        </button>
      )}
      {showPrefTrigger && (
        <button
          type="button"
          className={styles.glassPillBtn}
          aria-label="Language and currency"
          onClick={() => setActiveSheet("preferences")}
        >
          {(activeLocale ?? availableLocales[0] ?? "EN").toUpperCase()}
          {activeCurrency ? ` · ${activeCurrency.symbol.trim()}` : ""}
        </button>
      )}
    </>
  );

  const handleOpenItem = (item: MenuItemData) => {
    if (onItemClick) {
      onItemClick(item);
      return;
    }
    setDetailItem(item);
    setDetailQty(1);
  };

  return (
    <div
      className={classNames("rtylr-menu-root", styles.layout, className)}
      style={themeStyle}
      dir={rtl ? "rtl" : "ltr"}
    >
      {config.customCss && (
        <style data-testid="menu-custom-css">{config.customCss}</style>
      )}

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
        topStartSlot={topStartSlot}
        topEndSlot={topEndSlot}
      />

      {tableActionNotice && (
        <div className={styles.toastBar} role="status">
          {tableActionNotice}
        </div>
      )}

      {/* Active filter / search indicator strip (only when active) */}
      {(searchQuery.trim() || selectedDietary) && renderedPage === "menu" && (
        <div className={styles.activeFiltersStrip}>
          {searchQuery.trim() && (
            <button
              type="button"
              className={styles.filterTag}
              onClick={() => setSearchQuery("")}
            >
              “{searchQuery}” ×
            </button>
          )}
          {selectedDietary && (
            <button
              type="button"
              className={styles.filterTag}
              onClick={() => setSelectedDietary(null)}
            >
              {selectedDietary} ×
            </button>
          )}
        </div>
      )}

      {/* Main menu content starts immediately above the fold */}
      <main className={styles.layout__main}>
        {renderedPage === "menu" && (
          <>
            {/* Compact 1-line loyalty / promo pill (tappable, never bulky) */}
            {(showLoyalty || (features.enableBanners && banners.length > 0)) && (
              <div className={styles.compactHighlights}>
                {showLoyalty && (
                  <button
                    type="button"
                    className={styles.highlightPill}
                    onClick={() => setActiveSheet("loyalty")}
                  >
                    <span className={styles.highlightDot}>★</span>
                    <span className={styles.highlightText}>
                      <strong>{stampCards[0].title}</strong> · {stampCards[0].stampsRequired} stamps → {stampCards[0].rewardLabel}
                    </span>
                    <span className={styles.highlightChevron}>›</span>
                  </button>
                )}
                {features.enableBanners && banners.length > 0 && (
                  <div className={styles.bannerStrip}>
                    {banners.map((banner) => {
                      const img = cdnImageUrl(banner.imageUrl);
                      return (
                        <div key={banner.id} className={styles.bannerPill}>
                          {img && (
                            <img
                              src={img}
                              alt={banner.title}
                              className={styles.bannerPillImg}
                            />
                          )}
                          <div className={styles.bannerPillText}>
                            <strong>{banner.title}</strong>
                            {banner.body && <span>{banner.body}</span>}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}

            {decoratedCategories.map((cat) => (
              <MenuCategory
                key={cat.id}
                category={cat}
                displayConfig={config.items}
                onItemClick={handleOpenItem}
                onAddToCart={cartEnabled ? (item) => handleAddToCart(item, 1) : undefined}
              />
            ))}
          </>
        )}

        {renderedPage === "about" && <AboutPage business={business} />}
        {renderedPage === "gallery" && (
          <GalleryPage photos={business.photos ?? []} businessName={business.name} />
        )}
        {renderedPage === "reviews" && (
          <ReviewsPage
            reviews={business.reviews ?? []}
            averageRating={business.averageRating}
            reviewCount={business.reviewCount}
          />
        )}
        {renderedPage === "offers" && <OffersPage offers={business.offers ?? []} />}
      </main>

      <footer className={styles.layout__poweredBy}>Powered by RTYLR</footer>

      {/* Sleek floating frosted-glass bottom dock */}
      {showBottomDock && (
        <div
          className={classNames(styles.bottomDock, {
            [styles["bottomDock--compact"]]: !(cartEnabled && totalItems > 0),
          })}
        >
          {showTableActions && (
            <button
              type="button"
              className={styles.dockIconBtn}
              aria-label="Table service"
              onClick={() => setActiveSheet("service")}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
                <path d="M13.73 21a2 2 0 0 1-3.46 0" />
              </svg>
              <span>Service</span>
            </button>
          )}
          {showLoyalty && (
            <button
              type="button"
              className={styles.dockIconBtn}
              aria-label="Rewards"
              onClick={() => setActiveSheet("loyalty")}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
              </svg>
              <span>Rewards</span>
            </button>
          )}
          {cartEnabled && totalItems > 0 && (
            <button
              type="button"
              className={styles.dockCartBtn}
              onClick={() => setCartOpen(true)}
            >
              <span className={styles.dockCartBadge}>{totalItems}</span>
              <span>View Order</span>
              <span className={styles.dockCartTotal}>
                {formatConvertedPrice(totalBasePrice, activeCurrency)}
              </span>
            </button>
          )}
        </div>
      )}

      {/* Slide-up Bottom Sheets */}
      {activeSheet && (
        <div
          className={styles.sheetBackdrop}
          onClick={() => setActiveSheet(null)}
        >
          <div
            className={styles.sheetPanel}
            onClick={(e) => e.stopPropagation()}
            role="dialog"
          >
            <div className={styles.sheetHandle} />

            {activeSheet === "search" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>Search Menu</h3>
                <input
                  autoFocus
                  type="search"
                  className={styles.sheetInput}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search dishes, drinks, products..."
                  aria-label="Search menu input"
                />
                <button
                  type="button"
                  className={styles.sheetPrimaryBtn}
                  onClick={() => setActiveSheet(null)}
                >
                  Show Results
                </button>
              </div>
            )}

            {activeSheet === "preferences" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>Language &amp; Currency</h3>
                {features.enableLanguageSwitcher && availableLocales.length > 1 && (
                  <div className={styles.sheetSection}>
                    <span className={styles.sheetLabel}>Language</span>
                    <div className={styles.optionGrid}>
                      {availableLocales.map((loc) => (
                        <button
                          key={loc}
                          type="button"
                          className={classNames(styles.optionPill, {
                            [styles["optionPill--active"]]:
                              (activeLocale ?? availableLocales[0]) === loc,
                          })}
                          onClick={() => onLocaleChange?.(loc)}
                        >
                          {loc === "ar" ? "العربية" : loc === "fr" ? "Français" : loc === "en" ? "English" : loc.toUpperCase()}
                        </button>
                      ))}
                    </div>
                  </div>
                )}
                {features.enableCurrencySwitcher && currencies.length > 1 && (
                  <div className={styles.sheetSection}>
                    <span className={styles.sheetLabel}>Currency</span>
                    <div className={styles.optionGrid}>
                      {currencies.map((cur) => (
                        <button
                          key={cur.code}
                          type="button"
                          className={classNames(styles.optionPill, {
                            [styles["optionPill--active"]]: activeCurrency?.code === cur.code,
                          })}
                          onClick={() => {
                            setInternalCurrency(cur.code);
                            onCurrencyChange?.(cur.code);
                          }}
                        >
                          {cur.code} ({cur.symbol.trim()})
                        </button>
                      ))}
                    </div>
                  </div>
                )}
                <button
                  type="button"
                  className={styles.sheetPrimaryBtn}
                  onClick={() => setActiveSheet(null)}
                >
                  Done
                </button>
              </div>
            )}

            {activeSheet === "filters" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>Dietary Preferences</h3>
                <div className={styles.optionGrid}>
                  <button
                    type="button"
                    className={classNames(styles.optionPill, {
                      [styles["optionPill--active"]]: selectedDietary === null,
                    })}
                    onClick={() => setSelectedDietary(null)}
                  >
                    All Items
                  </button>
                  {dietaryOptions.map((tag) => (
                    <button
                      key={tag}
                      type="button"
                      className={classNames(styles.optionPill, {
                        [styles["optionPill--active"]]: selectedDietary === tag,
                      })}
                      onClick={() => setSelectedDietary(selectedDietary === tag ? null : tag)}
                    >
                      {tag}
                    </button>
                  ))}
                </div>
                <button
                  type="button"
                  className={styles.sheetPrimaryBtn}
                  onClick={() => setActiveSheet(null)}
                >
                  Apply Filter
                </button>
              </div>
            )}

            {activeSheet === "service" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>
                  Table Service {tableLabel ? `· ${tableLabel}` : ""}
                </h3>
                <div className={styles.serviceList}>
                  {features.tableActions.callWaiter && (
                    <button
                      type="button"
                      className={styles.serviceCard}
                      onClick={() => void triggerTableAction("call_waiter")}
                    >
                      <strong>Call Staff</strong>
                      <span>Alert a team member to your table</span>
                    </button>
                  )}
                  {features.tableActions.requestBill && (
                    <button
                      type="button"
                      className={styles.serviceCard}
                      onClick={() => void triggerTableAction("request_bill")}
                    >
                      <strong>Request Bill</strong>
                      <span>Bring the check or card terminal</span>
                    </button>
                  )}
                  {features.tableActions.splitBill && (
                    <button
                      type="button"
                      className={styles.serviceCard}
                      onClick={() => void triggerTableAction("split_bill")}
                    >
                      <strong>Split Bill</strong>
                      <span>Divide the table tab between guests</span>
                    </button>
                  )}
                </div>
              </div>
            )}

            {activeSheet === "loyalty" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>Digital Stamp Cards</h3>
                {stampCards.map((card) => (
                  <div key={card.id} className={styles.loyaltyModalCard}>
                    <strong>{card.title}</strong>
                    <p>
                      Collect {card.stampsRequired} stamps to unlock{" "}
                      <strong>{card.rewardLabel}</strong>.
                    </p>
                    <div className={styles.stampDots}>
                      {Array.from({ length: Math.min(card.stampsRequired, 10) }).map((_, i) => (
                        <span
                          key={i}
                          className={classNames(styles.stampDot, {
                            [styles["stampDot--filled"]]: i < 2,
                          })}
                        >
                          {i < 2 ? "★" : i + 1}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
                <button
                  type="button"
                  className={styles.sheetPrimaryBtn}
                  onClick={() => setActiveSheet(null)}
                >
                  Got it
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Interactive Item Detail Slide-Up Sheet */}
      {detailItem && (
        <div
          className={styles.sheetBackdrop}
          onClick={() => setDetailItem(null)}
        >
          <div
            className={styles.sheetPanel}
            onClick={(e) => e.stopPropagation()}
            role="dialog"
          >
            <button
              type="button"
              className={styles.sheetCloseBtn}
              aria-label="Close item details"
              onClick={() => setDetailItem(null)}
            >
              ×
            </button>
            {cdnImageUrl(detailItem.image) && (
              <img
                src={cdnImageUrl(detailItem.image)}
                alt={detailItem.name}
                className={styles.detailHeroImg}
              />
            )}
            <div className={styles.sheetBody}>
              <div className={styles.detailHeader}>
                <h3 className={styles.sheetTitle}>{detailItem.name}</h3>
                <span className={styles.detailPrice}>
                  {detailItem.formattedPrice ??
                    formatConvertedPrice(detailItem.price, activeCurrency, detailItem.currency)}
                </span>
              </div>
              {detailItem.description && (
                <p className={styles.detailDesc}>{detailItem.description}</p>
              )}
              {((detailItem.dietaryTags && detailItem.dietaryTags.length > 0) ||
                (detailItem.calories && detailItem.calories > 0)) && (
                <div className={styles.optionGrid}>
                  {(detailItem.dietaryTags ?? []).map((t) => (
                    <span key={t} className={styles.detailChip}>
                      {t}
                    </span>
                  ))}
                  {detailItem.calories ? (
                    <span className={styles.detailChip}>{detailItem.calories} kcal</span>
                  ) : null}
                </div>
              )}
              {cartEnabled && detailItem.available !== false && (
                <div className={styles.detailFooter}>
                  <div className={styles.qtyStepper}>
                    <button
                      type="button"
                      onClick={() => setDetailQty((q) => Math.max(1, q - 1))}
                    >
                      −
                    </button>
                    <span>{detailQty}</span>
                    <button
                      type="button"
                      onClick={() => setDetailQty((q) => q + 1)}
                    >
                      +
                    </button>
                  </div>
                  <button
                    type="button"
                    className={styles.sheetPrimaryBtn}
                    onClick={() => {
                      handleAddToCart(detailItem, detailQty);
                      setDetailItem(null);
                    }}
                  >
                    Add to Order ·{" "}
                    {formatConvertedPrice(detailItem.price * detailQty, activeCurrency, detailItem.currency)}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

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
      {enableAiChat && features.enableAiAssistant && onAiMessage && (
        <AiChat
          onSendMessage={onAiMessage}
          businessName={business.name}
          buttonLabel="Ask AI"
          placeholder="Ask about items, allergens, specials…"
        />
      )}
    </div>
  );
};

export default MenuLayoutCore;
