import { beforeEach, describe, expect, it, vi } from 'vitest';
import { loadPOSSettings } from './loadPOSSettings';
import { useSettingsStore } from '../../stores/settingsStore';

const { getPOSSettings } = vi.hoisted(() => ({ getPOSSettings: vi.fn() }));
vi.mock('@voxire/api', () => ({ createOneKitClient: () => ({ getPOSSettings }) }));

beforeEach(() => {
  getPOSSettings.mockReset();
  useSettingsStore.getState().reset();
});

describe('terminal configuration loading', () => {
  it('does not apply a response after its business context is left', async () => {
    let resolve!: (value: { settings: { receiptFooter: string; businessProfile: string } }) => void;
    getPOSSettings.mockReturnValue(new Promise((done) => { resolve = done; }));
    let current = true;
    const loading = loadPOSSettings(() => current);
    current = false;
    useSettingsStore.getState().setReceiptFooter('Business B');
    resolve({ settings: { receiptFooter: 'Business A', businessProfile: 'restaurant' } });
    await loading;
    expect(useSettingsStore.getState().receiptFooter).toBe('Business B');
    expect(useSettingsStore.getState().businessProfile).toBe('retail');
  });

  it('rejects missing configuration and preserves current preferences', async () => {
    useSettingsStore.getState().setReceiptFooter('Saved footer');
    getPOSSettings.mockResolvedValue({});
    await expect(loadPOSSettings()).rejects.toThrow('Could not load terminal configuration');
    expect(useSettingsStore.getState().receiptFooter).toBe('Saved footer');
  });
});
