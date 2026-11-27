// ─── Chat ─────────────────────────────────────────────────────────────────────

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

// ─── Primitive option types ───────────────────────────────────────────────────

export type CoverHeight = 'none' | 'sm' | 'md' | 'lg';
export type NavStyle = 'pill' | 'underline' | 'flat';
export type NavSize = 'sm' | 'md' | 'lg';
export type ItemLayout = 'grid' | 'list' | 'minimal';
export type ImagePosition = 'top' | 'right' | 'left' | 'none';
export type ImageSize = 'sm' | 'md' | 'lg';
export type CardRadius = 'none' | 'sm' | 'md' | 'lg';
export type ThemeMode = 'light' | 'dark';
export type MenuFontSize = 'sm' | 'md' | 'lg';
export type FulfillmentMode = 'dine_in' | 'self_service' | 'pickup' | 'delivery';
export type TableActionType = 'call_waiter' | 'request_bill' | 'split_bill';

// ─── Business & Merchandising Data ────────────────────────────────────────────

export interface BusinessReview {
  id: string;
  author: string;
  avatar?: string;
  rating: number; // 1–5
  comment: string;
  date: string;
}

export interface BusinessOffer {
  id: string;
  title: string;
  description: string;
  image?: string;
  badge?: string;
  validUntil?: string;
}

export interface MenuBannerData {
  id: string;
  title: string;
  body?: string;
  imageUrl?: string;
  linkUrl?: string;
}

export interface MenuCurrencyOption {
  code: string;
  symbol: string;
  exchangeRate: number;
  decimalPlaces?: number;
  isDefault?: boolean;
}

export interface MenuStampCardSummary {
  id: string;
  title: string;
  description?: string;
  stampsRequired: number;
  rewardLabel: string;
}

export interface CustomerStampProgressData {
  stampCardId: string;
  stampsCount: number;
  stampsRequired: number;
  rewardsUnlocked: number;
  rewardLabel?: string;
}

export interface CrmGuestInput {
  name?: string;
  phone?: string;
  email?: string;
  notes?: string;
  consent: boolean;
}

export interface MenuLayoutLabels {
  searchMenu: string;
  searchPlaceholder: string;
  showResults: string;
  dietaryFilters: string;
  dietaryPreferences: string;
  allItems: string;
  applyFilter: string;
  languageAndCurrency: string;
  language: string;
  currency: string;
  done: string;
  tableService: string;
  service: string;
  callStaff: string;
  callStaffSubtitle: string;
  requestBill: string;
  requestBillSubtitle: string;
  splitBill: string;
  splitBillSubtitle: string;
  callWaiterSuccess: string;
  requestBillSuccess: string;
  splitBillSuccess: string;
  tableActionError: string;
  rewards: string;
  digitalStampCards: string;
  stampsPhonePlaceholder: string;
  checkStamps: string;
  checkingStamps: string;
  gotIt: string;
  viewOrder: string;
  addToOrder: string;
  yourOrder: string;
  closeCart: string;
  emptyCart: string;
  subtotal: string;
  placeOrder: string;
  previewCheckoutNotice: string;
  askAi: string;
  aiPlaceholder: string;
  joinClub: string;
  crmGuestTitle: string;
  crmGuestSubtitle: string;
  crmGuestName: string;
  crmGuestPhone: string;
  crmGuestEmail: string;
  crmGuestConsent: string;
  crmGuestSubmit: string;
  crmGuestSubmitting: string;
  crmGuestSuccess: string;
  crmGuestError: string;
  crmGuestContactRequired: string;
}

export interface BusinessInfo {
  name: string;
  logo?: string;
  coverImage?: string;
  faviconUrl?: string;
  description?: string;
  address?: string;
  phone?: string;
  email?: string;
  hours?: string;
  website?: string;
  story?: string;
  photos?: string[];
  socials?: {
    instagram?: string;
    facebook?: string;
    tiktok?: string;
  };
  reviews?: BusinessReview[];
  averageRating?: number;
  reviewCount?: number;
  offers?: BusinessOffer[];
}

// ─── Layout & Brand Config ────────────────────────────────────────────────────

export interface PageConfig {
  menu?: boolean;
  about?: boolean;
  gallery?: boolean;
  reviews?: boolean;
  offers?: boolean;
}

export interface HeaderConfig {
  coverHeight: CoverHeight;
  coverImageUrl?: string;
  logoUrl?: string;
  showLogo: boolean;
  showDescription: boolean;
  showMeta: boolean; // address / phone / hours
  sticky: boolean;
}

export interface NavConfig {
  style: NavStyle;
  size: NavSize;
}

export interface ItemDisplayConfig {
  layout: ItemLayout;
  columns: 2 | 3 | 4;
  imagePosition: ImagePosition;
  imageSize: ImageSize;
  showDescription: boolean;
  showBadges: boolean;
  cardRadius: CardRadius;
}

export interface ColorsConfig {
  primary: string;
  accent: string;
  bg: string;
  secondary?: string;
  surface?: string;
  text?: string;
  subtext?: string;
  border?: string;
}

export interface TypographyConfig {
  headingFont: string;
  bodyFont: string;
  fontSize: MenuFontSize;
  customFontCssUrl?: string;
}

export interface FulfillmentModesConfig {
  dineIn: boolean;
  selfService: boolean;
  pickup: boolean;
  delivery: boolean;
}

export interface TableActionsConfig {
  callWaiter: boolean;
  requestBill: boolean;
  splitBill: boolean;
}

export interface MenuFeaturesConfig {
  enableOrdering: boolean;
  fulfillmentModes: FulfillmentModesConfig;
  tableActions: TableActionsConfig;
  enableLoyaltyStamps: boolean;
  enableCrmCapture: boolean;
  enableAiAssistant: boolean;
  enableBanners: boolean;
  enableUpsells: boolean;
  enableDietaryFilters: boolean;
  enableCurrencySwitcher: boolean;
  enableLanguageSwitcher: boolean;
  enableSearch: boolean;
}

export interface MenuPaymentConfig {
  mode: 'pos' | 'redirect';
  redirectUrl?: string;
}

export interface AboutPageContent {
  story?: string;
  hours?: string;
  address?: string;
  phone?: string;
  email?: string;
  website?: string;
  socials?: {
    instagram?: string;
    facebook?: string;
    tiktok?: string;
  };
}

export interface MenuPageContentConfig {
  about?: AboutPageContent;
  gallery?: string[];
  offers?: BusinessOffer[];
  includeBannersInOffers?: boolean;
}

export interface MenuLayoutConfig {
  pages: PageConfig;
  header: HeaderConfig;
  nav: NavConfig;
  items: ItemDisplayConfig;
  colors?: ColorsConfig;
  typography?: TypographyConfig;
  mode?: ThemeMode;
  features?: Partial<MenuFeaturesConfig>;
  payment?: MenuPaymentConfig;
  content?: MenuPageContentConfig;
  customCss?: string;
}

export const DEFAULT_FEATURES_CONFIG: MenuFeaturesConfig = {
  enableOrdering: true,
  fulfillmentModes: {
    dineIn: true,
    selfService: true,
    pickup: true,
    delivery: false,
  },
  tableActions: {
    callWaiter: true,
    requestBill: true,
    splitBill: false,
  },
  enableLoyaltyStamps: true,
  enableCrmCapture: true,
  enableAiAssistant: true,
  enableBanners: true,
  enableUpsells: true,
  enableDietaryFilters: true,
  enableCurrencySwitcher: true,
  enableLanguageSwitcher: true,
  enableSearch: true,
};

export const DEFAULT_LAYOUT_CONFIG: MenuLayoutConfig = {
  pages: {
    menu: true,
    about: false,
    gallery: false,
    reviews: false,
    offers: false,
  },
  header: {
    coverHeight: 'md',
    showLogo: true,
    showDescription: true,
    showMeta: true,
    sticky: true,
  },
  nav: {
    style: 'underline',
    size: 'md',
  },
  items: {
    layout: 'grid',
    columns: 2,
    imagePosition: 'top',
    imageSize: 'md',
    showDescription: true,
    showBadges: true,
    cardRadius: 'md',
  },
  typography: {
    headingFont: 'system',
    bodyFont: 'system',
    fontSize: 'md',
  },
  mode: 'light',
  features: DEFAULT_FEATURES_CONFIG,
  payment: {
    mode: 'pos',
    redirectUrl: '',
  },
};

export const GRID_PRESET: Partial<MenuLayoutConfig> = {
  items: {
    layout: 'grid',
    columns: 2,
    imagePosition: 'top',
    imageSize: 'md',
    showDescription: true,
    showBadges: true,
    cardRadius: 'md',
  },
};

export const LIST_PRESET: Partial<MenuLayoutConfig> = {
  items: {
    layout: 'list',
    columns: 2,
    imagePosition: 'right',
    imageSize: 'md',
    showDescription: true,
    showBadges: true,
    cardRadius: 'md',
  },
};

export const MINIMAL_PRESET: Partial<MenuLayoutConfig> = {
  items: {
    layout: 'minimal',
    columns: 2,
    imagePosition: 'none',
    imageSize: 'sm',
    showDescription: true,
    showBadges: false,
    cardRadius: 'none',
  },
};
