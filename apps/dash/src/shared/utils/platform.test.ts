import { describe, it, expect } from 'vitest';
import { isTauri, isWeb, getPlatform, getPlatformCapabilities, platformSwitch } from './platform';

describe('isTauri', () => {
  it('always returns false in the web build', () => {
    expect(isTauri()).toBe(false);
  });
});

describe('isWeb', () => {
  it('always returns true in the web build', () => {
    expect(isWeb()).toBe(true);
  });
});

describe('getPlatform', () => {
  it('returns web', () => {
    expect(getPlatform()).toBe('web');
  });
});

describe('getPlatformCapabilities', () => {
  it('has no file system access', () => {
    expect(getPlatformCapabilities().hasFileSystem).toBe(false);
  });
  it('has no system tray', () => {
    expect(getPlatformCapabilities().hasSystemTray).toBe(false);
  });
});

describe('platformSwitch', () => {
  it('calls the web handler', () => {
    expect(platformSwitch({ web: () => 42 })).toBe(42);
  });
  it('falls back to default', () => {
    expect(platformSwitch({ default: () => 'default' })).toBe('default');
  });
  it('returns undefined when no handler matches', () => {
    expect(platformSwitch({})).toBeUndefined();
  });
});
