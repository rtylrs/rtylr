// @vitest-environment happy-dom
import React from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

const apiState = vi.hoisted(() => ({
  projects: undefined as Record<string, unknown>[] | undefined,
  milestones: undefined as Record<string, unknown>[] | undefined,
}));
const storeState = vi.hoisted(() => ({
  projects: [] as Record<string, unknown>[],
  milestones: [] as Record<string, unknown>[],
}));

vi.mock('../projects/useProjects', () => ({
  useProjects: () => ({ data: apiState.projects }),
}));
vi.mock('../milestones/useMilestones', () => ({
  useMilestones: () => ({ data: apiState.milestones }),
}));
vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: (selector: (state: typeof storeState) => unknown) => selector(storeState),
}));
vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children: React.ReactNode }) => <a href="/">{children}</a>,
}));

import Timeline from './Timeline';

function daysFromToday(days: number): string {
  const now = new Date();
  return new Date(now.getFullYear(), now.getMonth(), now.getDate() + days).toISOString();
}

describe('Timeline project bars (#1258)', () => {
  afterEach(() => {
    apiState.projects = undefined;
    apiState.milestones = undefined;
    storeState.projects = [];
    storeState.milestones = [];
  });

  it('derives bar geometry and progress from the API project fields', () => {
    apiState.projects = [
      {
        id: 'p1',
        name: 'Web rebuild',
        color: '#123456',
        startDate: daysFromToday(-5),
        dueDate: daysFromToday(25),
        taskCount: 4,
        completedCount: 3,
        memberAvatars: [],
      },
    ];

    render(<Timeline />);

    // Raw ApiProject rows carry dueDate and no progress, so reading targetDate
    // left the bar at the invalid CSS width "NaN%" and the label without a
    // number. dueDate + 3/4 tasks must become a real width and "75%".
    const label = screen.getByText('75%');
    const bar = label.parentElement as HTMLElement;
    expect(bar.style.width).toMatch(/%$/);
    expect(bar.style.width).not.toContain('NaN');
    expect(Number.parseFloat(bar.style.width)).toBeGreaterThan(0);
  });

  it('falls back to the store projects until the API query has data', () => {
    storeState.projects = [
      {
        id: 'p2',
        name: 'From the store',
        color: '#654321',
        startDate: daysFromToday(-2),
        targetDate: daysFromToday(20),
        progress: 40,
      },
    ];

    render(<Timeline />);

    expect(screen.getByText('40%')).toBeInTheDocument();
  });
});
