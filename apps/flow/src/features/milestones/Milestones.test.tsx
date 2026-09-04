import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import Milestones from './Milestones';

const mutateSpy = vi.fn();

// The create form asks which project the milestone belongs to, so the screen
// now reads the project list too.
vi.mock('../projects/useProjects', () => ({
  useProjects: () => ({ data: [{ id: 'p-1', name: 'Website Redesign' }], isLoading: false }),
}));

const milestonesState = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));

vi.mock('./useMilestones', () => ({
  useMilestones: () => ({ data: milestonesState.data, isLoading: false }),
  useCreateMilestone: () => ({ mutate: mutateSpy, isPending: false }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ activeWorkspaceId: 'ws1' }),
}));

describe('Milestones create form', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-08-20T12:00:00.000Z'));
    mutateSpy.mockClear();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('submits the exact date picked by the user, not the current date', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.namePlaceholder'), {
      target: { value: 'Launch' },
    });

    fireEvent.click(screen.getByLabelText('flow.milestones.targetDate'));
    fireEvent.click(screen.getByRole('gridcell', { name: /august 31, 2026/i }));

    fireEvent.click(screen.getByText('flow.milestones.create'));

    expect(mutateSpy).toHaveBeenCalledTimes(1);
    expect(mutateSpy.mock.calls[0][0].targetDate).toBe('2026-08-31');
  });

  it('never submits a milestone without an explicit target date', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.namePlaceholder'), {
      target: { value: 'Launch' },
    });

    // No date picked - the Create button must be disabled so a click cannot
    // silently submit today's date in its place.
    const createButton = screen.getByText('flow.milestones.create');
    expect(createButton).toBeDisabled();

    fireEvent.click(createButton);
    expect(mutateSpy).not.toHaveBeenCalled();
  });

  it('gives the target date picker an accessible name that opens a calendar', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));

    fireEvent.click(screen.getByLabelText('flow.milestones.targetDate'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });
});

describe('Milestones progress bar (#1080)', () => {
  function milestone(over: Record<string, unknown> = {}): Record<string, unknown> {
    return {
      id: 'm-1',
      projectId: 'p-1',
      name: 'Launch',
      status: 'active',
      targetDate: '2026-09-30',
      taskCount: 0,
      doneCount: 0,
      ...over,
    };
  }

  beforeEach(() => {
    milestonesState.data = [];
  });

  afterEach(() => {
    milestonesState.data = [];
  });

  function fillWidth(container: HTMLElement): string | undefined {
    // The fill is the only element inside the rounded track.
    const track = container.querySelector('.rounded-full.overflow-hidden');
    return (track?.firstElementChild as HTMLElement | null)?.style.width;
  }

  it('renders an empty track when progress is unknown', () => {
    // computedProgress is optional. `width: undefined%` is invalid, the
    // browser drops the declaration, and the fill took its parent's width, so
    // a milestone with unknown progress looked like a completed one.
    milestonesState.data = [milestone({ computedProgress: undefined })];
    const { container } = render(<Milestones />);
    expect(fillWidth(container)).toBe('0%');
  });

  it('says there are no tasks yet instead of a bare percent sign', () => {
    milestonesState.data = [milestone({ computedProgress: undefined })];
    render(<Milestones />);
    expect(screen.getByText('flow.milestones.noTasksYet')).toBeInTheDocument();
    expect(screen.queryByText('%')).not.toBeInTheDocument();
  });

  it('renders the real percentage when there are tasks', () => {
    milestonesState.data = [milestone({ taskCount: 19, doneCount: 7, computedProgress: 37 })];
    const { container } = render(<Milestones />);
    expect(screen.getByText('37%')).toBeInTheDocument();
    expect(fillWidth(container)).toBe('37%');
  });

  it('clamps a percentage above a hundred', () => {
    milestonesState.data = [milestone({ taskCount: 3, doneCount: 3, computedProgress: 140 })];
    const { container } = render(<Milestones />);
    expect(fillWidth(container)).toBe('100%');
  });

  it('shows a full track only when the work is actually done', () => {
    milestonesState.data = [milestone({ taskCount: 4, doneCount: 4, computedProgress: 100 })];
    const { container } = render(<Milestones />);
    expect(fillWidth(container)).toBe('100%');
    expect(screen.getByText('100%')).toBeInTheDocument();
  });
});
