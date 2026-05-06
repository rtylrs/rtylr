import type React from 'react';
import { Sentry } from '@voxire/observability';

type CaptureException = (
  error: Error,
  context: {
    contexts: {
      react: {
        componentStack?: string | null;
      };
    };
    tags: {
      surface: 'react-error-boundary';
    };
  }
) => unknown;

export function reportReactError(
  error: Error,
  info: React.ErrorInfo,
  captureException: CaptureException = Sentry.captureException
): void {
  captureException(error, {
    contexts: {
      react: {
        componentStack: info.componentStack,
      },
    },
    tags: {
      surface: 'react-error-boundary',
    },
  });
}
