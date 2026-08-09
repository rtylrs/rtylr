import { describe, it, expect } from 'vitest';
import { getErrorMessage, isError, createError, safely, isUnauthorizedError } from '../error';

describe('getErrorMessage', () => {
  it('extracts message from Error instance', () => {
    expect(getErrorMessage(new Error('oops'))).toBe('oops');
  });

  it('returns string errors as-is', () => {
    expect(getErrorMessage('something went wrong')).toBe('something went wrong');
  });

  it('extracts message property from plain objects', () => {
    expect(getErrorMessage({ message: 'bad request' })).toBe('bad request');
  });

  it('returns fallback for unknown types', () => {
    expect(getErrorMessage(42)).toBe('An unexpected error occurred');
  });

  it('returns fallback for null', () => {
    expect(getErrorMessage(null)).toBe('An unexpected error occurred');
  });
});

describe('isError', () => {
  it('returns true for Error instances', () => {
    expect(isError(new Error('x'))).toBe(true);
  });

  it('returns false for plain objects', () => {
    expect(isError({ message: 'x' })).toBe(false);
  });
});

describe('createError', () => {
  it('creates an Error with the given message', () => {
    const err = createError('test error');
    expect(err).toBeInstanceOf(Error);
    expect(err.message).toBe('test error');
  });

  it('attaches a cause when provided', () => {
    const cause = new Error('root cause');
    const err = createError('wrapper', cause) as Error & { cause: unknown };
    expect(err.cause).toBe(cause);
  });
});

describe('safely', () => {
  it('returns the function result when it succeeds', () => {
    expect(safely(() => 42, 0)).toBe(42);
  });

  it('returns the fallback when the function throws', () => {
    expect(
      safely(() => {
        throw new Error('fail');
      }, 'default')
    ).toBe('default');
  });
});

describe('isUnauthorizedError', () => {
  it('returns true for Axios-style 401 errors', () => {
    expect(isUnauthorizedError({ response: { status: 401 } })).toBe(true);
  });

  it('returns false for other status codes', () => {
    expect(isUnauthorizedError({ response: { status: 403 } })).toBe(false);
  });

  it('returns true for flat status 401', () => {
    expect(isUnauthorizedError({ status: 401 })).toBe(true);
  });

  it('returns false for non-objects', () => {
    expect(isUnauthorizedError('error')).toBe(false);
  });
});
