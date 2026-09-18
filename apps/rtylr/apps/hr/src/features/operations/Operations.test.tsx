import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useUserStore: vi.fn(),
  useSwaps: vi.fn(),
  createSwapMutate: vi.fn(),
  resolveSwapMutate: vi.fn(),
  useBreaks: vi.fn(),
  startBreakMutate: vi.fn(),
  endBreakMutate: vi.fn(),
  useOvertimeAlerts: vi.fn(),
  useHrSettings: vi.fn(),
  useAvailability: vi.fn(),
  usePayroll: vi.fn(),
  useStaff: vi.fn(),
  refetchSwaps: vi.fn(),
  refetchAvailability: vi.fn(),
  refetchOvertimeAlerts: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: mocks.useUserStore,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../../shared/hooks/useHrExtras', () => ({
  useSwaps: mocks.useSwaps,
  useCreateSwap: () => ({ mutate: mocks.createSwapMutate }),
  useResolveSwap: () => ({ mutate: mocks.resolveSwapMutate }),
  useBreaks: mocks.useBreaks,
  useStartBreak: () => ({ mutate: mocks.startBreakMutate }),
  useEndBreak: () => ({ mutate: mocks.endBreakMutate }),
  useOvertimeAlerts: mocks.useOvertimeAlerts,
  useHrSettings: mocks.useHrSettings,
}));

vi.mock('../../shared/hooks/queries/useAvailability', () => ({
  useAvailability: mocks.useAvailability,
}));

vi.mock('../payroll/usePayroll', () => ({
  usePayroll: mocks.usePayroll,
}));

vi.mock('../staff/useStaff', () => ({
  useStaff: mocks.useStaff,
}));

import Operations from './Operations';

function role(r: string): void {
  mocks.useUserStore.mockImplementation(
    (selector: (state: { user: { role: string } }) => unknown) => selector({ user: { role: r } })
  );
}

describe('Operations empty states', () => {
  beforeEach(() => {
    Object.values(mocks).forEach((m) => {
      if (typeof m === 'function' && 'mockReset' in m) m.mockReset();
    });

    role('owner');
    mocks.useStaff.mockReturnValue({ data: [] });
    mocks.usePayroll.mockReturnValue({ data: [] });
    mocks.useSwaps.mockReturnValue({ data: [], refetch: mocks.refetchSwaps });
    mocks.useBreaks.mockReturnValue({ data: [] });
    mocks.useOvertimeAlerts.mockReturnValue({ data: [], refetch: mocks.refetchOvertimeAlerts });
    mocks.useHrSettings.mockReturnValue({ data: { currency: 'USD' } });
    mocks.useAvailability.mockReturnValue({ data: [], refetch: mocks.refetchAvailability });
  });

  it('shows the swaps empty state and refetches on CTA click (Swap Shifts tab)', () => {
    render(<Operations />);

    fireEvent.click(screen.getByRole('button', { name: 'Swap Shifts' }));

    expect(screen.getByText('operations.emptySwapsTitle')).toBeInTheDocument();
    expect(screen.getByText('operations.emptySwapsDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(mocks.refetchSwaps).toHaveBeenCalledOnce();
  });

  it('shows the availability empty state and refetches on CTA click', () => {
    render(<Operations />);

    fireEvent.click(screen.getByRole('button', { name: 'Availability' }));

    expect(screen.getByText('operations.availability.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('operations.availability.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(mocks.refetchAvailability).toHaveBeenCalledOnce();
  });

  it('shows the breaks empty state and starts a lunch break on CTA click', () => {
    render(<Operations />);

    fireEvent.click(screen.getByRole('button', { name: 'Breaks' }));

    expect(screen.getByText('operations.emptyBreaksTitle')).toBeInTheDocument();
    expect(screen.getByText('operations.emptyBreaksDescription')).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'operations.startLunchBreak' })[0]);

    expect(mocks.startBreakMutate).toHaveBeenCalledOnce();
  });

  it('shows the overtime empty state and refetches on CTA click', () => {
    render(<Operations />);

    fireEvent.click(screen.getByRole('button', { name: 'Overtime' }));

    expect(screen.getByText('operations.emptyOvertimeTitle')).toBeInTheDocument();
    expect(screen.getByText('operations.emptyOvertimeDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(mocks.refetchOvertimeAlerts).toHaveBeenCalledOnce();
  });
});
