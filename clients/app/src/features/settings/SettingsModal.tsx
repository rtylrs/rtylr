import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  createMemoryHistory,
  createRootRouteWithContext,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
  useNavigate,
  useRouteContext,
  useRouterState,
} from '@tanstack/react-router';
import { ErrorBoundary } from '@voxire/ui';
import { Dialog, DialogContent, DialogTitle } from '../../shared/components/ui';
import { useUserStore } from '../../shared/stores/userStore';
import { useAddons, isAddonEnabled } from '../../shared/hooks/queries/useAddons';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import SettingsNavigation from '../../shell/SettingsNavigation';
import { addonSettingsItems, settingsNavItems } from '../../shell/navigation';
import { createSettingsRoutes } from './settingsRoutes';
import { useSettingsModalStore } from './settingsModalStore';
import './settingsModal.css';

interface SettingsModalContext {
  /** Leaves the modal for a location outside settings, in the app router. */
  leave: (href: string) => void;
}

function SettingsModalLayout() {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const { data: addons = [] } = useAddons();
  const [query, setQuery] = useState('');
  const items = [
    ...settingsNavItems,
    ...addonSettingsItems.filter((item) => isAddonEnabled(addons, item.addonId)),
  ].filter((item) => canAccessPortalPath(user, item.path));

  return (
    <div className="settings-modal">
      <nav className="settings-modal-nav" aria-label={t('common.nav.settings')}>
        <p className="settings-modal-title" aria-hidden>
          {t('common.nav.settings')}
        </p>
        <SettingsNavigation
          items={items}
          query={query}
          onQueryChange={setQuery}
          onNavigate={() => undefined}
        />
      </nav>
      <div className="settings-modal-content">
        <ErrorBoundary>
          <Outlet />
        </ErrorBoundary>
      </div>
    </div>
  );
}

// Anything outside settings (a link to /team, a redirect to /) belongs to the
// app router: hand it over and close the modal.
function LeaveSettings() {
  const { leave } = useRouteContext({ strict: false }) as SettingsModalContext;
  const href = useRouterState({ select: (state) => state.location.href });
  useEffect(() => leave(href), [leave, href]);
  return null;
}

const modalRoot = createRootRouteWithContext<SettingsModalContext>()({
  component: SettingsModalLayout,
  notFoundComponent: LeaveSettings,
});
// Mirrors the app's pathless `business` layout so route ids match the app router
// (e.g. FlowSettings reads search from '/business/settings/flow').
const modalBusinessRoute = createRoute({
  getParentRoute: () => modalRoot,
  id: 'business',
  component: Outlet,
});
export const modalRouteTree = modalRoot.addChildren([
  modalBusinessRoute.addChildren([...createSettingsRoutes(modalBusinessRoute)]),
]);

/**
 * Settings as a modal over the current page. Its own router runs on memory
 * history, so every settings section, its sub-pages and links between them work
 * unchanged without touching the page behind it.
 */
export default function SettingsModal() {
  const { t } = useTranslation();
  const path = useSettingsModalStore((state) => state.path);
  const close = useSettingsModalStore((state) => state.close);
  const navigate = useNavigate();

  const router = useMemo(() => {
    if (!path) return null;
    return createRouter({
      routeTree: modalRouteTree,
      history: createMemoryHistory({ initialEntries: [path] }),
      context: {
        leave: (href: string) => {
          close();
          void navigate({ href });
        },
      },
      defaultNotFoundComponent: LeaveSettings,
    });
  }, [path, close, navigate]);

  return (
    <Dialog open={router !== null} onOpenChange={(open) => !open && close()}>
      {router && (
        <DialogContent className="settings-modal-dialog" aria-describedby={undefined}>
          {/* Outside the modal router so the dialog is named as soon as it opens. */}
          <DialogTitle className="sr-only">{t('common.nav.settings')}</DialogTitle>
          {/* The modal router is separate from the app's registered router type. */}
          <RouterProvider
            router={router as unknown as React.ComponentProps<typeof RouterProvider>['router']}
          />
        </DialogContent>
      )}
    </Dialog>
  );
}
