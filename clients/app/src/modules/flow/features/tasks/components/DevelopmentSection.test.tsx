import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const links = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));
vi.mock('../useTaskLinks', () => ({ useTaskLinks: () => ({ data: links.data }) }));

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
});
