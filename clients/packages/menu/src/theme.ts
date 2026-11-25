import type React from 'react';
import {
  DEFAULT_FEATURES_CONFIG,
  type MenuCurrencyOption,
  type MenuFontSize,
  type MenuLayoutConfig,
} from './types';

export const MENU_FONT_STACKS: Record<string, string> = {
  system:
    'system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  inter: '"Inter", system-ui, -apple-system, sans-serif',
  jakarta: '"Plus Jakarta Sans", system-ui, -apple-system, sans-serif',
  poppins: '"Poppins", system-ui, -apple-system, sans-serif',
  merriweather: '"Merriweather", Georgia, serif',
  playfair: '"Playfair Display", Georgia, serif',
  cairo: '"Cairo", "Segoe UI", Tahoma, sans-serif',
  tajawal: '"Tajawal", "Segoe UI", Tahoma, sans-serif',
};

const MENU_GOOGLE_FONTS: Record<string, string> = {
  inter: 'family=Inter:wght@400;500;600;700',
  jakarta: 'family=Plus+Jakarta+Sans:wght@400;500;600;700',
  poppins: 'family=Poppins:wght@400;500;600;700',
  merriweather: 'family=Merriweather:wght@400;700',
  playfair: 'family=Playfair+Display:wght@600;700',
  cairo: 'family=Cairo:wght@400;600;700',
  tajawal: 'family=Tajawal:wght@400;500;700',
};

/**
 * Idempotently injects a Google Fonts stylesheet for a menu font token.
 * Mirrors the webfont loader used by Recruit career templates.
 */
export function loadMenuFont(fontId?: string): void {
  if (!fontId || typeof document === 'undefined') return;
  const query = MENU_GOOGLE_FONTS[fontId];
  if (!query) return;
  const markerId = `menu-font-${fontId}`;
  if (document.getElementById(markerId)) return;
  const link = document.createElement('link');
  link.id = markerId;
  link.rel = 'stylesheet';
  link.href = `https://fonts.googleapis.com/css2?${query}&display=swap`;
  document.head.appendChild(link);
}

const FONT_SCALE_MAP: Record<MenuFontSize, string> = {
  sm: '14px',
  md: '16px',
  lg: '18px',
};

const RTL_LOCALES = new Set(['ar', 'he', 'fa', 'ur']);

export function isRtlLocale(locale?: string): boolean {
  if (!locale) return false;
  const base = locale.trim().toLowerCase().split(/[-_]/)[0];
  return RTL_LOCALES.has(base);
}

function contrastTextForHex(hex: string | undefined, fallback = '#ffffff'): string {
  if (!hex) return fallback;
  const cleaned = hex.replace('#', '').trim();
  const normalized =
    cleaned.length === 3
      ? cleaned
          .split('')
          .map((c) => c + c)
          .join('')
      : cleaned;
  if (!/^[0-9a-fA-F]{6}$/.test(normalized)) return fallback;
  const r = parseInt(normalized.slice(0, 2), 16);
  const g = parseInt(normalized.slice(2, 4), 16);
  const b = parseInt(normalized.slice(4, 6), 16);
  // Relative luminance heuristic (WCAG-inspired threshold)
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return luminance > 0.62 ? '#111827' : '#ffffff';
}

export function buildMenuThemeStyle(config: MenuLayoutConfig): React.CSSProperties {
  const isDark = config.mode === 'dark';
  const colors = config.colors;
  const typography = config.typography;

  const bg = colors?.bg || (isDark ? '#141414' : '#ffffff');
  const surface = colors?.surface || (isDark ? '#1c1c1c' : '#ffffff');
  const text = colors?.text || (isDark ? '#f5f5f5' : '#111111');
  const heading = colors?.primary || text;
  const subtext = colors?.subtext || (isDark ? 'rgba(245,245,245,0.64)' : '#6b6b6b');
  const accent = colors?.accent || colors?.primary || (isDark ? '#ffffff' : '#111111');
  const border = colors?.border || (isDark ? 'rgba(255,255,255,0.1)' : '#ebebeb');
  const imageBg = isDark ? '#242424' : '#f4f4f4';
  const bubbleBg = isDark ? 'rgba(255,255,255,0.08)' : '#f5f5f5';
  const inputBg = isDark ? 'rgba(255,255,255,0.06)' : '#f5f5f5';

  const headingFont = MENU_FONT_STACKS[typography?.headingFont ?? 'system'] ?? typography?.headingFont ?? MENU_FONT_STACKS.system;
  const bodyFont = MENU_FONT_STACKS[typography?.bodyFont ?? 'system'] ?? typography?.bodyFont ?? MENU_FONT_STACKS.system;
  const fontScale = FONT_SCALE_MAP[typography?.fontSize ?? 'md'] ?? '16px';

  return {
    '--menu-bg': bg,
    '--menu-card-bg': surface,
    '--menu-header-bg': surface,
    '--menu-border': border,
    '--menu-text': text,
    '--menu-heading': heading,
    '--menu-subtext': subtext,
    '--menu-price': heading,
    '--menu-accent': accent,
    '--menu-accent-text': contrastTextForHex(accent),
    '--menu-image-bg': imageBg,
    '--menu-bubble-bg': bubbleBg,
    '--menu-input-bg': inputBg,
    '--menu-font-heading': headingFont,
    '--menu-font-body': bodyFont,
    fontFamily: bodyFont,
    fontSize: fontScale,
  } as React.CSSProperties;
}

export function formatConvertedPrice(
  basePrice: number,
  currency?: MenuCurrencyOption,
  fallbackCurrency = 'USD',
): string {
  if (!currency) {
    try {
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: fallbackCurrency || 'USD',
      }).format(basePrice);
    } catch {
      return `${fallbackCurrency || 'USD'} ${basePrice.toFixed(2)}`;
    }
  }
  const rate = Number.isFinite(currency.exchangeRate) && currency.exchangeRate > 0 ? currency.exchangeRate : 1;
  const decimals = typeof currency.decimalPlaces === 'number' && currency.decimalPlaces >= 0 ? currency.decimalPlaces : 2;
  const converted = basePrice * rate;
  const formattedNumber = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(converted);
  return `${currency.symbol}${formattedNumber}`;
}

export interface MenuStarterTheme {
  id: string;
  name: string;
  subtitle: string;
  swatches: [string, string, string];
  locale?: string;
  patch: Partial<MenuLayoutConfig>;
}

export const MENU_ACCENT_SWATCHES: readonly string[] = [
  '#111111',
  '#059669',
  '#2563eb',
  '#dc2626',
  '#b45309',
  '#0f766e',
  '#4f46e5',
  '#374151',
];

export const MENU_STARTER_THEMES: readonly MenuStarterTheme[] = [
  {
    id: 'crisp-utility',
    name: 'Crisp Utility',
    subtitle: 'Pure white · High-contrast grid',
    swatches: ['#ffffff', '#f5f5f5', '#111111'],
    patch: {
      mode: 'light',
      colors: {
        primary: '#111111',
        accent: '#111111',
        bg: '#ffffff',
        surface: '#ffffff',
        text: '#111111',
      },
      typography: { headingFont: 'inter', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'md',
      },
    },
  },
  {
    id: 'compact-list',
    name: 'Compact List',
    subtitle: 'Fast scannable rows · Hairline dividers',
    swatches: ['#ffffff', '#f5f5f5', '#059669'],
    patch: {
      mode: 'light',
      colors: {
        primary: '#111111',
        accent: '#059669',
        bg: '#ffffff',
        surface: '#ffffff',
        text: '#111111',
      },
      typography: { headingFont: 'inter', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'list',
        columns: 2,
        imagePosition: 'right',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'sm',
      },
    },
  },
  {
    id: 'warm-paper',
    name: 'Warm Paper',
    subtitle: 'Off-white café · Calm neutral ink',
    swatches: ['#faf9f6', '#ffffff', '#2c1e16'],
    patch: {
      mode: 'light',
      colors: {
        primary: '#1c1917',
        accent: '#2c1e16',
        bg: '#faf9f6',
        surface: '#faf9f6',
        text: '#1c1917',
      },
      typography: { headingFont: 'jakarta', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'md',
      },
    },
  },
  {
    id: 'arabic-clean',
    name: 'Arabic Clean (RTL)',
    subtitle: 'Pure white · Native Cairo typography',
    swatches: ['#ffffff', '#f5f5f5', '#0f766e'],
    locale: 'ar',
    patch: {
      mode: 'light',
      colors: {
        primary: '#111111',
        accent: '#0f766e',
        bg: '#ffffff',
        surface: '#ffffff',
        text: '#111111',
      },
      typography: { headingFont: 'cairo', bodyFont: 'cairo', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'md',
      },
    },
  },
  {
    id: 'matte-dark',
    name: 'Matte Charcoal',
    subtitle: 'Neutral dark · Restrained monochrome',
    swatches: ['#141414', '#1c1c1c', '#ffffff'],
    patch: {
      mode: 'dark',
      colors: {
        primary: '#ffffff',
        accent: '#ffffff',
        bg: '#141414',
        surface: '#1c1c1c',
        text: '#f5f5f5',
      },
      typography: { headingFont: 'inter', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'md',
      },
    },
  },
];

export interface MenuStylePromptResult {
  patch: Partial<MenuLayoutConfig>;
  locale?: string;
}

export function applyMenuStylePrompt(
  rawPrompt: string,
  current: MenuLayoutConfig,
): MenuStylePromptResult {
  const prompt = rawPrompt.trim().toLowerCase();
  if (!prompt) return { patch: {} };

  const patch: Partial<MenuLayoutConfig> = {};
  let locale: string | undefined;

  if (prompt.includes('dark') || prompt.includes('luxury') || prompt.includes('noir')) {
    patch.mode = 'dark';
    patch.colors = {
      primary: '#f9fafb',
      accent: prompt.includes('gold') ? '#d4b483' : '#f59e0b',
      bg: '#0c0e14',
      surface: '#161923',
      text: '#f3f4f6',
    };
  } else if (prompt.includes('emerald') || prompt.includes('botanical') || prompt.includes('green')) {
    patch.mode = 'light';
    patch.colors = {
      primary: '#064e3b',
      accent: '#10b981',
      bg: '#f3f8f5',
      surface: '#ffffff',
      text: '#14332a',
    };
  } else if (prompt.includes('warm') || prompt.includes('cafe') || prompt.includes('beige') || prompt.includes('cream')) {
    patch.mode = 'light';
    patch.colors = {
      primary: '#2c1e16',
      accent: '#b45309',
      bg: '#faf6f0',
      surface: '#ffffff',
      text: '#292018',
    };
  }

  if (prompt.includes('minimal')) {
    patch.items = { ...current.items, layout: 'minimal', imagePosition: 'none' };
  } else if (prompt.includes('list') || prompt.includes('retail')) {
    patch.items = { ...current.items, layout: 'list', imagePosition: 'right' };
  } else if (prompt.includes('grid')) {
    patch.items = { ...current.items, layout: 'grid', imagePosition: 'top' };
  }

  const fontSize = current.typography?.fontSize ?? 'md';
  if (prompt.includes('cairo') || prompt.includes('arabic') || prompt.includes('rtl')) {
    patch.typography = { headingFont: 'cairo', bodyFont: 'cairo', fontSize };
    locale = 'ar';
  } else if (prompt.includes('serif') || prompt.includes('playfair')) {
    patch.typography = { headingFont: 'playfair', bodyFont: 'merriweather', fontSize };
  }

  if (prompt.includes('self-service') || prompt.includes('self service')) {
    const baseFeatures = {
      ...DEFAULT_FEATURES_CONFIG,
      ...current.features,
      fulfillmentModes: {
        ...DEFAULT_FEATURES_CONFIG.fulfillmentModes,
        ...current.features?.fulfillmentModes,
      },
      tableActions: {
        ...DEFAULT_FEATURES_CONFIG.tableActions,
        ...current.features?.tableActions,
      },
    };
    patch.features = {
      ...baseFeatures,
      fulfillmentModes: {
        ...baseFeatures.fulfillmentModes,
        dineIn: false,
        selfService: true,
      },
      tableActions: {
        callWaiter: false,
        requestBill: false,
        splitBill: false,
      },
    };
  }

  return { patch, locale };
}
