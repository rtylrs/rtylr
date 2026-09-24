import React from 'react';

/**
 * Whether the OS currently asks for a dark scheme.
 *
 * The Appearance panel needs this for two things the provider cannot answer for
 * it: which palette family to offer while the mode is `system`, and what the
 * preview pane should show when a mode card is being auditioned. Both are read
 * on the panel's own render path, so the panel reads the media query itself
 * rather than asking the provider what it resolved to.
 *
 * The subscription is unconditional here, unlike the provider's, which gates it
 * on the mode being `system`. The provider gates because a scheme change must
 * not re-write tokens for a user who picked light or dark explicitly. Nothing
 * here writes anything, so there is no such hazard, and an unconditional
 * subscription keeps the `system` card's two miniatures correct at all times.
 */
const DARK_QUERY = '(prefers-color-scheme: dark)';

function matchesDark(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false;
  return window.matchMedia(DARK_QUERY).matches;
}

export function useSystemDark(): boolean {
  const [dark, setDark] = React.useState<boolean>(matchesDark);

  React.useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return;
    const query = window.matchMedia(DARK_QUERY);
    setDark(query.matches);

    const onChange = (event: MediaQueryListEvent): void => setDark(event.matches);
    if (typeof query.addEventListener === 'function') {
      query.addEventListener('change', onChange);
      return () => {
        if (typeof query.removeEventListener === 'function') {
          query.removeEventListener('change', onChange);
        }
      };
    }

    // Safari 13 and older expose only the legacy listener API.
    if (typeof query.addListener === 'function') {
      query.addListener(onChange);
      return () => {
        if (typeof query.removeListener === 'function') {
          query.removeListener(onChange);
        }
      };
    }
  }, []);

  return dark;
}
