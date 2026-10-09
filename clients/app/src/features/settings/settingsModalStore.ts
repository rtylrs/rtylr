import { create } from 'zustand';
import { redirect, type ParsedLocation } from '@tanstack/react-router';

interface SettingsModalState {
  /** Settings location shown in the modal, or null when it is closed. */
  path: string | null;
  open: (path: string) => void;
  close: () => void;
}

export const useSettingsModalStore = create<SettingsModalState>((set) => ({
  path: null,
  open: (path) => set({ path }),
  close: () => set({ path: null }),
}));

const SETTINGS_PAGE = /^\/(settings|security|api-keys|webhooks)(\/|$)/;

export const isSettingsPath = (pathname: string) => SETTINGS_PAGE.test(pathname);

/**
 * beforeLoad for settings routes in the app router: open the settings modal on
 * that location and stay on the current page instead of navigating to it. A
 * direct visit (bookmark, reload) has no page behind it, so it lands on home.
 */
export function openInSettingsModal({
  location,
  preload,
}: {
  location: ParsedLocation;
  preload: boolean;
}): void {
  // Link preloading runs beforeLoad on hover; that must not open the modal.
  if (preload) return;
  useSettingsModalStore.getState().open(location.href);
  const { pathname, search, hash } = window.location;
  throw redirect({
    href: isSettingsPath(pathname) ? '/' : `${pathname}${search}${hash}`,
    replace: true,
  });
}
