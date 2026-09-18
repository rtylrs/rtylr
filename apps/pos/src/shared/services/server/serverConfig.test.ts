import { describe, it, expect, beforeEach } from 'vitest';
import {
  normalizeBaseUrl,
  isValidServerUrl,
  deriveSelfHostAuthUrl,
  resolveApiBaseUrl,
  resolveWsBaseUrl,
  resolveAuthUrl,
  SERVER_STORE_KEY,
  HOSTED_API_URL,
  HOSTED_AUTH_URL,
} from './serverConfig';

describe('serverConfig', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe('normalizeBaseUrl', () => {
    it('appends trailing slash to clean base URL', () => {
      expect(normalizeBaseUrl('http://localhost:8000')).toBe('http://localhost:8000/');
    });

    it('ensures only a single trailing slash', () => {
      expect(normalizeBaseUrl('http://localhost:8000///')).toBe('http://localhost:8000/');
    });

    it('trims leading/trailing whitespace', () => {
      expect(normalizeBaseUrl('  https://my-api.com/v1/   ')).toBe('https://my-api.com/v1/');
    });
  });

  describe('isValidServerUrl', () => {
    it('returns true for valid absolute http and https URLs', () => {
      expect(isValidServerUrl('http://localhost')).toBe(true);
      expect(isValidServerUrl('https://mycompany.com/api')).toBe(true);
    });

    it('returns false for invalid or relative URLs', () => {
      expect(isValidServerUrl('/relative/path')).toBe(false);
      expect(isValidServerUrl('ftp://my-server.com')).toBe(false);
      expect(isValidServerUrl('not-a-url')).toBe(false);
    });
  });

  describe('deriveSelfHostAuthUrl', () => {
    it('replaces api. subdomain with auth.', () => {
      expect(deriveSelfHostAuthUrl('https://api.mycompany.com')).toBe('https://auth.mycompany.com');
      expect(deriveSelfHostAuthUrl('http://api.company.local:8000/')).toBe('http://auth.company.local:8000');
    });

    it('replaces api substring variations', () => {
      expect(deriveSelfHostAuthUrl('https://rtylr-api.example.com')).toBe('https://rtylr-auth.example.com');
      expect(deriveSelfHostAuthUrl('https://api-rtylr.example.com')).toBe('https://auth-rtylr.example.com');
    });

    it('replaces /api path segment', () => {
      expect(deriveSelfHostAuthUrl('https://mycompany.com/api')).toBe('https://mycompany.com/auth');
      expect(deriveSelfHostAuthUrl('https://mycompany.com/api/v1/')).toBe('https://mycompany.com/auth');
      expect(deriveSelfHostAuthUrl('https://mycompany.com/custom/api/v1')).toBe('https://mycompany.com/custom/auth');
    });

    it('falls back to appending /auth if no matches found', () => {
      expect(deriveSelfHostAuthUrl('http://192.168.1.50:8000/')).toBe('http://192.168.1.50:8000/auth');
    });

    it('returns HOSTED_AUTH_URL on parsing error', () => {
      expect(deriveSelfHostAuthUrl('invalid-url')).toBe(HOSTED_AUTH_URL);
    });
  });

  describe('resolveApiBaseUrl', () => {
    it('returns hosted API URL by default', () => {
      expect(resolveApiBaseUrl()).toBe(HOSTED_API_URL);
    });

    it('returns persisted self-host URL in self-host mode', () => {
      localStorage.setItem(
        SERVER_STORE_KEY,
        JSON.stringify({
          state: {
            mode: 'self_host',
            customApiUrl: 'https://mycompany.com/api',
          },
        })
      );
      expect(resolveApiBaseUrl()).toBe('https://mycompany.com/api/');
    });
  });

  describe('resolveWsBaseUrl', () => {
    it('returns secure websocket URL for https base', () => {
      localStorage.setItem(
        SERVER_STORE_KEY,
        JSON.stringify({
          state: {
            mode: 'self_host',
            customApiUrl: 'https://api.mycompany.com',
          },
        })
      );
      expect(resolveWsBaseUrl()).toBe('wss://api.mycompany.com');
    });

    it('returns insecure websocket URL for http base', () => {
      localStorage.setItem(
        SERVER_STORE_KEY,
        JSON.stringify({
          state: {
            mode: 'self_host',
            customApiUrl: 'http://192.168.1.100:8000',
          },
        })
      );
      expect(resolveWsBaseUrl()).toBe('ws://192.168.1.100:8000');
    });
  });

  describe('resolveAuthUrl', () => {
    it('returns HOSTED_AUTH_URL by default', () => {
      expect(resolveAuthUrl()).toBe(HOSTED_AUTH_URL);
    });

    it('returns explicit authUrl in self-host mode if provided', () => {
      localStorage.setItem(
        SERVER_STORE_KEY,
        JSON.stringify({
          state: {
            mode: 'self_host',
            authUrl: 'https://auth.custom-domain.com/',
          },
        })
      );
      expect(resolveAuthUrl()).toBe('https://auth.custom-domain.com');
    });

    it('derives authUrl in self-host mode if authUrl is missing but customApiUrl is present', () => {
      localStorage.setItem(
        SERVER_STORE_KEY,
        JSON.stringify({
          state: {
            mode: 'self_host',
            customApiUrl: 'https://api.mycompany.com/',
            authUrl: null,
          },
        })
      );
      expect(resolveAuthUrl()).toBe('https://auth.mycompany.com');
    });
  });
});
