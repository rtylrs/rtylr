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

vi.mock('./useMilestones', () => ({
  useMilestones: () => ({ data: [], isLoading: false }),
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
