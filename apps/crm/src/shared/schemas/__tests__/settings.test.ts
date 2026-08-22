import { describe, expect, it } from 'vitest';
import { normalizeSettings } from '../settings';

describe('normalizeSettings', () => {
  it('trims persisted text values while preserving defaults', () => {
    expect(
      normalizeSettings({
        timezone: ' America/Los_Angeles ',
        fromEmail: ' owner@example.com ',
        fromName: ' Rtylr ',
        smsSenderId: ' Rtylr ',
      })
    ).toEqual({
      defaultConsent: false,
      timezone: 'America/Los_Angeles',
      fromEmail: 'owner@example.com',
      fromName: 'Rtylr',
      smsSenderId: 'Rtylr',
    });
  });
});
