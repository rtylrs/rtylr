import { describe, expect, it } from 'vitest';
import i18n from './i18next';

describe('Flow i18n startup', () => {
  it('loads app and shared strings before the routed app renders', async () => {
    if (!i18n.isInitialized) {
      await new Promise<void>((resolve) => i18n.on('initialized', () => resolve()));
    }

    expect(i18n.t('flow.shell.myTasks')).toBe('My tasks');
    expect(i18n.t('common.create')).toBe('Create');
  });
});
