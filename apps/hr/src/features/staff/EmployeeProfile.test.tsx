import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  useStaffById: vi.fn(),
  useEmployeeProfile: vi.fn(),
  useDocuments: vi.fn(),
  useTrainingProgress: vi.fn(),
  useBenefitEnrollments: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useNavigate: () => mocks.navigate,
  useParams: () => ({ id: '1' }),
}));

vi.mock('../../shared/hooks/useHrExtras', () => ({
  useStaffById: mocks.useStaffById,
  useEmployeeProfile: mocks.useEmployeeProfile,
  useDocuments: mocks.useDocuments,
  useTrainingProgress: mocks.useTrainingProgress,
  useBenefitEnrollments: mocks.useBenefitEnrollments,
}));

import EmployeeProfile from './EmployeeProfile';

const STAFF = {
  id: 1,
  name: 'Alex Chen',
  role: 'Server',
  department: 'FOH',
  employmentType: 'Full-time',
  status: 'Active',
  email: 'alex@rtylr.com',
};

describe('EmployeeProfile empty states', () => {
  beforeEach(() => {
    mocks.navigate.mockReset();
    mocks.useStaffById.mockReset();
    mocks.useEmployeeProfile.mockReset();
    mocks.useDocuments.mockReset();
    mocks.useTrainingProgress.mockReset();
    mocks.useBenefitEnrollments.mockReset();

    mocks.useEmployeeProfile.mockReturnValue({ data: undefined });
    mocks.useDocuments.mockReturnValue({ data: [] });
    mocks.useTrainingProgress.mockReturnValue({ data: [] });
    mocks.useBenefitEnrollments.mockReturnValue({ data: [] });
  });

  it('shows a not-found empty state and navigates back to staff on CTA click', () => {
    mocks.useStaffById.mockReturnValue({ data: undefined, isLoading: false });

    render(<EmployeeProfile />);

    expect(screen.getByText('staff.profile.notFoundTitle')).toBeInTheDocument();
    expect(screen.getByText('staff.profile.notFoundDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'staff.profile.backToStaff' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/staff' });
  });

  it('shows the documents, training and benefits empty states with working CTAs', () => {
    mocks.useStaffById.mockReturnValue({ data: STAFF, isLoading: false });

    render(<EmployeeProfile />);

    expect(screen.getByText('staff.profile.noDocuments')).toBeInTheDocument();
    expect(screen.getByText('staff.profile.noCourses')).toBeInTheDocument();
    expect(screen.getByText('staff.profile.noBenefits')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'staff.profile.goToDocuments' }));
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/documents' });

    fireEvent.click(screen.getByRole('button', { name: 'staff.profile.goToTraining' }));
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/training' });

    fireEvent.click(screen.getByRole('button', { name: 'staff.profile.goToBenefits' }));
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/benefits' });
  });

  it('does not show the documents empty state once documents exist for this staff member', () => {
    mocks.useStaffById.mockReturnValue({ data: STAFF, isLoading: false });
    mocks.useDocuments.mockReturnValue({
      data: [
        {
          id: 1,
          staffId: 1,
          name: 'Contract.pdf',
          type: 'PDF',
          sizeKb: 120,
          uploadedAt: '2026-01-01',
        },
      ],
    });

    render(<EmployeeProfile />);

    expect(screen.queryByText('staff.profile.noDocuments')).not.toBeInTheDocument();
    expect(screen.getByText('Contract.pdf')).toBeInTheDocument();
  });
});
