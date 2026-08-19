import { describe, expect, it } from 'vitest';
import { routeNeedsRealtime } from './realtimeRouteGate';

describe('routeNeedsRealtime', () => {
  it('is false for the settings route and its sub-paths', () => {
    expect(routeNeedsRealtime('/settings')).toBe(false);
    expect(routeNeedsRealtime('/settings/time')).toBe(false);
  });

  it('is true for workspace and project routes', () => {
    expect(routeNeedsRealtime('/w/1')).toBe(true);
    expect(routeNeedsRealtime('/w/1/p/2')).toBe(true);
    expect(routeNeedsRealtime('/inbox')).toBe(true);
    expect(routeNeedsRealtime('/')).toBe(true);
  });

  it('is not fooled by a path that merely contains "settings" mid-segment', () => {
    expect(routeNeedsRealtime('/w/1/p/settings-project')).toBe(true);
  });
});
