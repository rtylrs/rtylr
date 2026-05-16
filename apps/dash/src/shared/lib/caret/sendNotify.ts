/** Best-effort browser notification for completed Caret responses while
 *  the user is on a different tab. */
export function showCaretNotification(title: string, body: string): void {
  if (typeof Notification === 'undefined') return;
  if (Notification.permission !== 'granted') return;
  if (typeof document !== 'undefined' && document.visibilityState === 'visible') return;
  try {
    new Notification(title, { body, silent: true });
  } catch {
    // ignore
  }
}
