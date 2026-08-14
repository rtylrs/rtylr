import { describe, expect, it } from 'vitest';
import { camelizeRecruitPayload, snakeizeRecruitPayload } from './api';

describe('recruit API wire conversion', () => {
  it('decodes nested backend snake_case payloads', () => {
    expect(
      camelizeRecruitPayload({
        career_id: 7,
        scheduled_at: '2026-08-14T09:00:00Z',
        fields: [{ options_json: '["Remote"]' }],
      })
    ).toEqual({
      careerId: 7,
      scheduledAt: '2026-08-14T09:00:00Z',
      fields: [{ optionsJson: '["Remote"]' }],
    });
  });

  it('encodes nested frontend mutation payloads', () => {
    expect(
      snakeizeRecruitPayload({
        careerSlug: 'backend-engineer',
        candidateName: 'Jane Doe',
        answersJson: '{"1":"Remote"}',
      })
    ).toEqual({
      career_slug: 'backend-engineer',
      candidate_name: 'Jane Doe',
      answers_json: '{"1":"Remote"}',
    });
  });
});
