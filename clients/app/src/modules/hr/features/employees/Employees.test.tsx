import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { normalizeEmployeeRecord } from '../../shared/hooks/queries/useEmployees';

const mocks = vi.hoisted(() => ({
  useListEmployees: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useEmployees', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../shared/hooks/queries/useEmployees')>();
  return {
    ...actual,
    useListEmployees: mocks.useListEmployees,
  };
});

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('@tanstack/react-virtual', () => ({
  useVirtualizer: ({ count }: { count: number }) => ({
    getTotalSize: () => count * 56,
    getVirtualItems: () =>
      Array.from({ length: count }, (_, index) => ({
        key: index,
        index,
        start: index * 56,
      })),
    measureElement: () => undefined,
  }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import Employees from './Employees';

describe('Employees directory normalization and rendering', () => {
  beforeEach(() => {
    mocks.useListEmployees.mockReset();
  });

  it('normalizes staff-shaped rows and never renders hrEmployees.contractTypes.undefined', () => {
    const autoProvisioned = normalizeEmployeeRecord({
      id: '31',
      ownerId: '10',
      userId: '31',
      departmentId: '0',
      locationId: '0',
      firstName: 'Nadia',
      lastName: 'Barista',
      email: 'nadia@twoshots.test',
      employmentType: 'full_time',
      status: 'active',
      startDate: '0001-01-01',
      contractType: '',
      hireDate: '',
      currency: '',
      accountCreationMode: 'hr_created',
      jobTitle: '',
      createdAt: '0',
      updatedAt: '0',
    });

    const missingContract = normalizeEmployeeRecord({
      id: '32',
      ownerId: '10',
      userId: '32',
      departmentId: '0',
      locationId: '0',
      firstName: '',
      lastName: '',
      email: 'barista2@twoshots.test',
      status: 'active',
      contractType: '',
      hireDate: '',
      currency: '',
      accountCreationMode: 'hr_created',
      jobTitle: '',
      createdAt: '0',
      updatedAt: '0',
    });

    mocks.useListEmployees.mockReturnValue({
      data: { employees: [autoProvisioned, missingContract], total: 2, limit: 50, offset: 0 },
      isLoading: false,
      isError: false,
      isFetching: false,
      refetch: vi.fn(),
    });

    render(<Employees />);

    expect(screen.getByText('Nadia Barista')).toBeInTheDocument();
    expect(screen.getByText('barista2@twoshots.test')).toBeInTheDocument();
    expect(screen.getByText('hrEmployees.contractTypes.full_time')).toBeInTheDocument();
    expect(screen.queryByText('hrEmployees.contractTypes.undefined')).not.toBeInTheDocument();
    expect(screen.queryByText('0001-01-01')).not.toBeInTheDocument();
  });
});
