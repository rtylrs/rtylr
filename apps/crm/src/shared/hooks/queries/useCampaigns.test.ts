import { describe, expect, it } from 'vitest';
import { toCampaignRequest } from './useCampaigns';

describe('campaign wire mappings', () => {
  it('maps the UI campaign model to the live CRM campaign request', () => {
    expect(
      toCampaignRequest({
        name: 'Weekend sale',
        type: 'sms',
        body: 'Save today',
        audienceSegmentId: 4,
      })
    ).toMatchObject({
      name: 'Weekend sale',
      type: 'sms',
      audienceSegmentId: 4,
      subject: 'Weekend sale',
      body: 'Save today',
    });
  });

  it('does not invent a schedule for drafts and strips milliseconds for the backend parser', () => {
    expect(toCampaignRequest({ name: 'Draft', type: 'email', body: 'Content' }).scheduledAt).toBe(
      undefined
    );
    expect(
      toCampaignRequest({
        name: 'Scheduled sale',
        type: 'email',
        body: 'Content',
        scheduledAt: '2026-09-01T10:00:00.123Z',
      }).scheduledAt
    ).toBe('2026-09-01T10:00:00Z');
  });

  it('preserves an existing schedule when updating a campaign', () => {
    expect(
      toCampaignRequest({
        name: 'Scheduled sale',
        type: 'email',
        subject: 'A subject',
        body: 'Content',
        scheduledAt: '2026-09-01T10:00:00Z',
      }).scheduledAt
    ).toBe('2026-09-01T10:00:00Z');
  });

  it('keeps valid date-only schedules on the same calendar day', () => {
    expect(
      toCampaignRequest({
        name: 'Date-only schedule',
        type: 'email',
        body: 'Content',
        scheduledAt: '2026-09-01',
      }).scheduledAt
    ).toBe('2026-09-01T00:00:00Z');
  });

  it('drops impossible calendar dates instead of normalizing them', () => {
    expect(
      toCampaignRequest({
        name: 'Invalid schedule',
        type: 'email',
        body: 'Content',
        scheduledAt: '2026-02-30',
      }).scheduledAt
    ).toBeUndefined();
  });
});
