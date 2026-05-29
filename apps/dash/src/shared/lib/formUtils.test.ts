import { describe, it, expect } from 'vitest';
import {
  hasNoErrors,
  errorCount,
  allTouched,
  touchAll,
  resetTouched,
  dirtyFields,
  isDirty,
  trimStringFields,
} from './formUtils';

describe('hasNoErrors', () => {
  it('returns true when all errors are null', () => {
    expect(hasNoErrors({ name: null, email: null })).toBe(true);
  });

  it('returns false when any error is set', () => {
    expect(hasNoErrors({ name: 'Required', email: null })).toBe(false);
  });

  it('returns true for an empty error object', () => {
    expect(hasNoErrors({})).toBe(true);
  });
});

describe('errorCount', () => {
  it('counts non-null errors', () => {
    expect(errorCount({ name: 'Required', email: null, age: 'Too young' })).toBe(2);
  });

  it('returns 0 when no errors', () => {
    expect(errorCount({ name: null })).toBe(0);
  });
});

describe('allTouched', () => {
  it('returns true when all specified fields are touched', () => {
    expect(allTouched({ name: true, email: true }, ['name', 'email'])).toBe(true);
  });

  it('returns false when some fields are not touched', () => {
    expect(allTouched({ name: true, email: false }, ['name', 'email'])).toBe(false);
  });

  it('returns true for an empty fields list', () => {
    expect(allTouched({}, [])).toBe(true);
  });
});

describe('touchAll', () => {
  it('marks all given fields as touched', () => {
    const result = touchAll({ name: false, email: false }, ['name', 'email']);
    expect(result.name).toBe(true);
    expect(result.email).toBe(true);
  });

  it('does not mutate the original object', () => {
    const original = { name: false };
    touchAll(original, ['name']);
    expect(original.name).toBe(false);
  });
});

describe('resetTouched', () => {
  it('sets all touched flags to false', () => {
    const result = resetTouched({ name: true, email: true });
    expect(result.name).toBe(false);
    expect(result.email).toBe(false);
  });
});

describe('dirtyFields', () => {
  it('returns fields that have changed', () => {
    const dirty = dirtyFields({ name: 'Alice', age: 31 }, { name: 'Alice', age: 30 });
    expect(dirty).toEqual({ age: 31 });
  });

  it('returns empty object when nothing changed', () => {
    expect(dirtyFields({ name: 'Alice' }, { name: 'Alice' })).toEqual({});
  });
});

describe('isDirty', () => {
  it('returns true when any field changed', () => {
    expect(isDirty({ name: 'Bob' }, { name: 'Alice' })).toBe(true);
  });

  it('returns false when nothing changed', () => {
    expect(isDirty({ name: 'Alice' }, { name: 'Alice' })).toBe(false);
  });
});

describe('trimStringFields', () => {
  it('trims whitespace from string fields', () => {
    const result = trimStringFields({ name: '  Alice  ', age: 30 });
    expect(result.name).toBe('Alice');
    expect(result.age).toBe(30);
  });

  it('does not mutate the original object', () => {
    const original = { name: '  Bob  ' };
    trimStringFields(original);
    expect(original.name).toBe('  Bob  ');
  });
});
