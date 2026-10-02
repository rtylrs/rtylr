import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * CareersAdmin empty state - the old copy told the admin to "Click
 * New role" without the empty state's own button doing anything.
 * This verifies the EmptyState's CTA now opens the create-role
 * drawer directly (the same setCreating(true) the header's "New
 * role" button calls).
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

vi.mock('./CareerEditor', () => ({
  default: ({ onClose }: { onClose: () => void }) => (
    <div>
      <div>career-editor</div>
      <button type="button" onClick={onClose}>
        close-editor
      </button>
    </div>
  ),
}));

vi.mock('../../shared/hooks/queries/useAdmin', () => ({
  useAdminCareers: () => ({ data: { careers: [] }, isLoading: false }),
  useAdminSubmissions: () => ({ data: { submissions: [] } }),
  useCreateCareer: () => ({ mutate: vi.fn() }),
  useUpdateCareer: () => ({ mutate: vi.fn() }),
  useDeleteCareer: () => ({ mutate: vi.fn() }),
  useAdminCustomization: () => ({ data: { customization: null } }),
}));

import CareersAdmin from './CareersAdmin';
import { useUserStore } from '../../../../shared/stores/userStore';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

beforeEach(() => {
  vi.clearAllMocks();
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess('recruit', { isOwner: true }) },
  });
});

describe('CareersAdmin empty state', () => {
  it('shows the empty message and opens the create-role editor from its own CTA', () => {
    render(<CareersAdmin />);

    expect(screen.getByText('recruitCommon.emptyStates.rolesNone')).toBeInTheDocument();
    expect(screen.queryByText('career-editor')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'recruitCommon.emptyStates.createRole' }));

    // The CTA is wired to the same setCreating(true) the header's own
    // "New role" button uses - the editor drawer now renders.
    expect(screen.getByText('career-editor')).toBeInTheDocument();
  });
});

it('hides role creation from a viewer', () => {
  useUserStore.setState({
    user: {
      effectiveAccess: createEffectiveAccess('recruit', { permissions: ['recruit:jobs:view'] }),
    },
  });
  render(<CareersAdmin />);
  expect(
    screen.queryByRole('button', { name: 'recruitAdmin.careers.newRole' })
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole('button', { name: 'recruitCommon.emptyStates.createRole' })
  ).not.toBeInTheDocument();
});
