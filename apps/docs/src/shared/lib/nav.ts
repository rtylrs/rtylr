import type { DocContent } from '../content';

export interface NavItem {
  slug: string;
  title: string;
  active: boolean;
}

export interface Breadcrumb {
  label: string;
  href: string;
}

/**
 * Builds a sorted list of navigation items from the content map,
 * marking the current slug as active.
 */
export function buildNavItems(
  contentMap: Record<string, DocContent>,
  currentSlug: string
): NavItem[] {
  return Object.keys(contentMap)
    .sort()
    .map((slug) => ({
      slug,
      title: contentMap[slug].title,
      active: slug === currentSlug,
    }));
}

/**
 * Returns the slug that comes before the current one in sorted order,
 * or null if the current slug is first.
 */
export function prevSlug(
  contentMap: Record<string, DocContent>,
  currentSlug: string
): string | null {
  const slugs = Object.keys(contentMap).sort();
  const idx = slugs.indexOf(currentSlug);
  if (idx <= 0) return null;
  return slugs[idx - 1];
}

/**
 * Returns the slug that comes after the current one in sorted order,
 * or null if the current slug is last.
 */
export function nextSlug(
  contentMap: Record<string, DocContent>,
  currentSlug: string
): string | null {
  const slugs = Object.keys(contentMap).sort();
  const idx = slugs.indexOf(currentSlug);
  if (idx === -1 || idx === slugs.length - 1) return null;
  return slugs[idx + 1];
}

/**
 * Generates breadcrumb entries for a given slug.
 * Returns a root "Docs" entry plus the current page entry.
 */
export function getBreadcrumbs(contentMap: Record<string, DocContent>, slug: string): Breadcrumb[] {
  const crumbs: Breadcrumb[] = [{ label: 'Docs', href: '/docs' }];
  const doc = contentMap[slug];
  if (doc) {
    crumbs.push({ label: doc.title, href: `/docs/${slug}` });
  }
  return crumbs;
}
