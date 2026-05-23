/**
 * Manage the browser Notifications permission for Caret push notifications.
 * Surface-level helpers - actual notification triggering is up to callers.
 */

export type NotifyState = 'granted' | 'denied' | 'default' | 'unsupported';

export function notifyState(): NotifyState {
  if (typeof Notification === 'undefined') return 'unsupported';
  return Notification.permission;
}

export async function requestNotifyPermission(): Promise<NotifyState> {
  if (typeof Notification === 'undefined') return 'unsupported';
  if (Notification.permission !== 'default') return Notification.permission as NotifyState;
  try {
    const result = await Notification.requestPermission();
    return result as NotifyState;
  } catch {
    return 'denied';
  }
}
