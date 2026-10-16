import { describe, expect, it } from 'vitest';
import {
  actionItemsToTopics,
  buildAnniversaries,
  mapOneOnOne,
  mapReview,
  mapTrainingCourse,
  reviewStatusFromApi,
  reviewStatusToApi,
  topicsToActionItems,
} from './hrPeopleMappers';

const names = new Map<number, string>([
  [7, 'Ada Lovelace'],
  [9, 'Grace Hopper'],
]);

describe('reviews', () => {
  it('maps the API review into the UI shape with staff names', () => {
    const review = mapReview(
      {
        id: '3',
        staffId: '7',
        reviewerId: '9',
        period: 'Q3 2026',
        rating: 4.5,
        goalsRating: 4,
        valuesRating: 5,
        skillsRating: 4,
        notes: 'Great',
        status: 'submitted',
        createdAt: '2026-10-01T10:00:00Z',
      },
      names
    );
    expect(review).toMatchObject({
      id: 3,
      staffId: 7,
      staffName: 'Ada Lovelace',
      reviewer: 'Grace Hopper',
      overall: 4.5,
      status: 'Submitted',
      date: '2026-10-01',
    });
  });

  it('round-trips statuses through the enum the table accepts', () => {
    expect(reviewStatusToApi('Submitted')).toBe('submitted');
    expect(reviewStatusToApi('Completed')).toBe('acknowledged');
    expect(reviewStatusToApi('In Progress')).toBe('draft');
    expect(reviewStatusFromApi('acknowledged')).toBe('Completed');
    expect(reviewStatusFromApi(undefined)).toBe('Draft');
  });
});

describe('one-on-ones', () => {
  it('stores topics as a JSON array, which is what the JSON column requires', () => {
    expect(topicsToActionItems(['a', 'b'])).toBe('["a","b"]');
    expect(actionItemsToTopics('["a","b"]')).toEqual(['a', 'b']);
  });

  it('still reads plain-text legacy action items', () => {
    expect(actionItemsToTopics('first\nsecond')).toEqual(['first', 'second']);
    expect(actionItemsToTopics('')).toEqual([]);
  });

  it('maps manager and report names and the status', () => {
    const item = mapOneOnOne(
      {
        id: '2',
        managerId: '9',
        staffId: '7',
        scheduledAt: '2026-10-20T09:00:00Z',
        status: 'cancelled',
        actionItems: '["roadmap"]',
      },
      names
    );
    expect(item).toMatchObject({
      manager: 'Grace Hopper',
      report: 'Ada Lovelace',
      status: 'Cancelled',
      topics: ['roadmap'],
    });
  });
});

describe('training courses', () => {
  it('converts hours to minutes and picks a known category from the description', () => {
    expect(
      mapTrainingCourse({ id: '1', name: 'Fire safety', description: 'safety', durationHours: 1.5 })
    ).toMatchObject({ title: 'Fire safety', category: 'Safety', durationMin: 90, required: false });
  });

  it('falls back to Skills for free-text descriptions', () => {
    expect(mapTrainingCourse({ id: '2', name: 'x', description: 'Learn things' }).category).toBe(
      'Skills'
    );
  });
});

describe('buildAnniversaries', () => {
  const today = new Date(2026, 9, 16);

  it('lists upcoming anniversaries with the milestone they complete', () => {
    const result = buildAnniversaries(
      [
        { id: '1', firstName: 'Ada', lastName: 'L', startDate: '2021-10-20' },
        { id: '2', firstName: 'Far', lastName: 'Away', startDate: '2021-03-01' },
      ],
      today
    );
    expect(result).toEqual([{ id: 1, staffId: 1, name: 'Ada L', date: '10-20', years: 5 }]);
  });

  it('rolls past dates to next year and counts today as upcoming', () => {
    const result = buildAnniversaries(
      [
        { id: '1', firstName: 'Today', startDate: '2020-10-16' },
        { id: '2', firstName: 'Soon', startDate: '2024-12-01' },
      ],
      today
    );
    expect(result.map((a) => [a.name, a.date, a.years])).toEqual([
      ['Today', '10-16', 6],
      ['Soon', '12-01', 2],
    ]);
  });

  it('skips this year’s hires and inactive staff', () => {
    const result = buildAnniversaries(
      [
        { id: '1', firstName: 'New', startDate: '2026-10-20' },
        { id: '2', firstName: 'Gone', status: 'terminated', startDate: '2020-10-20' },
        { id: '3', firstName: 'NoDate' },
      ],
      today
    );
    expect(result).toEqual([]);
  });
});
