import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useGetEmployee: vi.fn(),
  useUserStore: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useEmployees', () => ({
  useGetEmployee: mocks.useGetEmployee,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: mocks.useUserStore,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '42' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('../documents/DocumentVault', () => ({
  default: () => null,
}));

import EmployeeDetail from './EmployeeDetail';

const baseEmployee = {
  id: 42,
  userId: 7,
  jobTitle: 'Line Cook',
  currency: 'USD',
  status: 'active',
  contractType: 'full_time',
  hireDate: '2024-01-01',
  probationEndDate: '2024-04-01',
  monthlySalary: 5000,
  hourlyRate: undefined,
  accountCreationMode: 'invite',
};

describe('EmployeeDetail compensation gating', () => {
  beforeEach(() => {
    mocks.useGetEmployee.mockReset();
    mocks.useUserStore.mockReset();
    mocks.useGetEmployee.mockReturnValue({ data: baseEmployee, isLoading: false });
  });

  it('hides salary for a staff-role user', () => {
    mocks.useUserStore.mockImplementation(
      (selector: (state: { user: { role: string } }) => unknown) =>
        selector({ user: { role: 'staff' } })
    );

    render(<EmployeeDetail />);
    fireEvent.click(screen.getByText('employees.employmentInfo'));

    expect(screen.queryByText('employees.salary')).not.toBeInTheDocument();
    expect(screen.queryByText('5000 USD/mo')).not.toBeInTheDocument();
  });

  it('shows salary for an admin-role user', () => {
    mocks.useUserStore.mockImplementation(
      (selector: (state: { user: { role: string } }) => unknown) =>
        selector({ user: { role: 'admin' } })
    );

    render(<EmployeeDetail />);
    fireEvent.click(screen.getByText('employees.employmentInfo'));

    expect(screen.getByText('employees.salary')).toBeInTheDocument();
    expect(screen.getByText('5000 USD/mo')).toBeInTheDocument();
  });

  it('shows salary for an owner-role user', () => {
    mocks.useUserStore.mockImplementation(
      (selector: (state: { user: { role: string } }) => unknown) =>
        selector({ user: { role: 'owner' } })
    );

    render(<EmployeeDetail />);
    fireEvent.click(screen.getByText('employees.employmentInfo'));

    expect(screen.getByText('employees.salary')).toBeInTheDocument();
  });
});
