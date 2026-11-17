import { render, screen, waitFor, cleanup } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';

const complete = vi.hoisted(() => vi.fn());
const navigate = vi.hoisted(() => vi.fn(() => Promise.resolve()));
const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }));
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigate }));
vi.mock('sonner', () => ({ toast }));
vi.mock('../../../../shared/lib/clients', () => ({
  githubClient: { completeGitHubLink: complete },
}));
import GitHubLinkCallback from './GitHubLinkCallback';

beforeAll(() => initTestI18n());
beforeEach(() => vi.clearAllMocks());
afterEach(cleanup);

function renderCallback(search: string) {
  window.history.replaceState(null, '', `/flow/github/callback${search}`);
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <GitHubLinkCallback />
    </QueryClientProvider>
  );
}

const settings = { to: '/settings/flow', search: { section: 'accounts' } };

describe('GitHub link callback', () => {
  it('hands the code and state to the server once, then opens Connected accounts', async () => {
    complete.mockResolvedValue({ link: { linked: true, login: 'octocat' } });
    renderCallback('?code=abc123&state=link.1.2.3.ff.sig');
    expect(screen.getByRole('status')).toHaveTextContent('Linking your GitHub account…');
    await waitFor(() => expect(navigate).toHaveBeenCalledWith(settings));
    expect(complete).toHaveBeenCalledTimes(1);
    expect(complete).toHaveBeenCalledWith({ code: 'abc123', state: 'link.1.2.3.ff.sig' });
    // The one-time code leaves the address bar before anything else.
    expect(navigate).toHaveBeenNthCalledWith(1, { to: '/flow/github/callback', replace: true });
    expect(toast.success).toHaveBeenCalledWith('GitHub account linked');
  });

  it('reports a cancelled authorization without calling the server', async () => {
    renderCallback('?error=access_denied&error_description=denied&state=s');
    await waitFor(() => expect(navigate).toHaveBeenCalledWith(settings));
    expect(complete).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('GitHub linking was cancelled');
  });
});
