import { describe, it, expect } from 'vitest';
import { crmKeys } from './queryKeys';

describe('crmKeys', () => {
  describe('customers', () => {
    it('returns a stable base key', () => {
      expect(crmKeys.customers.all()).toEqual(['crm', 'customers']);
    });

    it('returns a list key without filters', () => {
      expect(crmKeys.customers.list()).toEqual(['crm', 'customers', 'list', undefined]);
    });

    it('returns a list key with filters', () => {
      const key = crmKeys.customers.list({ tier: 'Gold' });
      expect(key[3]).toEqual({ tier: 'Gold' });
    });

    it('returns a detail key with the customer id', () => {
      expect(crmKeys.customers.detail(42)).toEqual(['crm', 'customers', 'detail', 42]);
    });

    it('returns a timeline key', () => {
      expect(crmKeys.customers.timeline(7)).toEqual(['crm', 'customers', 'timeline', 7]);
    });
  });

  describe('campaigns', () => {
    it('returns stable base key', () => {
      expect(crmKeys.campaigns.all()).toEqual(['crm', 'campaigns']);
    });

    it('returns stats key for a campaign id', () => {
      expect(crmKeys.campaigns.stats(5)).toEqual(['crm', 'campaigns', 'stats', 5]);
    });
  });

  describe('loyalty', () => {
    it('returns stable base key', () => {
      expect(crmKeys.loyalty.all()).toEqual(['crm', 'loyalty']);
    });

    it('returns distinct keys for config, members, rewards, tiers, transactions', () => {
      const keys = [
        crmKeys.loyalty.config(),
        crmKeys.loyalty.members(),
        crmKeys.loyalty.rewards(),
        crmKeys.loyalty.tiers(),
        crmKeys.loyalty.transactions(),
      ];
      const unique = new Set(keys.map((k) => JSON.stringify(k)));
      expect(unique.size).toBe(5);
    });
  });

  describe('segments', () => {
    it('returns a detail key with the segment id', () => {
      expect(crmKeys.segments.detail(3)).toEqual(['crm', 'segments', 'detail', 3]);
    });
  });

  describe('automations', () => {
    it('returns a detail key with the automation id', () => {
      expect(crmKeys.automations.detail(10)).toEqual(['crm', 'automations', 'detail', 10]);
    });
  });

  describe('dashboard', () => {
    it('returns a stats key', () => {
      expect(crmKeys.dashboard.stats()).toEqual(['crm', 'dashboard', 'stats']);
    });
  });

  describe('settings', () => {
    it('returns a base settings key', () => {
      expect(crmKeys.settings.all()).toEqual(['crm', 'settings']);
    });
  });
});
