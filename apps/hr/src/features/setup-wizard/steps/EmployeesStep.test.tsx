import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useListEmployees: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { info: vi.fn() },
}));

vi.mock('../../../shared/hooks/queries/useEmployees', () => ({
  useListEmployees: mocks.useListEmployees,
}));

import EmployeesStep from './EmployeesStep';

describe('EmployeesStep empty state', () => {
  const onNext = vi.fn();
  const onBack = vi.fn();
  const onAddEmployee = vi.fn();

  beforeEach(() => {
    mocks.useListEmployees.mockReset();
    onNext.mockReset();
    onBack.mockReset();
    onAddEmployee.mockReset();
  });

  it('shows the empty state and calls onAddEmployee when the CTA is clicked', () => {
    mocks.useListEmployees.mockReturnValue({ data: { employees: [] } });

    render(<EmployeesStep onNext={onNext} onBack={onBack} onAddEmployee={onAddEmployee} />);

    expect(screen.getByText('setup.employees.noEmployeesHint')).toBeInTheDocument();

    // Two buttons share this label: a persistent toolbar button and the
    // EmptyState's own CTA. Click the second (the EmptyState's).
    const addButtons = screen.getAllByRole('button', { name: 'setup.employees.addEmployee' });
    expect(addButtons).toHaveLength(2);
    fireEvent.click(addButtons[1]);

    expect(onAddEmployee).toHaveBeenCalledOnce();
  });

  it('does not show the empty state once employees exist', () => {
    mocks.useListEmployees.mockReturnValue({
      data: {
        employees: [{ id: 1, jobTitle: 'Server', contractType: 'Full-time', status: 'active' }],
      },
    });

    render(<EmployeesStep onNext={onNext} onBack={onBack} onAddEmployee={onAddEmployee} />);

    expect(screen.queryByText('setup.employees.noEmployeesHint')).not.toBeInTheDocument();
  });
});
