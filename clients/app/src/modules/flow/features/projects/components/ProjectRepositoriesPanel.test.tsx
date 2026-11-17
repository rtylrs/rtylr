import React from 'react';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: { when?: string }) =>
      options?.when ? `${key} ${options.when}` : key,
    i18n: { language: 'en' },
  }),
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

const github = vi.hoisted(() => ({
  connection: {
    data: { configured: true, connected: true, accountLogin: 'acme', status: 'active' },
    isLoading: false,
  } as Record<string, unknown>,
  selected: [{ id: '2', fullName: 'acme/web', private: true }],
  available: [
    { id: '1', fullName: 'acme/api', private: false },
    { id: '2', fullName: 'acme/web', private: true },
  ],
  availableCalls: [] as boolean[],
  mutate: vi.fn(),
}));
vi.mock('../../../../../shared/hooks/queries/useGitHub', () => ({
  useGitHubConnection: () => github.connection,
  useProjectRepositories: () => ({ data: github.selected, isLoading: false }),
  useInstallationRepositories: (_projectId: string, enabled: boolean) => {
    github.availableCalls.push(enabled);
    return { data: enabled ? github.available : undefined, isLoading: false };
  },
  useSetProjectRepositories: () => ({ mutate: github.mutate, isPending: false }),
}));

import ProjectRepositoriesPanel from './ProjectRepositoriesPanel';

const connectionData = github.connection.data;
beforeEach(() => {
  github.mutate.mockClear();
  github.availableCalls = [];
  github.connection.data = connectionData;
});
afterEach(cleanup);

describe('ProjectRepositoriesPanel', () => {
  it('lets a project admin pick repositories and saves the whole selection', () => {
    render(<ProjectRepositoriesPanel projectId="p1" canAdminister />);
    const save = screen.getByRole('button', { name: 'flow.projectSettings.code.save' });
    expect(save).toBeDisabled();
    expect(screen.getByRole('checkbox', { name: /acme\/web/ })).toBeChecked();

    fireEvent.click(screen.getByRole('checkbox', { name: /acme\/api/ }));
    expect(save).toBeEnabled();
    fireEvent.click(save);

    expect(github.mutate).toHaveBeenCalledTimes(1);
    expect([...github.mutate.mock.calls[0][0]].sort()).toEqual(['1', '2']);
  });

  it('shows a member the picked repositories without the installation list or a save button', () => {
    render(<ProjectRepositoriesPanel projectId="p1" canAdminister={false} />);
    expect(screen.getByText('flow.projectSettings.code.gateReason')).toBeVisible();
    expect(screen.getByText('acme/web')).toBeVisible();
    expect(screen.queryByRole('checkbox')).toBeNull();
    expect(screen.queryByRole('button', { name: 'flow.projectSettings.code.save' })).toBeNull();
    expect(github.availableCalls.every((enabled) => !enabled)).toBe(true);
  });

  it('says when the last webhook arrived, or that none has', () => {
    render(<ProjectRepositoriesPanel projectId="p1" canAdminister={false} />);
    expect(screen.getByText('flow.projectSettings.code.noWebhooks')).toBeVisible();
    cleanup();

    const twoHoursAgo = new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString();
    github.connection.data = { ...connectionData, lastWebhookAt: twoHoursAgo };
    render(<ProjectRepositoriesPanel projectId="p1" canAdminister={false} />);
    expect(screen.getByText('flow.projectSettings.code.lastWebhook 2 hours ago')).toBeVisible();
  });
});
