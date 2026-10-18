import { describe, expect, it } from 'vitest';
import { apiErrorMessage } from './apiErrorMessage';

describe('apiErrorMessage', () => {
  it('uses the message from a JSON error body', () => {
    expect(apiErrorMessage({ body: '{"message":"Insufficient stock"}' }, 'Failed')).toBe(
      'Insufficient stock'
    );
  });

  it('falls back for transport errors and bodies without a message', () => {
    expect(apiErrorMessage(new TypeError('Failed to fetch'), 'Failed')).toBe('Failed');
    expect(apiErrorMessage({ body: '<html>502</html>' }, 'Failed')).toBe('Failed');
    expect(apiErrorMessage({ body: '{"message":"  "}' }, 'Failed')).toBe('Failed');
    expect(apiErrorMessage(undefined, 'Failed')).toBe('Failed');
  });
});
