import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * SubmissionsAdmin empty-state coverage. The screen has two distinct
 * empty states - true-empty (no applications at all) and
 * filtered-empty (no applications at the selected stage) - each with
 * its own message and call-to-action. This test exercises both and
 * verifies the CTAs are wired to real handlers (refetch / clear
 * filter), not just prose that references a button that does nothing.
 */

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn(), warning: vi.fn() },
}));

const submissionsRefetch = vi.hoisted(() => vi.fn());
const useAdminSubmissionsMock = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/useAdmin', () => ({
  useAdminSubmissions: (...args: unknown[]) => useAdminSubmissionsMock(...args),
  useUpdateSubmissionStatus: () => ({ mutate: vi.fn() }),
  useAdminFormFields: () => ({ data: { fields: [] } }),
  fetchResumeURL: vi.fn(),
}));

import SubmissionsAdmin from './SubmissionsAdmin';

beforeEach(() => {
  vi.clearAllMocks();
  useAdminSubmissionsMock.mockReturnValue({
    data: { submissions: [] },
    isLoading: false,
    refetch: submissionsRefetch,
  });
});

describe('SubmissionsAdmin empty state', () => {
  it('shows the true-empty message and a refresh CTA when no filter is active', () => {
    render(<SubmissionsAdmin />);

    expect(screen.getByText('common.emptyStates.submissionsNone')).toBeInTheDocument();
    expect(
      screen.queryByText('common.emptyStates.submissionsNoneFiltered')
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.emptyStates.refresh' }));
    expect(submissionsRefetch).toHaveBeenCalledTimes(1);
  });

  it('switches to the filtered-empty message with a clear-filter CTA once a stage is picked', () => {
    render(<SubmissionsAdmin />);

    fireEvent.click(screen.getByRole('button', { name: 'reviewed' }));

    expect(screen.getByText('common.emptyStates.submissionsNoneFiltered')).toBeInTheDocument();
    expect(screen.queryByText('common.emptyStates.submissionsNone')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.emptyStates.clearFilter' }));

    // Clearing the filter flips the empty state back to the
    // unfiltered message - proving the CTA calls the real filter
    // reset handler rather than just being decorative.
    expect(screen.getByText('common.emptyStates.submissionsNone')).toBeInTheDocument();
  });
});
