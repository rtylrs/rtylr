import React, { useEffect, useRef } from 'react';

const JSDOS_JS = 'https://v8.js-dos.com/latest/js-dos.js';
const JSDOS_CSS = 'https://v8.js-dos.com/latest/js-dos.css';
const DOOM_BUNDLE =
  'https://cdn.dos.zone/original/2X/1/11fc26d58f8869aee0e34ce2bb1f4b04543ee2b3.jsdos';

declare global {
  interface Window {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Dos?: (el: HTMLElement, opts: Record<string, unknown>) => void;
  }
}

export default function DoomWindow(): React.JSX.Element {
  const containerRef = useRef<HTMLDivElement>(null);
  const initializedRef = useRef(false);

  useEffect(() => {
    if (initializedRef.current) return;

    let cssEl: HTMLLinkElement | null = null;
    let scriptEl: HTMLScriptElement | null = null;

    const initDos = (): void => {
      if (!containerRef.current || !window.Dos) return;
      initializedRef.current = true;
      window.Dos(containerRef.current, {
        url: DOOM_BUNDLE,
        theme: 'dark',
        noCloud: true,
      });
    };

    if (!document.querySelector(`link[href="${JSDOS_CSS}"]`)) {
      cssEl = document.createElement('link');
      cssEl.rel = 'stylesheet';
      cssEl.href = JSDOS_CSS;
      document.head.appendChild(cssEl);
    }

    if (window.Dos) {
      initDos();
    } else if (!document.querySelector(`script[src="${JSDOS_JS}"]`)) {
      scriptEl = document.createElement('script');
      scriptEl.src = JSDOS_JS;
      scriptEl.onload = initDos;
      document.head.appendChild(scriptEl);
    } else {
      const poll = setInterval(() => {
        if (window.Dos) {
          clearInterval(poll);
          initDos();
        }
      }, 100);
    }

    return () => {
      initializedRef.current = false;
    };
  }, []);

  return (
    <div
      ref={containerRef}
      style={{ width: '100%', height: '100%', background: '#000', display: 'block' }}
    />
  );
}
