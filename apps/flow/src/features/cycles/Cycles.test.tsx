import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import Cycles from './Cycles';

const mutateSpy = vi.fn();

vi.mock('./useCycles', () => ({
  useCycles: () => ({ data: [], isLoading: false }),
  useCreateCycle: () => ({ mutate: mutateSpy, isPending: false }),
}));

describe('Cycles create form', () => {
  beforeEach(() => {
    mutateSpy.mockClear();
  });

  it('gives the start and end date inputs accessible names', () => {
    render(<Cycles />);

    fireEvent.click(screen.getByText('flow.cycles.newCycle'));

    expect(screen.getByLabelText('flow.cycles.startDate')).toHaveAttribute('type', 'date');
    expect(screen.getByLabelText('flow.cycles.endDate')).toHaveAttribute('type', 'date');
  });

  it('submits the exact start and end dates picked by the user', () => {
    render(<Cycles />);

    fireEvent.click(screen.getByText('flow.cycles.newCycle'));
    fireEvent.change(screen.getByPlaceholderText('flow.cycles.namePlaceholder'), {
      target: { value: 'Sprint 1' },
    });
    fireEvent.change(screen.getByLabelText('flow.cycles.startDate'), {
      target: { value: '2026-09-01' },
    });
    fireEvent.change(screen.getByLabelText('flow.cycles.endDate'), {
      target: { value: '2026-09-14' },
    });

    fireEvent.click(screen.getByText('flow.cycles.create'));

    expect(mutateSpy).toHaveBeenCalledTimes(1);
    expect(mutateSpy.mock.calls[0][0]).toMatchObject({
      name: 'Sprint 1',
      startDate: '2026-09-01',
      endDate: '2026-09-14',
    });
  });
});
