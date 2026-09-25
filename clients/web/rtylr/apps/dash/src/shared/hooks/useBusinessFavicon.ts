import { useEffect } from 'react';

const DEFAULT_ICON_HREFS: Record<string, string> = {
  'image/x-icon': '/favicon.ico',
  'image/png:32x32': '/favicon/favicon-32x32.png',
  'image/png:16x16': '/favicon/favicon-16x16.png',
};

function iconLinks(): HTMLLinkElement[] {
  return Array.from(document.querySelectorAll<HTMLLinkElement>('link[rel="icon"]'));
}

function iconKey(link: HTMLLinkElement): string {
  const type = link.getAttribute('type') ?? '';
  const sizes = link.getAttribute('sizes');
  return sizes ? `${type}:${sizes}` : type;
}

/** Swaps the browser tab favicon to the active business's logo, falling back to the default rtylr icon. */
export function useBusinessFavicon(logoUrl?: string): void {
  useEffect(() => {
    const links = iconLinks();
    for (const link of links) {
      link.href = logoUrl || (DEFAULT_ICON_HREFS[iconKey(link)] ?? link.href);
    }
  }, [logoUrl]);
}
