import { describe, expect, it } from 'vitest';
import {
  HOSTED_API_URL,
  HOSTED_AUTH_URL,
  resolveApiBaseUrl,
  resolveAuthUrl,
  resolveWsBaseUrl,
} from './serverConfig';

describe('serverConfig', () => {
  it('uses the configured hosted API with one trailing slash', () => {
    expect(resolveApiBaseUrl()).toMatch(/^https?:\/\/.+\/$/);
    expect(resolveApiBaseUrl().endsWith('//')).toBe(false);
    expect(HOSTED_API_URL).toBe('https://api.voxire.com/');
  });

  it('derives the realtime scheme from the hosted API scheme', () => {
    const apiUrl = new URL(resolveApiBaseUrl());
    const expectedScheme = apiUrl.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = new URL(resolveWsBaseUrl());

    expect(wsUrl.protocol).toBe(expectedScheme);
    expect(wsUrl.host).toBe(apiUrl.host);
  });

  it('always sends sign-in through the hosted auth app', () => {
    expect(resolveAuthUrl()).toBe(HOSTED_AUTH_URL);
    expect(HOSTED_AUTH_URL).toBe('https://sso.rtylr.com');
  });
});
