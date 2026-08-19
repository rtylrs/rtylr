import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import Milestones from './Milestones';

const mutateSpy = vi.fn();

vi.mock('./useMilestones', () => ({
  useMilestones: () => ({ data: [], isLoading: false }),
  useCreateMilestone: () => ({ mutate: mutateSpy, isPending: false }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ activeWorkspaceId: 'ws1' }),
}));

describe('Milestones create form', () => {
  beforeEach(() => {
    mutateSpy.mockClear();
  });

  it('submits the exact date picked by the user, not the current date', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));
    fireEvent.change(screen.getByPlaceholderText('flow.milestones.namePlaceholder'), {
      target: { value: 'Launch' },
    });

    const dateInput = screen.getByLabelText('flow.milestones.targetDate') as HTMLInputElement;
    fireEvent.change(dateInput, { target: { value: '2026-08-31' } });

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

  it('gives the target date input an accessible name', () => {
    render(<Milestones />);

    fireEvent.click(screen.getByText('flow.milestones.createFirst'));

    expect(screen.getByLabelText('flow.milestones.targetDate')).toHaveAttribute('type', 'date');
  });
});
