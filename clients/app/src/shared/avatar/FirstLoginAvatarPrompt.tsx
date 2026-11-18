import React, { useEffect, useState } from 'react';
import { useLocation } from '@tanstack/react-router';
import { useUserStore } from '../stores/userStore';
import { AvatarPicker } from './AvatarPicker';

/** How long after the portal settles before the welcome opens, so it never races a page's own dialogs. */
const DELAY_MS = 1500;

const dismissedKey = (userId: number): string => `rtylr:avatar-welcome-seen:${userId}`;

function seen(userId: number): boolean {
  try {
    return localStorage.getItem(dismissedKey(userId)) === '1';
  } catch {
    // No storage (private mode): showing it once per session is still fine.
    return false;
  }
}

function markSeen(userId: number): void {
  try {
    localStorage.setItem(dismissedKey(userId), '1');
  } catch {
    // Ignored, see above.
  }
}

/** Screens where a welcome dialog would be in the way. */
const QUIET_PATHS = ['/onboarding', '/verify', '/login', '/signup', '/invite', '/menu/'];

/**
 * A one-time "pick your avatar" welcome for someone who has never chosen one.
 * Skippable; picking, skipping or closing all count, and it never shows again
 * for that user on this browser.
 */
export const FirstLoginAvatarPrompt: React.FC = () => {
  const user = useUserStore((state) => state.user);
  const { pathname } = useLocation();
  const [open, setOpen] = useState(false);

  const userId = user?.id;
  const eligible =
    userId !== undefined &&
    !!user?.businessId &&
    user.onboarded !== false &&
    !user.avatar &&
    !QUIET_PATHS.some((p) => pathname.startsWith(p));

  useEffect(() => {
    if (!eligible || userId === undefined || seen(userId)) return undefined;
    const timer = window.setTimeout(() => {
      // Seen the moment it opens: a reload mid-choice must not bring it back.
      markSeen(userId);
      setOpen(true);
    }, DELAY_MS);
    return () => window.clearTimeout(timer);
  }, [eligible, userId]);

  if (!open || userId === undefined) return null;
  return (
    <AvatarPicker
      open={open}
      onOpenChange={setOpen}
      userId={String(userId)}
      name={user?.fullName}
      current={user?.avatar}
      welcome
    />
  );
};

export default FirstLoginAvatarPrompt;
