import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTask } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children?: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useNavigate: () => navigateMock,
}));

const myIssuesState = vi.hoisted(() => ({ data: [] as ApiTask[] }));
vi.mock('../tasks/useMyIssues', () => ({
  useMyIssues: () => myIssuesState,
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ projects: [], activeWorkspaceId: 'ws-1' }),
}));

import FocusToday from './FocusToday';

beforeEach(() => {
  vi.clearAllMocks();
  myIssuesState.data = [];
});

describe('FocusToday empty state', () => {
  it('shows the empty title/hint with a CTA to My Tasks', () => {
    render(<FocusToday />);

    expect(screen.getByText('common.focusToday.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('common.focusToday.emptyHint')).toBeInTheDocument();

    const cta = screen.getByRole('button', { name: 'flow.shell.myTasks' });
    fireEvent.click(cta);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/my-issues' });
  });
});
