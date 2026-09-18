import { describe, expect, it } from 'vitest';
import { isDynamicImportError } from './ErrorBoundary';

describe('isDynamicImportError', () => {
  it('detects Vite dynamic import fetch failures', () => {
    expect(
      isDynamicImportError(
        new Error(
          'Failed to fetch dynamically imported module: https://dash.rtylr.com/assets/Account-Bu2-9eGC.js'
        )
      )
    ).toBe(true);
  });

  it('detects common chunk load failures', () => {
    expect(isDynamicImportError(new Error('Loading chunk 42 failed.'))).toBe(true);
    expect(isDynamicImportError(new Error('ChunkLoadError: Loading chunk settings failed.'))).toBe(
      true
    );
  });

  it('ignores ordinary runtime errors', () => {
    expect(isDynamicImportError(new Error('Cannot read properties of undefined'))).toBe(false);
  });
});
