import { describe, expect, it } from 'vitest';
import { activeLeadFilterGroups, clearLeadFilterGroup, DEFAULT_LEAD_FILTERS } from './leadFilters';
import { leadFilterParams } from './queries';

describe('lead filters', () => {
  it('counts each popover group once, ranges included', () => {
    expect(activeLeadFilterGroups(DEFAULT_LEAD_FILTERS)).toEqual([]);
    const filters = {
      ...DEFAULT_LEAD_FILTERS,
      search: 'cedar',
      hasPhone: 'yes' as const,
      industries: ['Retail'],
      attemptsMin: '1',
      attemptsMax: '3',
      createdTo: '2026-10-01',
    };
    expect(activeLeadFilterGroups(filters)).toEqual([
      'industries',
      'hasPhone',
      'attempts',
      'created',
    ]);
    expect(clearLeadFilterGroup('attempts')).toEqual({ attemptsMin: '', attemptsMax: '' });
    expect(clearLeadFilterGroup('importId')).toEqual({ importId: undefined });
  });

  it('sends text sets as JSON and leaves out half-typed ranges', () => {
    const params = leadFilterParams({
      ...DEFAULT_LEAD_FILTERS,
      reps: ['0', '7'],
      industries: ['Food, Beverage'],
      billing: ['monthly', 'yearly'],
      hasEmail: 'no',
      attemptsMin: '2',
      attemptsMax: 'x',
      valueMin: '10.5',
      valueMax: '1.234',
    });
    expect(params).toMatchObject({
      reps: '0,7',
      industries: '["Food, Beverage"]',
      billing: 'monthly,yearly',
      hasEmail: 'no',
      attemptsMin: 2,
      attemptsMax: undefined,
      valueMin: '10.5',
      valueMax: undefined,
      sources: undefined,
      hasPhone: undefined,
    });
  });
});
