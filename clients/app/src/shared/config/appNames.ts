export interface AppName {
  /** Short brand name, for badges and permission rows. */
  readonly short: string;
  /** Descriptive name, for surfaces that explain what the app does. */
  readonly full: string;
}

const name = (id: string, descriptive = false): AppName => ({
  short: `appNames.short.${id}`,
  full: descriptive ? `appNames.full.${id}` : `appNames.short.${id}`,
});

export const APP_NAMES = {
  pos: name('pos', true),
  erp: name('erp', true),
  hr: name('hr', true),
  crm: name('crm', true),
  finance: name('finance'),
  flow: name('flow'),
  insights: name('insights'),
  menu: name('menu'),
  recruit: name('recruit'),
  storefront: name('storefront'),
  market: name('market'),
  web: name('web'),
  dash: name('dash'),
} as const satisfies Readonly<Record<string, AppName>>;

export type AppId = keyof typeof APP_NAMES;

/** Every app a member can be granted, in the order the forms list them. */
export const GRANTABLE_APPS = [
  'pos',
  'erp',
  'hr',
  'crm',
  'finance',
  'flow',
  'insights',
  'menu',
  'recruit',
] as const satisfies readonly AppId[];
