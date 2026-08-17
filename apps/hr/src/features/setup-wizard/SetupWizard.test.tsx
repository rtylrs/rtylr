import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// Hoisted with the mocks: vi.mock factories are lifted above plain top-level
// consts, so the WizardProgress factory below cannot close over one.
const mocks = vi.hoisted(() => ({
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
  useGetSetupStatus: vi.fn(),
  completeStep: vi.fn(),
  upsertConfig: vi.fn(),
  setCurrentStepIndex: vi.fn(),
  navigate: vi.fn(),
  currentStepIndex: 0,
}));

const STEP_KEYS = mocks.STEP_KEYS;

vi.mock('../../shared/hooks/queries/useSetup', () => ({
  useGetSetupStatus: mocks.useGetSetupStatus,
  useCompleteSetupStep: () => ({ mutate: mocks.completeStep, isPending: false }),
  useUpsertHRConfig: () => ({ mutate: mocks.upsertConfig, isPending: false }),
}));

vi.mock('../../shared/stores/hrSetupStore', () => ({
  useHRSetupStore: () => ({
    currentStepIndex: mocks.currentStepIndex,
    setCurrentStepIndex: mocks.setCurrentStepIndex,
  }),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('./components/WizardProgress', () => ({
  default: () => <div data-testid="wizard-progress" />,
  STEP_KEYS: mocks.STEP_KEYS,
}));

vi.mock('./steps/CompanyProfileStep', () => ({
  default: () => <div data-testid="step-company_profile" />,
}));
vi.mock('./steps/WorkScheduleStep', () => ({
  default: () => <div data-testid="step-work_schedule" />,
}));
vi.mock('./steps/OrgStructureStep', () => ({
  default: () => <div data-testid="step-org_structure" />,
}));
vi.mock('./steps/LocationsStep', () => ({ default: () => <div data-testid="step-locations" /> }));
vi.mock('./steps/RolesPermissionsStep', () => ({
  default: () => <div data-testid="step-roles_permissions" />,
}));
vi.mock('./steps/PayrollConfigStep', () => ({ default: () => <div data-testid="step-payroll" /> }));
vi.mock('./steps/NotificationsStep', () => ({
  default: () => <div data-testid="step-notifications" />,
}));
vi.mock('./steps/EmployeesStep', () => ({ default: () => <div data-testid="step-employees" /> }));

import SetupWizard from './SetupWizard';

/** Every wizard step marked done, which is what drives `isComplete: true`. */
function allStepsDone() {
  return STEP_KEYS.map((stepKey, index) => ({
    id: index + 1,
    stepKey,
    completedAt: 1_700_000_000,
  }));
}

function statusResponse(overrides: {
  steps: ReturnType<typeof allStepsDone>;
  isComplete: boolean;
  setupCompletedAt?: string;
}) {
  return {
    data: {
      steps: overrides.steps,
      isComplete: overrides.isComplete,
      config: { setupCompletedAt: overrides.setupCompletedAt },
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  };
}

describe('SetupWizard', () => {
  beforeEach(() => {
    mocks.useGetSetupStatus.mockReset();
    mocks.setCurrentStepIndex.mockReset();
    mocks.navigate.mockReset();
    mocks.currentStepIndex = 0;
  });

  it('does not bounce back to / while the wizard is unfinalised', () => {
    // The loop case: every step is done, so /hr/setup/status reports
    // isComplete, but /hr/config's setupCompletedAt is still unset - which is
    // what HRConfigRoute gates on, so it keeps sending / here.
    mocks.useGetSetupStatus.mockReturnValue(
      statusResponse({ steps: allStepsDone(), isComplete: true, setupCompletedAt: undefined })
    );

    render(<SetupWizard />);

    expect(mocks.navigate).not.toHaveBeenCalled();
  });

  it('lands on the last step when every step is done but setup was never finalised', () => {
    mocks.useGetSetupStatus.mockReturnValue(
      statusResponse({ steps: allStepsDone(), isComplete: true, setupCompletedAt: undefined })
    );

    render(<SetupWizard />);

    // Otherwise the user restarts at step one with nothing left to complete and
    // can never reach the action that writes setupCompletedAt.
    expect(mocks.setCurrentStepIndex).toHaveBeenCalledWith(STEP_KEYS.length - 1);
  });

  it('redirects to / exactly once when setup is finalised', () => {
    mocks.useGetSetupStatus.mockReturnValue(
      statusResponse({ steps: allStepsDone(), isComplete: true, setupCompletedAt: '1700000000' })
    );

    render(<SetupWizard />);

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/', replace: true });
    expect(mocks.navigate).toHaveBeenCalledTimes(1);
  });

  it('stays on the first incomplete step while steps are outstanding', () => {
    const steps = allStepsDone().map((step, index) =>
      index >= 2 ? { ...step, completedAt: null } : step
    );
    mocks.useGetSetupStatus.mockReturnValue(
      statusResponse({ steps, isComplete: false, setupCompletedAt: undefined })
    );

    render(<SetupWizard />);

    expect(mocks.navigate).not.toHaveBeenCalled();
    expect(mocks.setCurrentStepIndex).toHaveBeenCalledWith(2);
    expect(screen.getByTestId('step-company_profile')).toBeInTheDocument();
  });
});
