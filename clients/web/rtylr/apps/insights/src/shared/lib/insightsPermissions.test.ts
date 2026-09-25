import { describe, expect, it } from 'vitest';
import { canInsightsPermission } from './insightsPermissions';

describe('canInsightsPermission', () => {
  it('allows admin and owner to export reports', () => {
    expect(canInsightsPermission('admin', 'insights:reports:export')).toBe(true);
    expect(canInsightsPermission('owner', 'insights:reports:export')).toBe(true);
  });

  it('denies every non-admin-tier role', () => {
    expect(canInsightsPermission('viewer', 'insights:reports:export')).toBe(false);
    expect(canInsightsPermission('staff', 'insights:reports:export')).toBe(false);
    expect(canInsightsPermission('operator', 'insights:reports:export')).toBe(false);
    expect(canInsightsPermission('billing', 'insights:reports:export')).toBe(false);
    expect(canInsightsPermission('menu_manager', 'insights:reports:export')).toBe(false);
    expect(canInsightsPermission('store_manager', 'insights:reports:export')).toBe(false);
  });

  it('denies a missing role', () => {
    expect(canInsightsPermission(null, 'insights:reports:export')).toBe(false);
    expect(canInsightsPermission(undefined, 'insights:reports:export')).toBe(false);
  });

  it('lets operators view reports but not export them', () => {
    expect(canInsightsPermission('operator', 'insights:reports')).toBe(true);
    expect(canInsightsPermission('operator', 'insights:reports:export')).toBe(false);
  });
});
