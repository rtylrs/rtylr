/**
 * The locale registry. One source of truth for which languages exist,
 * which are shipped, and what each one needs in the document.
 *
 * Everything derives from this: the routes that get exported, the `lang`
 * and `dir` on `<html>`, the hreflang set in the sitemap, and the language
 * switcher. Nothing else may keep its own list, because a second list is
 * how a locale ends up in the sitemap without a route behind it.
 *
 * ## `shipped` is the gate, and it is the important part
 *
 * A locale route that renders English under its own `lang` declaration is a
 * doorway page. Google's own words: "Localized versions of a page are only
 * considered duplicates if the main content of the page remains
 * untranslated." So the flag is not a feature toggle, it is a promise that
 * the copy exists.
 *
 * `shipped: false` means:
 *   - no route is exported for it
 *   - it is absent from the sitemap and from every hreflang set
 *   - the language switcher does not offer it
 *
 * Flip one to `true` only when `locales/<code>.json` actually covers the
 * pages being shipped, and the copy has been through the review its
 * language requires. `tests/unit/locales.test.mjs` asserts the coupling
 * both ways: nothing unshipped may appear in the build, and anything
 * shipped must have enough translated keys to be worth serving.
 *
 * Arabic and French are written, not translated. A translated page reads
 * like one and ranks like one.
 */

export interface Locale {
  /** ISO 639-1 code. Also the URL segment, except for the default. */
  code: 'en' | 'ar' | 'fr';
  /** Value for the `lang` attribute. */
  htmlLang: string;
  /** Writing direction. Drives `dir` and the logical-property stylesheet. */
  dir: 'ltr' | 'rtl';
  /** Name in its own language, for the switcher. */
  endonym: string;
  /** og:locale. Underscore form, not the hyphen form hreflang uses. */
  ogLocale: string;
  /**
   * Whether this locale is served. See the note above: this is a claim
   * that the copy exists, not a feature flag.
   */
  shipped: boolean;
}

/**
 * The default locale is served from the root, not from `/en/`. Two URLs
 * for the same content is a canonical problem, and the site has been on
 * `/` since launch.
 */
export const DEFAULT_LOCALE = 'en' as const;

export const LOCALES: Locale[] = [
  {
    code: 'en',
    htmlLang: 'en',
    dir: 'ltr',
    endonym: 'English',
    ogLocale: 'en_US',
    shipped: true,
  },
  {
    code: 'ar',
    htmlLang: 'ar',
    dir: 'rtl',
    endonym: 'العربية',
    ogLocale: 'ar_AR',
    // locales/ar.json is empty. Serving /ar/ now would put English under
    // lang="ar", which is the doorway case above.
    shipped: false,
  },
  {
    code: 'fr',
    htmlLang: 'fr',
    dir: 'ltr',
    endonym: 'Français',
    ogLocale: 'fr_FR',
    // locales/fr.json is empty.
    shipped: false,
  },
];

export const shippedLocales = (): Locale[] => LOCALES.filter((l) => l.shipped);

/**
 * The routes a non-default locale actually covers.
 *
 * This exists because `shipped` is per locale but coverage is per page. The
 * first version of `alternatesFor` took only a path, so the moment a second
 * locale shipped it would have advertised an alternate for all 21 pages in
 * both languages while only these were translated. Fourteen of those URLs
 * would 404, and Google drops a hreflang set that fails to reciprocate
 * silently and in full, so the whole annotation would have been worth
 * nothing while looking correct in the sitemap.
 *
 * Paths are the canonical trailing-slash form, matching what the export
 * serves. Add a route here only when its namespace is complete in every
 * shipped locale.
 *
 * The legal pages are absent by decision, not omission. See the note at the
 * top of each file in pages/legal/.
 */
export const LOCALISED_ROUTES: string[] = [
  '/',
  '/pricing/',
  '/pos/',
  '/erp/',
  '/contact/',
];

/** Whether one route is translated in the non-default locales. */
export const isLocalisedRoute = (path: string): boolean => {
  const norm = path === '/' ? '/' : `${path.replace(/\/$/, '')}/`;
  return LOCALISED_ROUTES.includes(norm);
};

export const findLocale = (code: string): Locale | undefined =>
  LOCALES.find((l) => l.code === code);

/**
 * Which locale a pathname belongs to, from the first segment.
 *
 * Only shipped locales are recognised. An unshipped prefix is not a locale
 * route, so `/ar/pricing/` reads as the default locale, which is correct:
 * that route does not exist and should not be claimed.
 */
export function localeFromPath(pathname: string): Locale {
  const first = pathname.split('/').filter(Boolean)[0];
  const hit = first ? shippedLocales().find((l) => l.code === first) : undefined;
  return hit ?? (findLocale(DEFAULT_LOCALE) as Locale);
}

/**
 * The URL path for one route in one locale. The default locale keeps the
 * bare path; others take a prefix.
 *
 *   ('/pricing/', 'en') -> '/pricing/'
 *   ('/pricing/', 'ar') -> '/ar/pricing/'
 *   ('/', 'ar')         -> '/ar/'
 */
export function localisePath(path: string, code: string): string {
  const clean = path.startsWith('/') ? path : `/${path}`;
  if (code === DEFAULT_LOCALE) return clean;
  return clean === '/' ? `/${code}/` : `/${code}${clean}`;
}

/**
 * The hreflang set for one route: every shipped locale, including itself,
 * plus x-default.
 *
 * Self-reference and reciprocity are not optional. Google ignores a set
 * where any member fails to point back, and it does so silently, which is
 * the worst failure mode available. Generating the whole set from one
 * function is the only arrangement that cannot drift page by page.
 *
 * Returns an empty array while only one locale ships: a set of one is
 * noise, and next-sitemap would emit an alternate pointing at the page
 * itself and nothing else.
 */
export function alternatesFor(path: string, siteUrl = 'https://rtylr.com'): {
  hreflang: string;
  href: string;
}[] {
  const shipped = shippedLocales();
  if (shipped.length < 2) return [];

  // A page nobody translated has no alternates. Emitting them would point
  // at URLs that do not exist, and one broken member voids the whole set.
  if (!isLocalisedRoute(path)) return [];

  const set = shipped.map((l) => ({
    hreflang: l.htmlLang,
    href: `${siteUrl}${localisePath(path, l.code)}`,
  }));
  // x-default points at the default locale, for a visitor whose language
  // matches nothing shipped.
  set.push({ hreflang: 'x-default', href: `${siteUrl}${localisePath(path, DEFAULT_LOCALE)}` });
  return set;
}
