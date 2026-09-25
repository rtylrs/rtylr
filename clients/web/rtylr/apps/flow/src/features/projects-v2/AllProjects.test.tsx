import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// t returns the key, so a label that goes through i18n is distinguishable from
// one that renders a raw contract value: 'flow.projectStatus.active' vs 'active'.
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

const refetchAllProjects = vi.hoisted(() => vi.fn());
const listState = vi.hoisted(() => ({
  value: {
    data: { projects: [] as Array<Record<string, unknown>>, total: 0 },
    isLoading: false,
    isError: false,
    isPaused: false,
    failureCount: 0,
    refetch: refetchAllProjects,
  },
}));

vi.mock('./useAllProjects', () => ({
  useAllProjects: () => listState.value,
}));

import AllProjects from './AllProjects';

beforeEach(() => {
  listState.value = { data: { projects: [], total: 0 }, isLoading: false };
});

describe('AllProjects filters', () => {
  // The status filter rendered the bare contract value while the visibility
  // filter beside it went through t(), so the two dropdowns disagreed about
  // whether they were translated.
  it('translates every status option instead of showing the contract value', () => {
    render(<AllProjects />);

    expect(screen.getByRole('option', { name: 'flow.projectStatus.active' })).toBeInTheDocument();
    expect(
      screen.getByRole('option', { name: 'flow.projectStatus.completed' })
    ).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'active' })).not.toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'completed' })).not.toBeInTheDocument();
  });

  it('translates the visibility options too', () => {
    render(<AllProjects />);

    expect(screen.getByRole('option', { name: 'flow.visibility.private' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'flow.visibility.public' })).toBeInTheDocument();
  });
});

// P9: a project the caller sees only because they manage someone on it is
// grouped separately. Mixed into one list it would read as the caller's own.
describe('AllProjects manager-proxy grouping', () => {
  const own = { id: '1', name: 'My Project', visibleViaReport: false };
  const report = { id: '2', name: 'Report Project', visibleViaReport: true };

  it("groups a report project under its own heading, apart from the caller's", () => {
    listState.value = { data: { projects: [own, report], total: 2 }, isLoading: false };

    render(<AllProjects />);

    expect(screen.getByText('flow.allProjects.yourProjects')).toBeInTheDocument();
    expect(screen.getByText('flow.allProjects.yourReports')).toBeInTheDocument();
    expect(screen.getByText('My Project')).toBeInTheDocument();
    expect(screen.getByText('Report Project')).toBeInTheDocument();
  });

  // The common case. A caller who manages nobody gets no headings at all, so
  // the grouping is invisible rather than an empty section.
  it('renders one ungrouped list when nothing is visible via a report', () => {
    listState.value = { data: { projects: [own], total: 1 }, isLoading: false };

    render(<AllProjects />);

    expect(screen.queryByText('flow.allProjects.yourReports')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.allProjects.yourProjects')).not.toBeInTheDocument();
    expect(screen.getByText('My Project')).toBeInTheDocument();
  });
});

describe('AllProjects: a failed load is not an empty project list', () => {
  beforeEach(() => {
    listState.value = {
      data: { projects: [], total: 0 },
      isLoading: false,
      isError: false,
      isPaused: false,
      failureCount: 0,
      refetch: refetchAllProjects,
    };
    refetchAllProjects.mockClear();
  });

  it('shows the failure state instead of "create your first project"', () => {
    listState.value = { ...listState.value, isError: true };
    render(<AllProjects />);
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.allProjects.emptyTitle')).not.toBeInTheDocument();
  });

  it('retries the query from the failure state', () => {
    listState.value = { ...listState.value, isError: true };
    render(<AllProjects />);
    fireEvent.click(screen.getByText('flow.loadFailed.retry'));
    expect(refetchAllProjects).toHaveBeenCalled();
  });

  it('still shows the empty state when the request succeeded and there is nothing', () => {
    render(<AllProjects />);
    expect(screen.getByText('flow.allProjects.emptyTitle')).toBeInTheDocument();
  });
});
