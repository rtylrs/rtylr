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

  const bg = colors?.bg || (isDark ? '#0f1115' : '#f9f9f7');
  const surface = colors?.surface || (isDark ? '#181b22' : '#ffffff');
  const text = colors?.text || (isDark ? '#f3f4f6' : '#1a1a1a');
  const heading = colors?.primary || text;
  const subtext = colors?.subtext || (isDark ? 'rgba(243,244,246,0.68)' : 'rgba(26,26,26,0.62)');
  const accent = colors?.accent || colors?.primary || (isDark ? '#f59e0b' : '#1a1a1a');
  const border = colors?.border || (isDark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.08)');
  const imageBg = isDark ? '#222631' : '#f0f0ee';
  const bubbleBg = isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.05)';
  const inputBg = isDark ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.03)';

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
  '#c8a96e',
  '#4583c4',
  '#b45309',
  '#10b981',
  '#e07a5f',
  '#8b5cf6',
  '#e11d48',
  '#111827',
];

export const MENU_STARTER_THEMES: readonly MenuStarterTheme[] = [
  {
    id: 'noir-bistro',
    name: 'Noir Bistro',
    subtitle: 'Dark fine dining · Gold accent',
    swatches: ['#0c0e14', '#161923', '#d4b483'],
    patch: {
      mode: 'dark',
      colors: {
        primary: '#f8fafc',
        accent: '#d4b483',
        bg: '#0c0e14',
        surface: '#161923',
        text: '#f3f4f6',
      },
      typography: { headingFont: 'playfair', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'lg',
      },
    },
  },
  {
    id: 'warm-artisan',
    name: 'Warm Artisan Café',
    subtitle: 'Cream editorial · Espresso tone',
    swatches: ['#faf6f0', '#ffffff', '#b45309'],
    patch: {
      mode: 'light',
      colors: {
        primary: '#2c1e16',
        accent: '#b45309',
        bg: '#faf6f0',
        surface: '#ffffff',
        text: '#292018',
      },
      typography: { headingFont: 'merriweather', bodyFont: 'jakarta', fontSize: 'md' },
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
    id: 'crisp-commerce',
    name: 'Crisp Commerce',
    subtitle: 'High-speed retail & QSR list',
    swatches: ['#f8fafc', '#ffffff', '#2563eb'],
    patch: {
      mode: 'light',
      colors: {
        primary: '#0f172a',
        accent: '#2563eb',
        bg: '#f8fafc',
        surface: '#ffffff',
        text: '#1e293b',
      },
      typography: { headingFont: 'jakarta', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'list',
        columns: 2,
        imagePosition: 'right',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'md',
      },
    },
  },
  {
    id: 'botanical-fresh',
    name: 'Botanical Fresh',
    subtitle: 'Organic bowls · Soft mint',
    swatches: ['#f3f8f5', '#ffffff', '#10b981'],
    patch: {
      mode: 'light',
      colors: {
        primary: '#064e3b',
        accent: '#10b981',
        bg: '#f3f8f5',
        surface: '#ffffff',
        text: '#14332a',
      },
      typography: { headingFont: 'poppins', bodyFont: 'inter', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'lg',
      },
    },
  },
  {
    id: 'beirut-rtl',
    name: 'Beirut Contemporary (RTL)',
    subtitle: 'Native Arabic Cairo · Terracotta',
    swatches: ['#12100e', '#1c1815', '#e07a5f'],
    locale: 'ar',
    patch: {
      mode: 'dark',
      colors: {
        primary: '#faf3eb',
        accent: '#e07a5f',
        bg: '#12100e',
        surface: '#1c1815',
        text: '#f5ece3',
      },
      typography: { headingFont: 'cairo', bodyFont: 'cairo', fontSize: 'md' },
      items: {
        layout: 'grid',
        columns: 2,
        imagePosition: 'top',
        imageSize: 'md',
        showDescription: true,
        showBadges: true,
        cardRadius: 'lg',
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
