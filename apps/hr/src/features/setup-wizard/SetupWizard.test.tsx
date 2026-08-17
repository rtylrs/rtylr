import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { GetSetupStatusResponse } from '@voxire/voxie/hr/setup/v1/types';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  setCurrentStepIndex: vi.fn(),
  useGetSetupStatus: vi.fn(),
  completeStep: vi.fn(),
  upsertConfig: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useSetup', () => ({
  useGetSetupStatus: mocks.useGetSetupStatus,
  useCompleteSetupStep: () => ({ mutate: mocks.completeStep, isPending: false }),
  useUpsertHRConfig: () => ({ mutate: mocks.upsertConfig, isPending: false }),
}));

vi.mock('../../shared/stores/hrSetupStore', () => ({
  useHRSetupStore: () => ({
    currentStepIndex: 0,
    setCurrentStepIndex: mocks.setCurrentStepIndex,
  }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

// The wizard's own redirect logic is what is under test here, so every step
// component is stubbed - they each pull in their own queries and forms.
vi.mock('./components/WizardProgress', () => ({
  default: () => <div data-testid="wizard-progress" />,
  STEP_KEYS: [
    'company_profile',
    'work_schedule',
    'org_structure',
    'locations',
    'roles_permissions',
    'payroll',
    'notifications',
    'employees',
  ],
}));
vi.mock('./steps/CompanyProfileStep', () => ({ default: () => <div data-testid="step-first" /> }));
vi.mock('./steps/WorkScheduleStep', () => ({ default: () => null }));
vi.mock('./steps/OrgStructureStep', () => ({ default: () => null }));
vi.mock('./steps/LocationsStep', () => ({ default: () => null }));
vi.mock('./steps/RolesPermissionsStep', () => ({ default: () => null }));
vi.mock('./steps/PayrollConfigStep', () => ({ default: () => null }));
vi.mock('./steps/NotificationsStep', () => ({ default: () => null }));
vi.mock('./steps/EmployeesStep', () => ({ default: () => <div data-testid="step-last" /> }));

import SetupWizard from './SetupWizard';

const LAST_STEP_INDEX = 7;

function statusQuery(data: GetSetupStatusResponse) {
  return { data, isLoading: false, isError: false, refetch: vi.fn() };
}

/** All eight steps ticked, mirroring what GET /hr/setup/status derives. */
const allStepsDone = [
  'company_profile',
  'work_schedule',
  'org_structure',
  'locations',
  'roles_permissions',
  'payroll',
  'notifications',
  'employees',
].map((stepKey) => ({ stepKey, completedAt: '1699999999' }));

describe('SetupWizard completion redirect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // The freeze in issue #725: /hr/setup/status reports every step done while
  // /hr/config has no setup_completed_at, so HRConfigRoute sends / -> /setup
  // and the wizard sent /setup -> / on every render.
  it('does not bounce back to the dashboard when the config was never finalised', () => {
    mocks.useGetSetupStatus.mockReturnValue(
      statusQuery({
        steps: allStepsDone,
        config: { setupCompletedAt: undefined },
        isComplete: true,
      })
    );

    render(<SetupWizard />);

    expect(mocks.navigate).not.toHaveBeenCalled();
    expect(screen.getByTestId('wizard-progress')).toBeInTheDocument();
  });

  it('opens the final step so an unfinalised setup can be completed', () => {
    mocks.useGetSetupStatus.mockReturnValue(
      statusQuery({
        steps: allStepsDone,
        config: { setupCompletedAt: undefined },
        isComplete: true,
      })
    );

    render(<SetupWizard />);

    expect(mocks.setCurrentStepIndex).toHaveBeenCalledWith(LAST_STEP_INDEX);
  });

  it('returns to the dashboard once the config carries a completion timestamp', () => {
    mocks.useGetSetupStatus.mockReturnValue(
      statusQuery({
        steps: allStepsDone,
        config: { setupCompletedAt: '1699999999' },
        isComplete: true,
      })
    );

    render(<SetupWizard />);

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/', replace: true });
  });

  it('stays in the wizard while steps are still outstanding', () => {
    mocks.useGetSetupStatus.mockReturnValue(
      statusQuery({
        steps: [{ stepKey: 'company_profile', completedAt: '1699999999' }],
        config: { setupCompletedAt: undefined },
        isComplete: false,
      })
    );

    render(<SetupWizard />);

    expect(mocks.navigate).not.toHaveBeenCalled();
    expect(mocks.setCurrentStepIndex).toHaveBeenCalledWith(1);
  });
});
