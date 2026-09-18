// Types
export type {
  ChatMessage,
  BusinessInfo,
  BusinessReview,
  BusinessOffer,
  PageConfig,
  HeaderConfig,
  NavConfig,
  ItemDisplayConfig,
  ColorsConfig,
  MenuLayoutConfig,
  CoverHeight,
  NavStyle,
  NavSize,
  ItemLayout,
  ImagePosition,
  ImageSize,
  CardRadius,
} from './types';
export { DEFAULT_LAYOUT_CONFIG, GRID_PRESET, LIST_PRESET, MINIMAL_PRESET } from './types';

// Components
export { default as MenuItem } from './components/MenuItem';
export type { MenuItemData } from './components/MenuItem';

export { default as MenuCategory } from './components/MenuCategory';
export type { MenuCategoryData } from './components/MenuCategory';

export { default as MenuHeader } from './components/MenuHeader';

export { default as MenuCart } from './components/MenuCart';
export type { CartItem } from './components/MenuCart';

export { default as AiChat } from './components/AiChat';

// Pages
export { default as AboutPage } from './pages/AboutPage';
export { default as GalleryPage } from './pages/GalleryPage';
export { default as ReviewsPage } from './pages/ReviewsPage';
export { default as OffersPage } from './pages/OffersPage';

// Layouts
export { default as MenuLayout } from './layouts/MenuLayoutCore';
export type { MenuLayoutCoreProps as MenuLayoutProps } from './layouts/MenuLayoutCore';
export { default as GridLayout } from './layouts/GridLayout';
export { default as ListLayout } from './layouts/ListLayout';
export { default as MinimalLayout } from './layouts/MinimalLayout';
