import React from 'react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const location = vi.hoisted(() => ({ pathname: '/flow/my-issues' }));
vi.mock('@tanstack/react-router', () => ({ useLocation: () => location }));
vi.mock('../lib/clients', () => ({
  userClient: { updateAvatar: vi.fn(), listBusinessAvatars: vi.fn() },
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (key: string) => key }) }));

import { useUserStore } from '../stores/userStore';
import { FirstLoginAvatarPrompt } from './FirstLoginAvatarPrompt';

const renderPrompt = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <FirstLoginAvatarPrompt />
    </QueryClientProvider>
  );

const settle = () =>
  act(() => {
    vi.advanceTimersByTime(2000);
  });

beforeEach(() => {
  vi.useFakeTimers();
  localStorage.clear();
  location.pathname = '/flow/my-issues';
  useUserStore.setState({ user: { id: 7, businessId: 1, fullName: 'Abed', onboarded: true } });
});

afterEach(() => {
  vi.useRealTimers();
});

describe('first-login avatar welcome', () => {
  it('opens once for someone with no avatar, and offers Skip for now', () => {
    renderPrompt();
    expect(screen.queryByText('settings.avatar.welcomeTitle')).toBeNull();
    settle();
    expect(screen.getByText('settings.avatar.welcomeTitle')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'settings.avatar.skip' }));
    expect(screen.queryByText('settings.avatar.welcomeTitle')).toBeNull();
  });

  it('never opens again after it was shown', () => {
    const first = renderPrompt();
    settle();
    first.unmount();
    renderPrompt();
    settle();
    expect(screen.queryByText('settings.avatar.welcomeTitle')).toBeNull();
  });

  it('stays away from someone who already chose an avatar', () => {
    useUserStore.setState({
      user: { id: 7, businessId: 1, onboarded: true, avatar: { kind: 'initials', color: 1 } },
    });
    renderPrompt();
    settle();
    expect(screen.queryByText('settings.avatar.welcomeTitle')).toBeNull();
  });

  it('waits until onboarding is finished', () => {
    location.pathname = '/onboarding';
    renderPrompt();
    settle();
    expect(screen.queryByText('settings.avatar.welcomeTitle')).toBeNull();
  });
});
