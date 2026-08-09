import { describe, it, expect } from 'vitest';
import { getLayoutTypeForRoute } from './layoutConfig';

describe('getLayoutTypeForRoute', () => {
  it('returns custom-sidebar for /settings', () => {
    expect(getLayoutTypeForRoute('/settings')).toBe('custom-sidebar');
  });
  it('returns custom-sidebar for nested settings paths', () => {
    expect(getLayoutTypeForRoute('/settings/account')).toBe('custom-sidebar');
  });
  it('returns default for dashboard routes', () => {
    expect(getLayoutTypeForRoute('/app/business/dashboard')).toBe('default');
  });
  it('returns default for root path', () => {
    expect(getLayoutTypeForRoute('/')).toBe('default');
  });
});
