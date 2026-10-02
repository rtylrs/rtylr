import { describe, expect, it } from 'vitest';
import { useUserStore } from '../../../../shared/stores/userStore';
import { useSettings } from './settings';

describe('Finance settings business isolation', () => {
  it('restores each business preferences without copying them into another business', () => {
    useUserStore.getState().setUser({ businessId: 901 });
    useSettings.getState().setSetting('companyName', 'Company A');
    useSettings.getState().setSetting('currency', 'EUR');
    useUserStore.getState().setUser({ businessId: 902 });
    expect(useSettings.getState().companyName).toBe('');
    expect(useSettings.getState().currency).toBe('USD');
    useSettings.getState().setSetting('companyName', 'Company B');
    useUserStore.getState().setUser({ businessId: 901 });
    expect(useSettings.getState().companyName).toBe('Company A');
    expect(useSettings.getState().currency).toBe('EUR');
    useUserStore.getState().setUser({ businessId: 902 });
    expect(useSettings.getState().companyName).toBe('Company B');
    useUserStore.getState().logout();
    expect(useSettings.getState().companyName).toBe('');
  });
});
