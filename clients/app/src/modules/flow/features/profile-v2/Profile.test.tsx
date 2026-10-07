import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const profileState = vi.hoisted(() => ({
  data: undefined as Record<string, unknown> | undefined,
  isLoading: false,
  isPending: false,
  isSuccess: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: () => profileState,
}));

vi.mock('../../shared/lib/clients', () => ({ profileClient: { getProfile: vi.fn() } }));

import { Profile } from './Profile';

beforeEach(() => {
  profileState.data = {
    name: 'Ada Amouneh',
    initials: 'AA',
    roleTitle: 'Site foreman',
    publicProjects: [],
  };
  profileState.isLoading = false;
  profileState.isPending = false;
  profileState.isSuccess = true;
  profileState.isError = false;
  profileState.isPaused = false;
  profileState.failureCount = 0;
});

describe('Profile identity header (#1095)', () => {
  it('renders the name on the largest type step', () => {
    const { container } = render(<Profile userId="9" />);
    const heading = container.querySelector('h1');
    expect(heading?.textContent).toBe('Ada Amouneh');
    expect(heading?.className).toContain('text-flow-screen');
  });

  it('sizes the avatar to 64px', () => {
    const { container } = render(<Profile userId="9" />);
    expect(container.querySelector('.h-16.w-16')).not.toBeNull();
  });

  it('shows the role at body size, muted', () => {
    render(<Profile userId="9" />);
    const role = screen.getByText('Site foreman');
    expect(role.className).toContain('text-flow-body');
  });

  it('shows retry rather than calling a failed profile request not found', () => {
    profileState.data = undefined;
    profileState.isError = true;
    profileState.isSuccess = false;
    render(<Profile userId="9" />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.profile.notFound')).not.toBeInTheDocument();
  });
});

describe('Profile public projects as chips (#1095)', () => {
  it('renders each project as a chip carrying its colour', () => {
    profileState.data = {
      ...profileState.data,
      publicProjects: [
        { id: '1', name: 'Site fitout', color: '#f56d8f' },
        { id: '2', name: 'Warehouse racking', color: '#5ed6a0' },
      ],
    };
    const { container } = render(<Profile userId="9" />);

    expect(screen.getByText('Site fitout')).toBeInTheDocument();
    expect(screen.getByText('Warehouse racking')).toBeInTheDocument();

    // The colour is the point: it did not appear at all before.
    const dots = [...container.querySelectorAll('.h-2.w-2.rounded-full')];
    expect(dots).toHaveLength(2);
    // happy-dom keeps the authored hex rather than normalising to rgb().
    expect(dots.map((d) => (d as HTMLElement).style.background)).toEqual(['#f56d8f', '#5ed6a0']);
  });

  it('falls back to the accent for a project with no colour of its own', () => {
    profileState.data = {
      ...profileState.data,
      publicProjects: [{ id: '1', name: 'Site fitout' }],
    };
    const { container } = render(<Profile userId="9" />);
    const dot = container.querySelector('.h-2.w-2.rounded-full') as HTMLElement;
    expect(dot.style.background).toContain('--accent');
  });

  it('renders no emoji icon, which would carry a raw name through as text', () => {
    profileState.data = {
      ...profileState.data,
      publicProjects: [{ id: '1', name: 'Site fitout', icon: 'folder' }],
    };
    render(<Profile userId="9" />);
    expect(screen.queryByText('folder')).not.toBeInTheDocument();
    expect(screen.queryByText('📋')).not.toBeInTheDocument();
  });

  it('says so when there are no public projects', () => {
    render(<Profile userId="9" />);
    expect(screen.getByText('flow.profile.noPublicProjects')).toBeInTheDocument();
  });

  it('translates every string it shows, including the counts it used to hardcode', () => {
    profileState.data = {
      ...profileState.data,
      publicProjects: [{ id: '1', name: 'Site fitout', taskCount: 4, completedCount: 2 }],
    };
    render(<Profile userId="9" />);
    // "4 tasks" and "2 done" were hardcoded English. The chip carries neither.
    expect(screen.queryByText(/tasks$/)).not.toBeInTheDocument();
    expect(screen.queryByText(/done$/)).not.toBeInTheDocument();
  });
});

describe('Profile not-found state (#1095)', () => {
  it('is a designed notice rather than a bare sentence', () => {
    profileState.isError = true;
    profileState.data = undefined;
    render(<Profile userId="9" />);

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('flow.profile.notFound');
    expect(screen.getByText('flow.profile.notFoundHint')).toBeInTheDocument();
  });
});
