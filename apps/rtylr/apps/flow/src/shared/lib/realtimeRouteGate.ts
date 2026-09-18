/**
 * True when the current route needs the realtime connection.
 *
 * A single predicate drives two things in `FlowLayout`: whether
 * `useLayoutRealtime` subscribes to any topics, and whether the
 * `RealtimeConnectionIndicator` is shown at all. They must stay in lock step -
 * Settings intentionally never subscribes (none of the sidebar/task data it
 * depends on is on screen there, see `useLayoutRealtime`), so on a route where
 * this is `false` the socket may never have been asked to connect. Before this
 * predicate also gated the indicator, that showed as a persistent
 * "Disconnected. Reload to see the latest." status on Settings screens - true
 * in the narrow sense that no connection was open, but misleading, since
 * nothing on screen was actually stale and reloading while still on Settings
 * would not change it. See issue #761.
 */
export function routeNeedsRealtime(pathname: string): boolean {
  return !pathname.startsWith('/settings');
}
