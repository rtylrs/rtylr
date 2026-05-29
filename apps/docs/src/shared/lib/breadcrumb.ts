/** Breadcrumb helpers for the docs app. */

export interface Crumb {
  label: string;
  href?: string;
}

/**
 * Builds a breadcrumb trail from a URL path, e.g. "/guides/getting-started".
 */
export function pathToCrumbs(path: string, base = ''): Crumb[] {
  const parts = path
    .replace(/^\/|\/$/g, '')
    .split('/')
    .filter(Boolean);
  const crumbs: Crumb[] = [];
  let current = base;
  for (const part of parts) {
    current = `${current}/${part}`;
    const label = part
      .split('-')
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ');
    crumbs.push({ label, href: current });
  }
  return crumbs;
}

/**
 * Returns only the labels from a crumb list.
 */
export function crumbLabels(crumbs: Crumb[]): string[] {
  return crumbs.map((c) => c.label);
}

/**
 * Returns the last crumb (current page).
 */
export function currentCrumb(crumbs: Crumb[]): Crumb | undefined {
  return crumbs[crumbs.length - 1];
}

/**
 * Returns all crumbs except the last one (ancestors).
 */
export function ancestorCrumbs(crumbs: Crumb[]): Crumb[] {
  return crumbs.slice(0, -1);
}
