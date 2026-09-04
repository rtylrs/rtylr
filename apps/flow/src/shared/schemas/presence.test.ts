import { describe, expect, it } from 'vitest';
import { presenceResponseSchema } from './presence';

/**
 * The endpoint answers snake_case, but `@voxire/api`'s response interceptor
 * runs `convertKeys`/`snakeToCamel` over every response first, so what reaches
 * this schema is camelCase. These fixtures are the post-interceptor shape,
 * which is the only shape `usePresence` can ever hand it.
 */
const wireEntry = {
  employee_id: 41,
  user_id: '9319',
  first_name: 'Jane',
  last_name: 'Doe',
  status: 'onsite',
};

/** What convertKeys does to the above, spelled out rather than imported. */
const receivedEntry = {
  employeeId: 41,
  userId: '9319',
  firstName: 'Jane',
  lastName: 'Doe',
  status: 'onsite',
};

describe('presenceResponseSchema (#1067)', () => {
  it('parses the camelCased payload the api client delivers', () => {
    const parsed = presenceResponseSchema.parse({
      presence: [receivedEntry],
      date: '2026-09-04',
    });

    expect(parsed.presence).toEqual([
      { employeeId: '41', userId: '9319', firstName: 'Jane', lastName: 'Doe', status: 'onsite' },
    ]);
  });

  it('coerces a numeric id to a string, whichever way it arrives', () => {
    const parsed = presenceResponseSchema.parse({
      presence: [{ ...receivedEntry, employeeId: 41, userId: 9319 }],
      date: '2026-09-04',
    });

    expect(parsed.presence[0].employeeId).toBe('41');
    expect(parsed.presence[0].userId).toBe('9319');
  });

  it('accepts an entry with no name, which the directory allows', () => {
    const parsed = presenceResponseSchema.parse({
      presence: [{ employeeId: '1', userId: '2', status: 'not_started' }],
      date: '2026-09-04',
    });

    expect(parsed.presence[0].firstName).toBeUndefined();
  });

  it('rejects the raw snake_case shape, which this client never receives', () => {
    // Parsing for snake_case was the defect: both id unions failed on
    // undefined, the whole parse threw, and /presence fell back to its error
    // state on a 200 with a full payload. Accepting both shapes would hide the
    // same mistake next time, so the wire shape is deliberately invalid here.
    expect(() =>
      presenceResponseSchema.parse({ presence: [wireEntry], date: '2026-09-04' })
    ).toThrow();
  });

  it('rejects a response with no presence array', () => {
    expect(() => presenceResponseSchema.parse({ date: '2026-09-04' })).toThrow();
  });
});
