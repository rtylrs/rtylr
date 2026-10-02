import { useEffect, useMemo, useState } from 'react';
import { Router, createPath, parsePath, type Navigator, type To } from 'react-router-dom';
import { useRouter, useRouterState } from '@tanstack/react-router';
import { Spinner } from '@voxire/ui';
import {
  POSApp,
  POSSettings,
  POSPortalContext,
  preparePortalTerminal,
  loadPOSSettings,
  useSettingsStore,
} from 'rtylr-pos/portal';
import { useConfigStore } from '../../shared/stores/configStore';
import { useUserStore } from '../../shared/stores/userStore';
function portalTarget(to: To) {
  const path = typeof to === 'string' ? parsePath(to) : to;
  return createPath({
    ...path,
    pathname:
      path.pathname === '/settings'
        ? '/settings/pos'
        : `/pos${!path.pathname || path.pathname === '/' ? '' : path.pathname}`,
  });
}
export default function PosModule({ settings = false }: { settings?: boolean }) {
  const router = useRouter();
  const location = useRouterState({ select: (state) => state.location });
  const businessId = useUserStore((state) => state.user?.businessId);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState(false);
  useEffect(() => {
    let active = true;
    setReady(false);
    setError(false);
    if (!businessId) return;
    void preparePortalTerminal(businessId)
      .then(async () => {
        if (!active) return;
        if (settings) {
          await loadPOSSettings(() => active);
          if (!active) return;
          const profile = useConfigStore.getState().websiteData;
          if (typeof profile?.siteName === 'string') {
            useSettingsStore.getState().setBusiness({
              name: profile.siteName,
              address: typeof profile.address === 'string' ? profile.address : '',
              phone: typeof profile.contactNumber === 'string' ? profile.contactNumber : '',
            });
          }
        }
        if (active) setReady(true);
      })
      .catch(() => {
        if (active) setError(true);
      });
    return () => {
      active = false;
    };
  }, [businessId, settings]);
  const navigator = useMemo<Navigator>(
    () => ({
      createHref: portalTarget,
      go: (delta) => router.history.go(delta),
      push: (to, state) => {
        void router.history.push(portalTarget(to), state);
      },
      replace: (to, state) => {
        void router.history.replace(portalTarget(to), state);
      },
    }),
    [router]
  );
  if (error) throw new Error('Could not restore terminal preferences');
  if (!ready)
    return (
      <div className="flex min-h-48 items-center justify-center">
        <Spinner />
      </div>
    );
  const pathname = settings ? '/settings' : location.pathname.slice(4) || '/';
  return (
    <POSPortalContext.Provider value={true}>
      <Router
        location={{ pathname, search: location.searchStr, hash: location.hash }}
        navigator={navigator}
      >
        {settings ? <POSSettings /> : <POSApp />}
      </Router>
    </POSPortalContext.Provider>
  );
}
