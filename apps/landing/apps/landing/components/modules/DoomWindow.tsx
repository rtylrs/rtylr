import React, { useEffect, useRef } from 'react';

declare global {
  interface Window {
    Dos?: (el: HTMLElement, options: Record<string, unknown>) => { stop?: () => void } | undefined;
  }
}

const JSDOS_JS = '/jsdos/js-dos.js';
const JSDOS_CSS = '/jsdos/js-dos.css';

function loadJsDos(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined') {
      resolve();
      return;
    }
    if (!document.querySelector(`link[href="${JSDOS_CSS}"]`)) {
      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = JSDOS_CSS;
      document.head.appendChild(link);
    }
    if (window.Dos) {
      resolve();
      return;
    }
    const existing = document.querySelector(`script[src="${JSDOS_JS}"]`) as HTMLScriptElement | null;
    if (existing) {
      existing.addEventListener('load', () => resolve());
      existing.addEventListener('error', () => reject(new Error('js-dos failed to load')));
      return;
    }
    const script = document.createElement('script');
    script.src = JSDOS_JS;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('js-dos failed to load'));
    document.head.appendChild(script);
  });
}

export default function DoomWindow(): React.JSX.Element {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let ci: { stop?: () => void } | undefined;
    let cancelled = false;

    loadJsDos()
      .then(() => {
        if (cancelled || !ref.current || !window.Dos) return;
        ci = window.Dos(ref.current, {
          url: '/doom/doom.jsdos',
          pathPrefix: '/jsdos/emulators/',
          noCloud: true,
          autoStart: true,
        });
      })
      .catch(() => {});

    return () => {
      cancelled = true;
      try {
        ci?.stop?.();
      } catch {
        /* noop */
      }
    };
  }, []);

  return <div ref={ref} style={{ height: '100%', width: '100%', background: '#000' }} />;
}
