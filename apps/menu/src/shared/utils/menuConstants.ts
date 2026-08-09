/**
 * Named constants for the menu application.
 * Centralises magic strings and numbers used across components.
 */

/** Default currency code when no business currency is available. */
export const DEFAULT_CURRENCY = 'USD';

/** Default locale for number and date formatting. */
export const DEFAULT_LOCALE = 'en-US';

/** Polling interval in milliseconds for order status updates. */
export const ORDER_POLL_INTERVAL_MS = 30_000;

/** Maximum number of notes characters for a single cart item. */
export const ITEM_NOTE_MAX_LENGTH = 200;

/** Maximum number of characters for order-level notes. */
export const ORDER_NOTE_MAX_LENGTH = 500;

/** sessionStorage key used by the cart Zustand store. */
export const CART_STORAGE_KEY = 'rtylr-cart';

/** localStorage key used by the theme Zustand store. */
export const THEME_STORAGE_KEY = 'rtylr-menu-theme';

/** Minimum tap target size in pixels (WCAG 2.5.5). */
export const MIN_TAP_TARGET_PX = 44;

/** CSS variable names consumed by the menu app. */
export const CSS_VARS = {
  accent: '--menu-accent',
  accentText: '--menu-accent-text',
  background: '--menu-bg',
} as const;
