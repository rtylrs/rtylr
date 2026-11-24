import { useEffect } from 'react';

const DEFAULT_MANIFEST_HREF = '/site.webmanifest';

/**
 * Updates `<link rel="manifest">` to point `start_url` at the current customer
 * ticket or venue page. On iOS Safari, "Add to Home Screen" snapshots the
 * active manifest's `start_url` — without this, launching the installed PWA
 * would lose the customer's ticket token (because iOS isolates PWA storage
 * from Safari).
 */
export function useDynamicManifest(startPath: string | null, appName?: string): void {
  useEffect(() => {
    if (
      !startPath ||
      typeof document === 'undefined' ||
      typeof URL.createObjectURL !== 'function'
    ) {
      return undefined;
    }
    const origin = window.location.origin;
    if (!origin || origin === 'null') return undefined;

    const manifest = {
      name: appName ? `${appName} · Rennili` : 'Rennili',
      short_name: appName ? appName.slice(0, 12) : 'Rennili',
      start_url: new URL(startPath, origin).href,
      scope: `${origin}/`,
      display: 'standalone',
      background_color: '#f4f7f5',
      theme_color: '#0e4d3b',
      icons: [
        {
          src: `${origin}/favicon/android-chrome-192x192.png`,
          sizes: '192x192',
          type: 'image/png',
        },
        {
          src: `${origin}/favicon/android-chrome-512x512.png`,
          sizes: '512x512',
          type: 'image/png',
        },
        {
          src: `${origin}/favicon/maskable-512x512.png`,
          sizes: '512x512',
          type: 'image/png',
          purpose: 'maskable',
        },
      ],
    };

    const blob = new Blob([JSON.stringify(manifest)], { type: 'application/manifest+json' });
    const blobUrl = URL.createObjectURL(blob);

    let link = document.querySelector<HTMLLinkElement>('link[rel="manifest"]');
    const created = !link;
    const previousHref = link?.getAttribute('href') ?? DEFAULT_MANIFEST_HREF;
    if (!link) {
      link = document.createElement('link');
      link.rel = 'manifest';
      document.head.appendChild(link);
    }
    link.href = blobUrl;

    return () => {
      if (created) link?.remove();
      else if (link) link.href = previousHref;
      URL.revokeObjectURL(blobUrl);
    };
  }, [startPath, appName]);
}
