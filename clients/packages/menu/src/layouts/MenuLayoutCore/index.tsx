import React, { useEffect, useMemo, useRef, useState } from "react";
import classNames from "classnames";
import { cdnImageUrl } from "@voxire/config";
import MenuHeader from "../../components/MenuHeader";
import MenuCategory, { MenuCategoryData } from "../../components/MenuCategory";
import MenuCart, { CartItem } from "../../components/MenuCart";
import AiChat, { ChatMessage } from "../../components/AiChat";
import { MenuItemData } from "../../components/MenuItem";
import {
  BusinessInfo,
  CrmGuestInput,
  CustomerStampProgressData,
  DEFAULT_FEATURES_CONFIG,
  DEFAULT_LAYOUT_CONFIG,
  MenuBannerData,
  MenuCurrencyOption,
  MenuFeaturesConfig,
  MenuLayoutConfig,
  MenuLayoutLabels,
  MenuStampCardSummary,
  TableActionType,
} from "../../types";
import {
  buildMenuThemeStyle,
  formatConvertedPrice,
  isRtlLocale,
  loadMenuFont,
} from "../../theme";
import AboutPage from "../../pages/AboutPage";
import GalleryPage from "../../pages/GalleryPage";
import ReviewsPage from "../../pages/ReviewsPage";
import OffersPage from "../../pages/OffersPage";
import styles from "./MenuLayoutCore.module.scss";

const DEFAULT_LABELS: MenuLayoutLabels = {
  searchMenu: "Search Menu",
  searchPlaceholder: "Search dishes, drinks, products...",
  showResults: "Show Results",
  dietaryFilters: "Dietary filters",
  dietaryPreferences: "Dietary Preferences",
  allItems: "All Items",
  applyFilter: "Apply Filter",
  languageAndCurrency: "Language & Currency",
  language: "Language",
  currency: "Currency",
  done: "Done",
  tableService: "Table service",
  service: "Service",
  callStaff: "Call Staff",
  callStaffSubtitle: "Alert a team member to your table",
  requestBill: "Request Bill",
  requestBillSubtitle: "Bring the check or card terminal",
  splitBill: "Split Bill",
  splitBillSubtitle: "Divide the table tab between guests",
  callWaiterSuccess: "Staff notified — someone is on the way",
  requestBillSuccess: "Bill requested for your table",
  splitBillSuccess: "Split bill requested",
  tableActionError: "Could not send request — please try again",
  rewards: "Rewards",
  digitalStampCards: "Digital Stamp Cards",
  stampsPhonePlaceholder: "Enter phone number to check stamps",
  checkStamps: "Check",
  checkingStamps: "Checking…",
  gotIt: "Got it",
  viewOrder: "View Order",
  addToOrder: "Add to Order",
  yourOrder: "Your Order",
  closeCart: "Close cart",
  emptyCart: "Your cart is empty",
  subtotal: "Subtotal",
  placeOrder: "Place Order",
  previewCheckoutNotice: "Preview only — live checkout runs on your published menu link",
  askAi: "Ask AI",
  aiPlaceholder: "Ask about items, allergens, specials…",
  joinClub: "Stay in touch",
  crmGuestTitle: "Join our guest list",
  crmGuestSubtitle: "Leave your details for offers, events, and personalized rewards.",
  crmGuestName: "Full name",
  crmGuestPhone: "Phone number",
  crmGuestEmail: "Email address",
  crmGuestConsent: "Save my contact info so this business can send me updates & offers",
  crmGuestSubmit: "Save my details",
  crmGuestSubmitting: "Saving…",
  crmGuestSuccess: "Thanks! Your details have been saved.",
  crmGuestError: "Could not save your details — please try again.",
  crmGuestContactRequired: "Enter a phone number or email address to continue.",
};

function safeExternalHref(url?: string): string | null {
  if (!url) return null;
  const trimmed = url.trim();
  if (/^https?:\/\//i.test(trimmed)) return trimmed;
  return null;
}

const SCROLL_OFFSET_PADDING_PX = 12;

function findScrollContainer(element: HTMLElement | null): HTMLElement | null {
  let node = element?.parentElement ?? null;
  while (node && node !== document.body && node !== document.documentElement) {
    const overflowY = window.getComputedStyle(node).overflowY;
    const isScrollable =
      overflowY === "auto" || overflowY === "scroll" || overflowY === "overlay";
    if (isScrollable && node.scrollHeight > node.clientHeight) {
      return node;
    }
    node = node.parentElement;
  }
  return null;
}

export interface MenuLayoutCoreProps {
  business: BusinessInfo;
  categories: MenuCategoryData[];
  config?: Partial<MenuLayoutConfig>;
  banners?: MenuBannerData[];
  currencies?: MenuCurrencyOption[];
  activeCurrencyCode?: string;
  onCurrencyChange?: (code: string) => void;
  onCurrencyOptionChange?: (currency: MenuCurrencyOption) => void;
  availableLocales?: string[];
  activeLocale?: string;
  onLocaleChange?: (locale: string) => void;
  stampCards?: MenuStampCardSummary[];
  customerStamps?: CustomerStampProgressData[];
  onLookupStamps?: (phone: string) => Promise<void> | void;
  tableLabel?: string;
  onTableAction?: (action: TableActionType) => Promise<void> | void;
  onCaptureCrmGuest?: (input: CrmGuestInput) => Promise<void> | void;
  previewMode?: boolean;
  labels?: Partial<MenuLayoutLabels>;
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
  | "loyalty"
  | "crm";

const MenuLayoutCore: React.FC<MenuLayoutCoreProps> = ({
  business,
  categories,
  config: configOverride,
  banners = [],
  currencies = [],
  activeCurrencyCode,
  onCurrencyChange,
  onCurrencyOptionChange,
  availableLocales = [],
  activeLocale,
  onLocaleChange,
  stampCards = [],
  customerStamps = [],
  onLookupStamps,
  tableLabel,
  onTableAction,
  onCaptureCrmGuest,
  previewMode = false,
  labels: labelsOverride,
  enableCart = true,
  enableAiChat = true,
  onAiMessage,
  onItemClick,
  onAddItem,
  onCheckout,
  className,
}) => {
  const t = useMemo(
    () => ({ ...DEFAULT_LABELS, ...labelsOverride }),
    [labelsOverride],
  );
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
  const [stampLookupPhone, setStampLookupPhone] = useState("");
  const [lookingUpStamps, setLookingUpStamps] = useState(false);
  const [crmName, setCrmName] = useState("");
  const [crmPhone, setCrmPhone] = useState("");
  const [crmEmail, setCrmEmail] = useState("");
  const [crmConsent, setCrmConsent] = useState(true);
  const [crmSubmitting, setCrmSubmitting] = useState(false);
  const [crmError, setCrmError] = useState<string | null>(null);
  const rootRef = useRef<HTMLDivElement>(null);
  const [stickyOffset, setStickyOffset] = useState(SCROLL_OFFSET_PADDING_PX);
  const [activeCategoryId, setActiveCategoryId] = useState<string>(
    categories[0]?.id ?? "",
  );

  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;
    const header = root.querySelector("header");
    if (!header || !config.header.sticky) {
      setStickyOffset(SCROLL_OFFSET_PADDING_PX);
      return;
    }
    const updateOffset = () => {
      const height = header.getBoundingClientRect().height;
      setStickyOffset(Math.ceil(height) + SCROLL_OFFSET_PADDING_PX);
    };
    updateOffset();
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(updateOffset);
    observer.observe(header);
    return () => observer.disconnect();
  }, [config.header.sticky, config.header.coverHeight, renderedPage, categories.length]);

  useEffect(() => {
    loadMenuFont(config.typography?.headingFont);
    loadMenuFont(config.typography?.bodyFont);
  }, [config.typography?.headingFont, config.typography?.bodyFont]);

  useEffect(() => {
    if (!activeSheet && !detailItem) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setActiveSheet(null);
        setDetailItem(null);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [activeSheet, detailItem]);

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
    const root = rootRef.current;
    const el =
      (root?.querySelector(`[id="category-${id.replace(/"/g, '\\"')}"]`) as HTMLElement | null) ??
      document.getElementById(`category-${id}`);
    if (!el) return;

    const container = findScrollContainer(el);
    const isFirstCategory = decoratedCategories[0]?.id === id;
    if (isFirstCategory) {
      if (!container) {
        window.scrollTo({ top: 0, behavior: "smooth" });
      } else if (typeof container.scrollTo === "function") {
        container.scrollTo({ top: 0, behavior: "smooth" });
      } else {
        container.scrollTop = 0;
      }
      return;
    }

    const header = root?.querySelector("header") ?? null;
    const headerHeight =
      config.header.sticky && header ? header.getBoundingClientRect().height : 0;
    const offset = Math.ceil(headerHeight) + SCROLL_OFFSET_PADDING_PX;

    if (!container) {
      const targetTop = Math.max(0, window.scrollY + el.getBoundingClientRect().top - offset);
      window.scrollTo({ top: targetTop, behavior: "smooth" });
      return;
    }

    const containerTop = container.getBoundingClientRect().top;
    const targetTop = Math.max(
      0,
      container.scrollTop + (el.getBoundingClientRect().top - containerTop) - offset,
    );
    if (typeof container.scrollTo === "function") {
      container.scrollTo({ top: targetTop, behavior: "smooth" });
    } else {
      container.scrollTop = targetTop;
    }
  };

  const triggerTableAction = async (action: TableActionType) => {
    if (!onTableAction) return;
    try {
      await onTableAction(action);
      const noticeByAction: Record<TableActionType, string> = {
        call_waiter: t.callWaiterSuccess,
        request_bill: t.requestBillSuccess,
        split_bill: t.splitBillSuccess,
      };
      setTableActionNotice(noticeByAction[action]);
      setActiveSheet(null);
    } catch {
      setTableActionNotice(t.tableActionError);
    }
    setTimeout(() => setTableActionNotice(null), 3500);
  };

  const handleLookupStamps = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!onLookupStamps || !stampLookupPhone.trim()) return;
    setLookingUpStamps(true);
    try {
      await onLookupStamps(stampLookupPhone.trim());
    } finally {
      setLookingUpStamps(false);
    }
  };

  const handleSubmitCrmGuest = async (event: React.FormEvent) => {
    event.preventDefault();
    setCrmError(null);
    const trimmedPhone = crmPhone.trim();
    const trimmedEmail = crmEmail.trim();
    if (!trimmedPhone && !trimmedEmail) {
      setCrmError(t.crmGuestContactRequired);
      return;
    }
    if (!crmConsent || !onCaptureCrmGuest) return;
    setCrmSubmitting(true);
    try {
      await onCaptureCrmGuest({
        name: crmName.trim() || undefined,
        phone: trimmedPhone || undefined,
        email: trimmedEmail || undefined,
        consent: true,
      });
      setCrmName("");
      setCrmPhone("");
      setCrmEmail("");
      setActiveSheet(null);
      setTableActionNotice(t.crmGuestSuccess);
      setTimeout(() => setTableActionNotice(null), 3500);
    } catch {
      setCrmError(t.crmGuestError);
    } finally {
      setCrmSubmitting(false);
    }
  };

  const totalItems = cartItems.reduce((sum, c) => sum + c.quantity, 0);
  const totalBasePrice = cartItems.reduce((sum, c) => sum + c.price * c.quantity, 0);
  const cartEnabled = enableCart && features.enableOrdering && config.items.layout !== "minimal";
  const categoryNavItems = decoratedCategories.map((c) => ({ id: c.id, name: c.name }));
  const themeStyle = useMemo(
    () =>
      ({
        ...buildMenuThemeStyle(config),
        "--menu-scroll-margin-top": `${stickyOffset}px`,
      }) as React.CSSProperties,
    [config, stickyOffset],
  );
  const rtl = isRtlLocale(activeLocale);

  const showPrefTrigger =
    (features.enableLanguageSwitcher && availableLocales.length > 1) ||
    (features.enableCurrencySwitcher && currencies.length > 1);
  const showFilterTrigger = features.enableDietaryFilters && dietaryOptions.length > 0;
  const showTableActions =
    features.fulfillmentModes.dineIn &&
    Boolean(onTableAction) &&
    (features.tableActions.callWaiter ||
      features.tableActions.requestBill ||
      features.tableActions.splitBill);
  const showLoyalty = features.enableLoyaltyStamps && stampCards.length > 0;
  const showCrmCapture = features.enableCrmCapture && Boolean(onCaptureCrmGuest);
  const showBottomDock =
    renderedPage === "menu" &&
    (showTableActions || showLoyalty || showCrmCapture || (cartEnabled && totalItems > 0));

  const topStartSlot =
    tableLabel && features.fulfillmentModes.dineIn ? (
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
      ref={rootRef}
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
          const container = findScrollContainer(rootRef.current);
          if (!container) {
            window.scrollTo({ top: 0, behavior: "smooth" });
          } else if (typeof container.scrollTo === "function") {
            container.scrollTo({ top: 0, behavior: "smooth" });
          } else {
            container.scrollTop = 0;
          }
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
      {React.createElement(
        previewMode ? "div" : "main",
        { className: styles.layout__main },
        <>
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
                        const href = safeExternalHref(banner.linkUrl);
                        const inner = (
                          <>
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
                          </>
                        );
                        return href ? (
                          <a
                            key={banner.id}
                            href={href}
                            target="_blank"
                            rel="noopener noreferrer"
                            className={styles.bannerPill}
                            style={{ textDecoration: "none", color: "inherit" }}
                          >
                            {inner}
                          </a>
                        ) : (
                          <div key={banner.id} className={styles.bannerPill}>
                            {inner}
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
        </>,
      )}

      <footer className={styles.layout__poweredBy} data-rtylr-powered-by="true">
        Powered by RTYLR
      </footer>

      {/* Sleek floating bottom dock */}
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
              aria-label={t.tableService}
              onClick={() => setActiveSheet("service")}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
                <path d="M13.73 21a2 2 0 0 1-3.46 0" />
              </svg>
              <span>{t.service}</span>
            </button>
          )}
          {showLoyalty && (
            <button
              type="button"
              className={styles.dockIconBtn}
              aria-label={t.rewards}
              onClick={() => setActiveSheet("loyalty")}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
              </svg>
              <span>{t.rewards}</span>
            </button>
          )}
          {showCrmCapture && (
            <button
              type="button"
              className={styles.dockIconBtn}
              aria-label={t.joinClub}
              onClick={() => {
                setCrmError(null);
                setActiveSheet("crm");
              }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                <circle cx="9" cy="7" r="4" />
                <line x1="19" x2="19" y1="8" y2="14" />
                <line x1="22" x2="16" y1="11" y2="11" />
              </svg>
              <span>{t.joinClub}</span>
            </button>
          )}
          {cartEnabled && totalItems > 0 && (
            <button
              type="button"
              className={styles.dockCartBtn}
              onClick={() => setCartOpen(true)}
            >
              <span className={styles.dockCartBadge}>{totalItems}</span>
              <span>{t.viewOrder}</span>
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
            aria-modal="true"
            aria-label={
              activeSheet === "search"
                ? t.searchMenu
                : activeSheet === "preferences"
                  ? t.languageAndCurrency
                  : activeSheet === "filters"
                    ? t.dietaryPreferences
                    : activeSheet === "service"
                      ? t.tableService
                      : activeSheet === "crm"
                        ? t.crmGuestTitle
                        : t.digitalStampCards
            }
          >
            <div className={styles.sheetHandle} />

            {activeSheet === "search" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>{t.searchMenu}</h3>
                <input
                  autoFocus
                  type="search"
                  className={styles.sheetInput}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder={t.searchPlaceholder}
                  aria-label={t.searchMenu}
                />
                <button
                  type="button"
                  className={styles.sheetPrimaryBtn}
                  onClick={() => setActiveSheet(null)}
                >
                  {t.showResults}
                </button>
              </div>
            )}

            {activeSheet === "preferences" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>{t.languageAndCurrency}</h3>
                {features.enableLanguageSwitcher && availableLocales.length > 1 && (
                  <div className={styles.sheetSection}>
                    <span className={styles.sheetLabel}>{t.language}</span>
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
                    <span className={styles.sheetLabel}>{t.currency}</span>
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
                            onCurrencyOptionChange?.(cur);
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
                  {t.done}
                </button>
              </div>
            )}

            {activeSheet === "filters" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>{t.dietaryPreferences}</h3>
                <div className={styles.optionGrid}>
                  <button
                    type="button"
                    className={classNames(styles.optionPill, {
                      [styles["optionPill--active"]]: selectedDietary === null,
                    })}
                    onClick={() => setSelectedDietary(null)}
                  >
                    {t.allItems}
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
                  {t.applyFilter}
                </button>
              </div>
            )}

            {activeSheet === "service" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>
                  {t.tableService} {tableLabel ? `· ${tableLabel}` : ""}
                </h3>
                <div className={styles.serviceList}>
                  {features.tableActions.callWaiter && (
                    <button
                      type="button"
                      className={styles.serviceCard}
                      onClick={() => void triggerTableAction("call_waiter")}
                    >
                      <strong>{t.callStaff}</strong>
                      <span>{t.callStaffSubtitle}</span>
                    </button>
                  )}
                  {features.tableActions.requestBill && (
                    <button
                      type="button"
                      className={styles.serviceCard}
                      onClick={() => void triggerTableAction("request_bill")}
                    >
                      <strong>{t.requestBill}</strong>
                      <span>{t.requestBillSubtitle}</span>
                    </button>
                  )}
                  {features.tableActions.splitBill && (
                    <button
                      type="button"
                      className={styles.serviceCard}
                      onClick={() => void triggerTableAction("split_bill")}
                    >
                      <strong>{t.splitBill}</strong>
                      <span>{t.splitBillSubtitle}</span>
                    </button>
                  )}
                </div>
              </div>
            )}

            {activeSheet === "loyalty" && (
              <div className={styles.sheetBody}>
                <h3 className={styles.sheetTitle}>{t.digitalStampCards}</h3>
                {onLookupStamps && (
                  <form
                    onSubmit={(e) => void handleLookupStamps(e)}
                    style={{ display: "flex", gap: 8, marginBottom: 12 }}
                  >
                    <input
                      type="tel"
                      className={styles.sheetInput}
                      value={stampLookupPhone}
                      onChange={(e) => setStampLookupPhone(e.target.value)}
                      placeholder={t.stampsPhonePlaceholder}
                      aria-label={t.stampsPhonePlaceholder}
                      style={{ flex: 1, margin: 0 }}
                    />
                    <button
                      type="submit"
                      className={styles.optionPill}
                      disabled={lookingUpStamps || !stampLookupPhone.trim()}
                    >
                      {lookingUpStamps ? t.checkingStamps : t.checkStamps}
                    </button>
                  </form>
                )}
                {stampCards.map((card) => {
                  const progress = customerStamps.find(
                    (s) => String(s.stampCardId) === String(card.id),
                  );
                  const earned = progress
                    ? progress.stampsCount % Math.max(card.stampsRequired, 1)
                    : 0;
                  const unlocked = progress?.rewardsUnlocked ?? 0;
                  return (
                    <div key={card.id} className={styles.loyaltyModalCard}>
                      <strong>{card.title}</strong>
                      <p>
                        Collect {card.stampsRequired} stamps to unlock{" "}
                        <strong>{card.rewardLabel}</strong>.
                        {unlocked > 0 ? ` (${unlocked} unlocked)` : ""}
                      </p>
                      <div className={styles.stampDots}>
                        {Array.from({ length: Math.min(card.stampsRequired, 10) }).map((_, i) => (
                          <span
                            key={i}
                            className={classNames(styles.stampDot, {
                              [styles["stampDot--filled"]]: i < earned,
                            })}
                          >
                            {i < earned ? "★" : i + 1}
                          </span>
                        ))}
                      </div>
                    </div>
                  );
                })}
                <button
                  type="button"
                  className={styles.sheetPrimaryBtn}
                  onClick={() => setActiveSheet(null)}
                >
                  {t.gotIt}
                </button>
              </div>
            )}

            {activeSheet === "crm" && (
              <form className={styles.sheetBody} onSubmit={(e) => void handleSubmitCrmGuest(e)}>
                <h3 className={styles.sheetTitle}>{t.crmGuestTitle}</h3>
                <p style={{ margin: "0 0 12px", fontSize: "0.875rem", opacity: 0.72 }}>
                  {t.crmGuestSubtitle}
                </p>
                <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                  <input
                    type="text"
                    className={styles.sheetInput}
                    value={crmName}
                    onChange={(e) => setCrmName(e.target.value)}
                    placeholder={t.crmGuestName}
                    aria-label={t.crmGuestName}
                    style={{ margin: 0 }}
                  />
                  <input
                    type="tel"
                    className={styles.sheetInput}
                    value={crmPhone}
                    onChange={(e) => setCrmPhone(e.target.value)}
                    placeholder={t.crmGuestPhone}
                    aria-label={t.crmGuestPhone}
                    style={{ margin: 0 }}
                  />
                  <input
                    type="email"
                    className={styles.sheetInput}
                    value={crmEmail}
                    onChange={(e) => setCrmEmail(e.target.value)}
                    placeholder={t.crmGuestEmail}
                    aria-label={t.crmGuestEmail}
                    style={{ margin: 0 }}
                  />
                  <label
                    style={{
                      display: "flex",
                      alignItems: "flex-start",
                      gap: 8,
                      fontSize: "0.8rem",
                      opacity: 0.85,
                      cursor: "pointer",
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={crmConsent}
                      onChange={(e) => setCrmConsent(e.target.checked)}
                      style={{ marginTop: 2 }}
                    />
                    <span>{t.crmGuestConsent}</span>
                  </label>
                  {crmError && (
                    <p role="alert" style={{ margin: 0, fontSize: "0.8rem", color: "#dc2626" }}>
                      {crmError}
                    </p>
                  )}
                  <button
                    type="submit"
                    className={styles.sheetPrimaryBtn}
                    disabled={crmSubmitting || !crmConsent}
                  >
                    {crmSubmitting ? t.crmGuestSubmitting : t.crmGuestSubmit}
                  </button>
                </div>
              </form>
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
            aria-modal="true"
            aria-label={detailItem.name}
          >
            <button
              type="button"
              className={styles.sheetCloseBtn}
              aria-label="Close item details"
              onClick={() => setDetailItem(null)}
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
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
                      aria-label="Decrease quantity"
                      onClick={() => setDetailQty((q) => Math.max(1, q - 1))}
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                        <line x1="5" y1="12" x2="19" y2="12" />
                      </svg>
                    </button>
                    <span>{detailQty}</span>
                    <button
                      type="button"
                      aria-label="Increase quantity"
                      onClick={() => setDetailQty((q) => q + 1)}
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                        <line x1="12" y1="5" x2="12" y2="19" />
                        <line x1="5" y1="12" x2="19" y2="12" />
                      </svg>
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
                    {t.addToOrder} ·{" "}
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
          activeCurrency={activeCurrency}
          labels={t}
          onCheckout={(items) => {
            setCartOpen(false);
            if (previewMode || !onCheckout) {
              setTableActionNotice(t.previewCheckoutNotice);
              setTimeout(() => setTableActionNotice(null), 3500);
              return;
            }
            onCheckout(items);
          }}
        />
      )}

      {/* AI chat */}
      {enableAiChat && features.enableAiAssistant && onAiMessage && (
        <AiChat
          onSendMessage={onAiMessage}
          businessName={business.name}
          buttonLabel={t.askAi}
          placeholder={t.aiPlaceholder}
        />
      )}
    </div>
  );
};

export default MenuLayoutCore;
