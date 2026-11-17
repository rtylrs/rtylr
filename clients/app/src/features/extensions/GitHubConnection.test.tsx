import path from 'node:path';
import { render, screen, waitFor, within, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeAll, beforeEach, afterEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';

const api = vi.hoisted(() => ({
  get: vi.fn(),
  start: vi.fn(),
  complete: vi.fn(),
  disconnect: vi.fn(),
}));
const navigate = vi.hoisted(() => vi.fn());
const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn(), warning: vi.fn() }));
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigate }));
vi.mock('sonner', () => ({ toast }));
vi.mock('../../shared/lib/clients', () => ({
  githubClient: {
    getGitHubConnection: api.get,
    startGitHubInstall: api.start,
    completeGitHubInstall: api.complete,
    disconnectGitHub: api.disconnect,
  },
}));
import GitHubConnection from './GitHubConnection';

const locales = path.resolve(import.meta.dirname, '../../locales');
beforeAll(() => initTestI18nFrom(locales));
beforeEach(() => {
  vi.clearAllMocks();
  window.history.replaceState(null, '', '/extensions');
  api.get.mockResolvedValue({ connection: { configured: true, canManage: true } });
});
afterEach(cleanup);

function renderCard() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <GitHubConnection />
    </QueryClientProvider>
  );
  return userEvent.setup();
}

describe('GitHub connection card', () => {
  it('sends an admin to the GitHub install page the server returns', async () => {
    const assign = vi.fn();
    vi.spyOn(window, 'location', 'get').mockReturnValue({
      ...window.location,
      search: '',
      assign,
    });
    api.start.mockResolvedValue({
      installUrl: 'https://github.com/apps/rtylr/installations/new?state=s',
    });
    const ui = renderCard();
    await ui.click(await screen.findByRole('button', { name: 'Connect GitHub' }));
    await waitFor(() =>
      expect(assign).toHaveBeenCalledWith('https://github.com/apps/rtylr/installations/new?state=s')
    );
    vi.restoreAllMocks();
  });

  it('completes the install once when GitHub sends the admin back, then clears the URL', async () => {
    window.history.replaceState(
      null,
      '',
      '/extensions?installation_id=42&setup_action=install&state=abc.def'
    );
    api.complete.mockResolvedValue({});
    renderCard();
    await waitFor(() =>
      expect(api.complete).toHaveBeenCalledWith({ installationId: '42', state: 'abc.def' })
    );
    expect(api.complete).toHaveBeenCalledTimes(1);
    expect(navigate).toHaveBeenCalledWith({ to: '/extensions', replace: true });
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith('GitHub connected'));
  });

  it('only clears the URL when GitHub returns from its own settings without a state', async () => {
    window.history.replaceState(null, '', '/extensions?installation_id=42&setup_action=update');
    renderCard();
    await waitFor(() =>
      expect(navigate).toHaveBeenCalledWith({ to: '/extensions', replace: true })
    );
    expect(api.complete).not.toHaveBeenCalled();
  });

  it('shows health and asks before disconnecting', async () => {
    api.get.mockResolvedValue({
      connection: {
        configured: true,
        connected: true,
        canManage: true,
        accountLogin: 'acme',
        accountType: 'Organization',
        status: 'suspended',
      },
    });
    api.disconnect.mockResolvedValue({});
    const ui = renderCard();
    expect(await screen.findByText('Suspended on GitHub')).toBeVisible();
    expect(screen.getByText(/Connected to acme \(organisation\)/)).toBeVisible();
    await ui.click(screen.getByRole('button', { name: 'Disconnect' }));
    expect(api.disconnect).not.toHaveBeenCalled();
    await ui.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Disconnect' }));
    await waitFor(() => expect(api.disconnect).toHaveBeenCalledTimes(1));
    // uninstalled was omitted (false), so the admin is told to finish on GitHub.
    await waitFor(() => expect(toast.warning).toHaveBeenCalled());
  });

  it('does not let a non-admin connect, and says when the server has no app', async () => {
    api.get.mockResolvedValue({ connection: { configured: true } });
    renderCard();
    expect(await screen.findByRole('button', { name: 'Connect GitHub' })).toBeDisabled();
    cleanup();
    api.get.mockResolvedValue({ connection: { canManage: true } });
    renderCard();
    expect(await screen.findByText('GitHub is not set up on this server yet.')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Connect GitHub' })).toBeDisabled();
  });
});
