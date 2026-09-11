/**
 * The name of every rtylr app, in one place, as i18n keys.
 *
 * Dash is the hub, so it names the other apps more than anywhere else does:
 * the app badge, the business switcher, the invite form, the member detail
 * screen, the permission resource table and the vendor dashboard each carried
 * their own copy. Six tables, one set of products.
 *
 * Two vocabularies are deliberate and both are kept. Most surfaces want the
 * short brand name, which is what a badge has room for and what the permission
 * rows read as. The vendor dashboard explains what each app is, so it wants the
 * descriptive name: "Point of Sale" rather than "POS", "Human Resources"
 * rather than "HR". An app whose descriptive name is its brand name carries one
 * key for both.
 *
 * These are product names, so `en`, `ar` and `fr` currently hold the same
 * string for most of them. Routing them through i18n anyway costs nothing and
 * means localising a name later is a locale edit rather than a code change.
 */
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
