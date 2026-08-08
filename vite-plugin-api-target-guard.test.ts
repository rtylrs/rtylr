import { describe, expect, it } from 'vitest';
import {
  buildWarningLines,
  isLocalApiTarget,
  resolveApiTarget,
} from './vite-plugin-api-target-guard';

describe('isLocalApiTarget', () => {
  it.each([
    'http://localhost:9999/',
    'http://localhost/',
    'http://127.0.0.1:9999/',
    'http://0.0.0.0:9999/',
    'http://[::1]:9999/',
    'http://host.docker.internal:9999/',
    'http://onion.localhost:9999/',
  ])('treats %s as local', (url) => {
    expect(isLocalApiTarget(url)).toBe(true);
  });

  it.each([
    'https://api.voxire.com/',
    'https://api.rtylr.com/',
    'http://staging.internal/',
    // Deliberately NOT local: a private-LAN address is more often a shared
    // staging box than someone's laptop, and a false "you're safe" is the
    // failure this guard exists to prevent.
    'http://192.168.1.10:9999/',
    'http://10.0.0.4/',
  ])('treats %s as non-local', (url) => {
    expect(isLocalApiTarget(url)).toBe(false);
  });

  it('treats an unparseable target as non-local so it still warns', () => {
    expect(isLocalApiTarget('not a url')).toBe(false);
    expect(isLocalApiTarget('')).toBe(false);
  });
});

describe('resolveApiTarget', () => {
  it('falls back to the hosted default when VITE_API_URL is unset', () => {
    expect(resolveApiTarget({})).toEqual({
      url: 'https://api.voxire.com/',
      source: 'hosted-default',
      local: false,
    });
  });

  it('falls back to the hosted default when VITE_API_URL is blank', () => {
    expect(resolveApiTarget({ VITE_API_URL: '   ' }).source).toBe('hosted-default');
    expect(resolveApiTarget({ VITE_API_URL: '' }).local).toBe(false);
  });

  it('reports a configured local target as local', () => {
    expect(resolveApiTarget({ VITE_API_URL: 'http://localhost:9999/' })).toEqual({
      url: 'http://localhost:9999/',
      source: 'env',
      local: true,
    });
  });

  it('reports a configured non-local target as non-local', () => {
    const target = resolveApiTarget({ VITE_API_URL: 'https://api.voxire.com/' });
    expect(target.source).toBe('env');
    expect(target.local).toBe(false);
  });
});

describe('buildWarningLines', () => {
  it('names the app, the target, and the exact file to create', () => {
    const lines = buildWarningLines('flow', resolveApiTarget({}));
    const text = lines.join('\n');

    expect(text).toContain('NON-LOCAL API');
    expect(text).toContain('flow');
    expect(text).toContain('https://api.voxire.com/');
    expect(text).toContain('cp .env.local.example apps/flow/.env.local');
  });

  it('says the variable is unset when the hosted default was used', () => {
    const text = buildWarningLines('dash', resolveApiTarget({})).join('\n');
    expect(text).toContain('VITE_API_URL is not set');
  });

  it('attributes an explicitly configured target to VITE_API_URL', () => {
    const text = buildWarningLines(
      'dash',
      resolveApiTarget({ VITE_API_URL: 'https://api.voxire.com/' })
    ).join('\n');
    expect(text).toContain('source  VITE_API_URL');
    expect(text).not.toContain('is not set');
  });
});
