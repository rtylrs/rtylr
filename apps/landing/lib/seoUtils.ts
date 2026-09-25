/**
 * SEO utility helpers for the landing app.
 * Pure functions with no browser or Next.js dependencies.
 */

const BASE_URL = 'https://rtylr.com';
const DEFAULT_TITLE = 'rtylr - One Suite for Your Entire Business';
const DEFAULT_DESCRIPTION =
  'The rtylr suite connects sales, inventory, people, customers, finance and operations in one workspace.';

/**
 * Builds the full page title, appending the site name when the page
 * title differs from the default.
 */
export function buildPageTitle(pageTitle?: string): string {
  if (!pageTitle || pageTitle === DEFAULT_TITLE) return DEFAULT_TITLE;
  return `${pageTitle} | rtylr`;
}

/**
 * Truncates a meta description to the recommended 155-character limit.
 * Returns the default description when none is provided.
 */
export function buildMetaDescription(description?: string): string {
  const desc = description ?? DEFAULT_DESCRIPTION;
  if (desc.length <= 155) return desc;
  return `${desc.slice(0, 152)}...`;
}

/**
 * Builds the canonical URL for a page path.
 * Always adds a trailing slash to match the canonical format.
 */
export function buildCanonicalUrl(path: string): string {
  const normalized = path.startsWith('/') ? path : `/${path}`;
  const withSlash = normalized.endsWith('/') ? normalized : `${normalized}/`;
  return `${BASE_URL}${withSlash}`;
}

/**
 * Returns the full absolute URL for an OG image path.
 */
export function buildOgImageUrl(imagePath?: string): string {
  if (!imagePath) return `${BASE_URL}/og-default.png`;
  if (imagePath.startsWith('http')) return imagePath;
  const normalized = imagePath.startsWith('/') ? imagePath : `/${imagePath}`;
  return `${BASE_URL}${normalized}`;
}

/**
 * Generates the hreflang alternate links object for a page path.
 * rtylr supports English (default) and Arabic.
 */
export function buildHreflangLinks(path: string): Record<string, string> {
  const canonical = buildCanonicalUrl(path);
  return {
    'x-default': canonical,
    en: canonical,
    ar: `${BASE_URL}/ar${path.startsWith('/') ? path : `/${path}`}`,
  };
}

/**
 * Converts a page path to a human-readable breadcrumb label.
 * E.g. "/flow/" -> "Flow", "/hr/pricing/" -> "HR Pricing".
 */
export function pathToBreadcrumbLabel(path: string): string {
  const segments = path
    .split('/')
    .filter(Boolean)
    .map((s) =>
      s
        .split('-')
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ')
    );
  return segments.join(' ') || 'Home';
}

/**
 * Returns true when a meta description is within the recommended length range.
 */
export function isMetaDescriptionValid(description: string): boolean {
  return description.length >= 50 && description.length <= 160;
}

/**
 * Returns true when a page title is within the recommended length range.
 */
export function isPageTitleValid(title: string): boolean {
  return title.length >= 10 && title.length <= 70;
}
