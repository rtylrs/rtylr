import { render, screen, waitFor, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../../shared/testing/i18n';

const api = vi.hoisted(() => ({
  connection: vi.fn(),
  link: vi.fn(),
  start: vi.fn(),
  unlink: vi.fn(),
}));
const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }));
vi.mock('sonner', () => ({ toast }));
vi.mock('../../../../../shared/lib/clients', () => ({
  githubClient: {
    getGitHubConnection: api.connection,
    getMyGitHubLink: api.link,
    startGitHubLink: api.start,
    unlinkGitHub: api.unlink,
  },
}));
import AccountsSection from './AccountsSection';

const connected = { configured: true, connected: true, accountLogin: 'acme', linkingEnabled: true };

beforeAll(() => initTestI18n());
beforeEach(() => {
  vi.clearAllMocks();
  api.connection.mockResolvedValue({ connection: connected });
  api.link.mockResolvedValue({ link: {} });
});
afterEach(cleanup);

function renderSection() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <AccountsSection />
    </QueryClientProvider>
  );
  return userEvent.setup();
}

describe('Connected accounts settings', () => {
  it('hides the GitHub row until the business connects GitHub', async () => {
    api.connection.mockResolvedValue({ connection: { configured: true } });
    renderSection();
    expect(
      await screen.findByText('Nothing to link yet. GitHub is not connected for this business.')
    ).toBeVisible();
    expect(screen.queryByText('GitHub account')).toBeNull();
    expect(api.link).not.toHaveBeenCalled();
  });

  it('sends the member to the GitHub authorize page the server returns', async () => {
    const assign = vi.fn();
    vi.spyOn(window, 'location', 'get').mockReturnValue({ ...window.location, assign });
    api.start.mockResolvedValue({
      authorizeUrl: 'https://github.com/login/oauth/authorize?client_id=x&state=s',
    });
    const ui = renderSection();
    expect(await screen.findByText('Not linked')).toBeVisible();
    await ui.click(screen.getByRole('button', { name: 'Link GitHub' }));
    await waitFor(() =>
      expect(assign).toHaveBeenCalledWith(
        'https://github.com/login/oauth/authorize?client_id=x&state=s'
      )
    );
    vi.restoreAllMocks();
  });

  it('shows the linked login and unlinks it', async () => {
    api.link.mockResolvedValue({ link: { linked: true, login: 'octocat' } });
    api.unlink.mockResolvedValue({});
    const ui = renderSection();
    expect(await screen.findByText('Linked as @octocat')).toBeVisible();
    await ui.click(screen.getByRole('button', { name: 'Unlink' }));
    await waitFor(() => expect(api.unlink).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith('GitHub account unlinked'));
  });

  it('cannot start linking when the server has no OAuth credentials', async () => {
    api.connection.mockResolvedValue({ connection: { ...connected, linkingEnabled: false } });
    renderSection();
    expect(
      await screen.findByText('Linking accounts is not set up on this server yet.')
    ).toBeVisible();
    expect(screen.getByRole('button', { name: 'Link GitHub' })).toBeDisabled();
  });
});
