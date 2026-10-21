import { useCallback, useEffect, useState } from 'react';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

const MANIFEST = '/caret/manifest.webmanifest';

/** Add a head element for the Caret route and return a cleanup. */
function addHeadElement(tag: 'link' | 'meta', attributes: Record<string, string>) {
  const element = document.createElement(tag);
  for (const [name, value] of Object.entries(attributes)) element.setAttribute(name, value);
  element.dataset.caretInstall = 'true';
  document.head.appendChild(element);
  return () => element.remove();
}

/**
 * Makes /caret installable as its own "Caret" app (desktop and Android via
 * the manifest, iOS via Add to Home Screen) and exposes the browser's install
 * prompt when one is available.
 */
export function useCaretInstall() {
  const [prompt, setPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [installed, setInstalled] = useState(
    () => window.matchMedia?.('(display-mode: standalone)').matches ?? false
  );

  useEffect(() => {
    // The portal's own icon stays for every other route.
    const portalTouchIcon = document.querySelector<HTMLLinkElement>('link[rel="apple-touch-icon"]');
    const previousTouchIcon = portalTouchIcon?.getAttribute('href') ?? null;
    portalTouchIcon?.setAttribute('href', '/caret/apple-touch-icon.png');

    const cleanups = [
      addHeadElement('link', { rel: 'manifest', href: MANIFEST }),
      addHeadElement('meta', { name: 'theme-color', content: '#111214' }),
      addHeadElement('meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }),
      addHeadElement('meta', { name: 'mobile-web-app-capable', content: 'yes' }),
      addHeadElement('meta', { name: 'apple-mobile-web-app-title', content: 'Caret' }),
      addHeadElement('meta', {
        name: 'apple-mobile-web-app-status-bar-style',
        content: 'black-translucent',
      }),
    ];

    const onPrompt = (event: Event) => {
      event.preventDefault();
      setPrompt(event as BeforeInstallPromptEvent);
    };
    const onInstalled = () => {
      setPrompt(null);
      setInstalled(true);
    };
    window.addEventListener('beforeinstallprompt', onPrompt);
    window.addEventListener('appinstalled', onInstalled);
    return () => {
      cleanups.forEach((cleanup) => cleanup());
      if (portalTouchIcon && previousTouchIcon !== null)
        portalTouchIcon.setAttribute('href', previousTouchIcon);
      window.removeEventListener('beforeinstallprompt', onPrompt);
      window.removeEventListener('appinstalled', onInstalled);
    };
  }, []);

  const install = useCallback(async () => {
    if (!prompt) return;
    await prompt.prompt();
    const { outcome } = await prompt.userChoice;
    setPrompt(null);
    if (outcome === 'accepted') setInstalled(true);
  }, [prompt]);

  return { canInstall: prompt !== null && !installed, installed, install };
}
