import { describe, it, expect } from 'vitest';
import {
  getStatusLabel,
  isTerminalStatus,
  sortSubmissions,
  parseAnswersJson,
  formatAnswer,
  getRequiredFields,
  parseOptionsJson,
} from './submission';
import type { Submission, FormField } from './types';

function makeSubmission(overrides: Partial<Submission> = {}): Submission {
  return {
    id: 1,
    careerId: 1,
    ownerId: 1,
    candidateName: 'Alice',
    email: 'alice@example.com',
    phone: '',
    resumeUrl: '',
    coverLetter: '',
    status: 'pending',
    answersJson: '',
    careerTitle: 'Engineer',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeField(overrides: Partial<FormField> = {}): FormField {
  return {
    id: 1,
    careerId: 1,
    ownerId: 1,
    kind: 'short_text',
    label: 'Portfolio URL',
    placeholder: '',
    required: false,
    optionsJson: '',
    position: 0,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('getStatusLabel', () => {
  it('returns the label for known statuses', () => {
    expect(getStatusLabel('pending')).toBe('Pending');
    expect(getStatusLabel('reviewing')).toBe('Reviewing');
    expect(getStatusLabel('shortlisted')).toBe('Shortlisted');
    expect(getStatusLabel('rejected')).toBe('Rejected');
    expect(getStatusLabel('hired')).toBe('Hired');
  });

  it('capitalizes unknown statuses gracefully', () => {
    expect(getStatusLabel('archived')).toBe('Archived');
    expect(getStatusLabel('new')).toBe('New');
  });
});

describe('isTerminalStatus', () => {
  it('returns true for rejected', () => {
    expect(isTerminalStatus('rejected')).toBe(true);
  });

  it('returns true for hired', () => {
    expect(isTerminalStatus('hired')).toBe(true);
  });

  it('returns false for non-terminal statuses', () => {
    expect(isTerminalStatus('pending')).toBe(false);
    expect(isTerminalStatus('reviewing')).toBe(false);
    expect(isTerminalStatus('shortlisted')).toBe(false);
  });
});

describe('sortSubmissions', () => {
  it('sorts pending before reviewing before terminal', () => {
    const list = [
      makeSubmission({ id: 3, status: 'hired', createdAt: '2026-03-01T00:00:00Z' }),
      makeSubmission({ id: 1, status: 'pending', createdAt: '2026-01-01T00:00:00Z' }),
      makeSubmission({ id: 2, status: 'reviewing', createdAt: '2026-02-01T00:00:00Z' }),
    ];
    const sorted = sortSubmissions(list);
    expect(sorted[0].status).toBe('pending');
    expect(sorted[1].status).toBe('reviewing');
    expect(sorted[2].status).toBe('hired');
  });

  it('sorts by createdAt descending within the same status', () => {
    const list = [
      makeSubmission({ id: 1, status: 'pending', createdAt: '2026-01-01T00:00:00Z' }),
      makeSubmission({ id: 2, status: 'pending', createdAt: '2026-03-01T00:00:00Z' }),
    ];
    const sorted = sortSubmissions(list);
    expect(sorted[0].id).toBe(2);
    expect(sorted[1].id).toBe(1);
  });

  it('does not mutate the original array', () => {
    const original = [makeSubmission({ id: 1, status: 'hired' })];
    sortSubmissions(original);
    expect(original[0].id).toBe(1);
  });
});

describe('parseAnswersJson', () => {
  it('returns an empty object for an empty string', () => {
    expect(parseAnswersJson('')).toEqual({});
  });

  it('parses a valid JSON object string', () => {
    const result = parseAnswersJson('{"q1":"yes","q2":"no"}');
    expect(result).toEqual({ q1: 'yes', q2: 'no' });
  });

  it('returns an empty object for invalid JSON', () => {
    expect(parseAnswersJson('{bad json}')).toEqual({});
  });

  it('returns an empty object for a JSON array (wrong shape)', () => {
    expect(parseAnswersJson('["a","b"]')).toEqual({});
  });
});

describe('formatAnswer', () => {
  it('returns "-" for null or undefined', () => {
    expect(formatAnswer(null, 'short_text')).toBe('-');
    expect(formatAnswer(undefined, 'short_text')).toBe('-');
  });

  it('returns "-" for empty string', () => {
    expect(formatAnswer('', 'short_text')).toBe('-');
  });

  it('formats boolean true as "Yes"', () => {
    expect(formatAnswer(true, 'boolean')).toBe('Yes');
    expect(formatAnswer('true', 'boolean')).toBe('Yes');
  });

  it('formats boolean false as "No"', () => {
    expect(formatAnswer(false, 'boolean')).toBe('No');
    expect(formatAnswer('false', 'boolean')).toBe('No');
  });

  it('joins multi-select arrays with commas', () => {
    expect(formatAnswer(['React', 'TypeScript'], 'multi_select')).toBe('React, TypeScript');
  });

  it('converts other values to string', () => {
    expect(formatAnswer(42, 'short_text')).toBe('42');
    expect(formatAnswer('hello', 'url')).toBe('hello');
  });
});

describe('getRequiredFields', () => {
  it('returns only required fields', () => {
    const fields = [
      makeField({ id: 1, required: false, position: 0 }),
      makeField({ id: 2, required: true, position: 1 }),
      makeField({ id: 3, required: true, position: 2 }),
    ];
    const result = getRequiredFields(fields);
    expect(result).toHaveLength(2);
    expect(result.every((f) => f.required)).toBe(true);
  });

  it('sorts required fields by position', () => {
    const fields = [
      makeField({ id: 1, required: true, position: 3 }),
      makeField({ id: 2, required: true, position: 1 }),
    ];
    const result = getRequiredFields(fields);
    expect(result[0].id).toBe(2);
    expect(result[1].id).toBe(1);
  });

  it('returns an empty array when no fields are required', () => {
    expect(getRequiredFields([makeField({ required: false })])).toHaveLength(0);
  });
});

describe('parseOptionsJson', () => {
  it('returns an empty array for an empty string', () => {
    expect(parseOptionsJson('')).toEqual([]);
  });

  it('parses a JSON array of strings', () => {
    expect(parseOptionsJson('["Option A","Option B"]')).toEqual(['Option A', 'Option B']);
  });

  it('coerces non-string array items to strings', () => {
    expect(parseOptionsJson('[1,2,3]')).toEqual(['1', '2', '3']);
  });

  it('returns an empty array for invalid JSON', () => {
    expect(parseOptionsJson('{not: an array}')).toEqual([]);
  });

  it('returns an empty array for a JSON object (wrong shape)', () => {
    expect(parseOptionsJson('{"a":1}')).toEqual([]);
  });
});
