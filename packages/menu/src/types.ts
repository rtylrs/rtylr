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

// ─── Business data ─────────────────────────────────────────────────────────────

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

export interface BusinessInfo {
  name: string;
  logo?: string;
  coverImage?: string;
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

// ─── Layout config ─────────────────────────────────────────────────────────────

export interface PageConfig {
  menu?: boolean;
  about?: boolean;
  gallery?: boolean;
  reviews?: boolean;
  offers?: boolean;
}

export interface HeaderConfig {
  coverHeight: CoverHeight;
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
}

export interface MenuLayoutConfig {
  pages: PageConfig;
  header: HeaderConfig;
  nav: NavConfig;
  items: ItemDisplayConfig;
  colors?: ColorsConfig;
}

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
};

// Preset configs for convenience
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
