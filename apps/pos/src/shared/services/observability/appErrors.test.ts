import { describe, expect, it, vi } from 'vitest';
import { reportReactError } from './appErrors';

describe('reportReactError', () => {
  it('captures React boundary errors with component stack metadata', () => {
    const captureException = vi.fn();
    const error = new Error('render failed');

    reportReactError(
      error,
      {
        componentStack: '\n    at CheckoutScreen',
      },
      captureException
    );

    expect(captureException).toHaveBeenCalledWith(error, {
      contexts: {
        react: {
          componentStack: '\n    at CheckoutScreen',
        },
      },
      tags: {
        surface: 'react-error-boundary',
      },
    });
  });
});
