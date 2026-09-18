import React from 'react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { AppGuard } from './AppGuard';

vi.mock('@voxire/config', () => ({
  resolveAuthUrl: () => 'https://auth.example.com',
  resolveAppUrl: () => 'https://dash.example.com',
}));

vi.mock('@voxire/observability', () => ({
  Sentry: { captureMessage: vi.fn() },
}));

vi.mock('../cookies', () => ({
  getCookie: vi.fn(() => 'token-value'),
}));

vi.mock('../signOut', () => ({
  signOut: vi.fn(),
}));

const fetchUserAllowed = (): Promise<Record<string, unknown>> =>
  Promise.resolve({ id: 1, email: 'a@b.c', fullName: 'A', role: 'owner', apps: ['crm'] });

afterEach(() => cleanup());

describe('AppGuard', () => {
  it('announces the default loading spinner to assistive technology (#43)', () => {
    render(
      <AppGuard appId="crm" fetchUser={() => new Promise<Record<string, unknown>>(() => {})}>
        <span>app</span>
      </AppGuard>
    );
    const status = screen.getByRole('status');
    expect(status).toHaveAccessibleName('Loading');
  });

  it('renders the default session-error UI when verification fails', async () => {
    render(
      <AppGuard appId="crm" fetchUser={() => Promise.reject(new Error('boom'))}>
        <span>app</span>
      </AppGuard>
    );
    expect(await screen.findByText('Could not verify your session.')).toBeInTheDocument();
  });

  it('lets apps override the error state and hands back a working retry (#42)', async () => {
    let calls = 0;
    render(
      <AppGuard
        appId="crm"
        fetchUser={(): Promise<Record<string, unknown>> => {
          calls += 1;
          return calls === 1 ? Promise.reject(new Error('boom')) : fetchUserAllowed();
        }}
        error={(retry) => (
          <button type="button" onClick={retry}>
            translated-error
          </button>
        )}
      >
        <span>app</span>
      </AppGuard>
    );

    expect(await screen.findByText('translated-error')).toBeInTheDocument();
    expect(screen.queryByText('Could not verify your session.')).toBeNull();

    fireEvent.click(screen.getByText('translated-error'));
    expect(await screen.findByText('app')).toBeInTheDocument();
    expect(calls).toBe(2);
  });
});
