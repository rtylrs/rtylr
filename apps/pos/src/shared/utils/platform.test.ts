import { describe, it, expect } from 'vitest';
import {
  isTauri,
  isWeb,
  getPlatform,
  getPlatformCapabilities,
} from './platform';

// In the test environment window.__TAURI__ is not set, so we run as 'web'.

describe('isTauri', () => {
  it('returns false in the jsdom test environment', () => {
    expect(isTauri()).toBe(false);
  });
});

describe('isWeb', () => {
  it('returns true in the jsdom test environment', () => {
    expect(isWeb()).toBe(true);
  });
});

describe('getPlatform', () => {
  it('returns "web" in the test environment', () => {
    expect(getPlatform()).toBe('web');
  });
});

describe('getPlatformCapabilities', () => {
  it('reports no file system on web', () => {
    expect(getPlatformCapabilities().hasFileSystem).toBe(false);
  });

  it('reports no system tray on web', () => {
    expect(getPlatformCapabilities().hasSystemTray).toBe(false);
  });

  it('reports no menu bar on web', () => {
    expect(getPlatformCapabilities().hasMenuBar).toBe(false);
  });

  it('reports no deep linking on web', () => {
    expect(getPlatformCapabilities().supportsDeepLinking).toBe(false);
  });

  it('reports cannot minimize to tray on web', () => {
    expect(getPlatformCapabilities().canMinimizeToTray).toBe(false);
  });
});
