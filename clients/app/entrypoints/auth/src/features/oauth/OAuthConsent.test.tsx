// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, vars?: { scope?: string }) => (vars?.scope ? `${key}:${vars.scope}` : key),
  }),
}));

vi.mock('../../shared/hooks/useAuth', () => ({
  useLogin: () => ({
    mutate: (_creds: unknown, opts: { onSuccess: () => void }) => opts.onSuccess(),
    isPending: false,
  }),
}));

vi.mock('./useOAuth', () => ({
  useIssueAuthCode: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('../../shared/components/AuthShell', () => ({
  default: ({ children }: { children: React.ReactNode }) =>
    React.createElement('div', null, children),
}));

import OAuthConsent from './OAuthConsent';

function openAuthorize(redirectUri: string, scope: string): void {
  const q = new URLSearchParams({
    client_id: 'rtylr-mcp',
    redirect_uri: redirectUri,
    scope,
    state: 'abc',
    code_challenge: 'challenge',
    code_challenge_method: 'S256',
  });
  window.history.pushState({}, '', `/oauth?${q.toString()}`);
}

describe('OAuthConsent', () => {
  beforeEach(() => {
    window.history.pushState({}, '', '/');
  });

  it('accepts the loopback redirect the MCP client uses and lists every requested scope', async () => {
    openAuthorize('http://127.0.0.1:51234/callback', 'flow:read pos:write bogus:scope');
    render(<OAuthConsent />);

    expect(screen.queryByText('auth.oauth.errInvalidRedirect')).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('auth.common.email'), {
      target: { value: 'user@example.com' },
    });
    fireEvent.change(screen.getByLabelText('auth.common.password'), {
      target: { value: 'Password123' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'auth.oauth.continue' }));

    await waitFor(() => {
      expect(screen.getByText('auth.oauth.scopeFlowRead')).toBeInTheDocument();
    });
    expect(screen.getByText('auth.oauth.scopePosWrite')).toBeInTheDocument();
    expect(screen.getByText('auth.oauth.scopeUnknown:bogus:scope')).toBeInTheDocument();
  });

  it('refuses a redirect outside rtylr and loopback', () => {
    openAuthorize('https://evil.example.com/callback', 'flow:read');
    render(<OAuthConsent />);
    expect(screen.getByText('auth.oauth.errInvalidRedirect')).toBeInTheDocument();
  });

  it('refuses a plain-http rtylr redirect', () => {
    openAuthorize('http://app.rtylr.com/callback', 'flow:read');
    render(<OAuthConsent />);
    expect(screen.getByText('auth.oauth.errInvalidRedirect')).toBeInTheDocument();
  });

  it('refuses a request that asks for no scope', () => {
    openAuthorize('http://127.0.0.1:51234/callback', '');
    render(<OAuthConsent />);
    expect(screen.getByText('auth.oauth.errNoScopes')).toBeInTheDocument();
  });
});
