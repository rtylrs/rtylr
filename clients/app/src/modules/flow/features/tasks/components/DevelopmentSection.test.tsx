import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const links = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));
vi.mock('../useTaskLinks', () => ({ useTaskLinks: () => ({ data: links.data }) }));
const pullRequests = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));
vi.mock('../useTaskPullRequests', () => ({
  useTaskPullRequests: () => ({ data: pullRequests.data }),
}));

import DevelopmentSection from './DevelopmentSection';

describe('DevelopmentSection', () => {
  it('lists code links and leaves design and docs links to the Links section', () => {
    links.data = [
      {
        id: '1',
        url: 'https://github.com/a/b/pull/2',
        kind: 'github_pr',
        title: 'Fix login',
        ref: 'a/b#2',
        state: 'open',
      },
      {
        id: '2',
        url: 'https://www.figma.com/design/x',
        kind: 'figma',
        title: 'Mockups',
        ref: 'figma.com',
        state: '',
      },
    ];
    render(<DevelopmentSection taskId="1" />);
    expect(screen.getByRole('link', { name: 'Fix login' })).toHaveAttribute(
      'href',
      'https://github.com/a/b/pull/2'
    );
    expect(screen.queryByRole('link', { name: 'Mockups' })).toBeNull();
    expect(screen.queryByText('flow.taskDetail.developmentEmpty')).toBeNull();
  });

  it('shows the empty line when the task has no code links', () => {
    links.data = [];
    render(<DevelopmentSection taskId="1" />);
    expect(screen.getByText('flow.taskDetail.developmentEmpty')).toBeInTheDocument();
  });

  it('shows the CI checks under a linked pull request, with links to their logs (T-376)', () => {
    links.data = [
      {
        id: '1',
        url: 'https://github.com/a/b/pull/2',
        kind: 'github_pr',
        title: 'Fix login',
        ref: 'a/b#2',
        state: 'open',
      },
      {
        id: '3',
        url: 'https://github.com/a/b/pull/3',
        kind: 'github_pr',
        title: 'Docs',
        ref: 'a/b#3',
        state: 'open',
      },
    ];
    pullRequests.data = [
      {
        ref: 'a/b#2',
        known: true,
        checksState: 'failing',
        checks: [
          {
            name: 'make verify',
            status: 'completed',
            conclusion: 'failure',
            detailsUrl: 'https://ci.example/1',
          },
          { name: 'lint', status: 'in_progress', conclusion: '', detailsUrl: '' },
        ],
      },
      { ref: 'a/b#3', known: false, checksState: 'none', checks: [] },
    ];
    render(<DevelopmentSection taskId="1" />);
    expect(screen.getByRole('img', { name: 'flow.checks.state.failing' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'make verify' })).toHaveAttribute(
      'href',
      'https://ci.example/1'
    );
    expect(screen.getByText('flow.checks.check.failed')).toBeInTheDocument();
    expect(screen.getByText('flow.checks.check.running')).toBeInTheDocument();
    expect(screen.getByText('flow.checks.waiting')).toBeInTheDocument();
  });
});
