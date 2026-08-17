import React, { useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import {
  useGetSetupStatus,
  useCompleteSetupStep,
  useUpsertHRConfig,
} from '../../shared/hooks/queries/useSetup';
import { useHRSetupStore } from '../../shared/stores/hrSetupStore';
import WizardProgress, { STEP_KEYS } from './components/WizardProgress';
import CompanyProfileStep from './steps/CompanyProfileStep';
import WorkScheduleStep from './steps/WorkScheduleStep';
import OrgStructureStep from './steps/OrgStructureStep';
import LocationsStep from './steps/LocationsStep';
import RolesPermissionsStep from './steps/RolesPermissionsStep';
import PayrollConfigStep from './steps/PayrollConfigStep';
import NotificationsStep from './steps/NotificationsStep';
import EmployeesStep from './steps/EmployeesStep';
import type {
  CompanyProfileFormData,
  WorkScheduleFormData,
  OrgStructureFormData,
  LocationsFormData,
  PayrollConfigFormData,
  NotificationsFormData,
} from '../../shared/schemas/wizardSchemas';

function PageLoader(): React.JSX.Element {
  return (
    <div className="flex items-center justify-center py-20">
      <div
        className="h-8 w-8 rounded-full border-2 animate-spin"
        style={{ borderColor: 'var(--hr-border)', borderTopColor: 'var(--accent)' }}
      />
    </div>
  );
}

const SetupWizard = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data, isLoading, isError, refetch } = useGetSetupStatus();
  const completeStep = useCompleteSetupStep();
  const upsertConfig = useUpsertHRConfig();
  const { currentStepIndex, setCurrentStepIndex } = useHRSetupStore();
  const [showNewEmployee, setShowNewEmployee] = useState(false);

  const steps = data?.steps ?? [];

  // The single source of truth for "HR setup is done". HRConfigRoute gates every
  // other HR route on this same field, so reading `isComplete` here instead let
  // the two guards disagree: with every step ticked but the wizard never
  // finalised, / redirected to /setup and /setup redirected straight back.
  // `isComplete` only means "all steps are ticked", which is a precondition -
  // the wizard's own last step is what writes setupCompletedAt.
  const isSetupFinalised = Boolean(data?.config?.setupCompletedAt);

  // Redirecting from the render body raised "Cannot update a component while
  // rendering a different component" and ran the bounce synchronously, which
  // froze the tab outright rather than merely flickering.
  useEffect(() => {
    if (isSetupFinalised) {
      navigate({ to: '/', replace: true });
    }
  }, [isSetupFinalised, navigate]);

  useEffect(() => {
    if (!data || isSetupFinalised) return;
    const firstIncomplete = STEP_KEYS.findIndex(
      (key) => !steps.some((s) => s.stepKey === key && s.completedAt != null)
    );
    // Every step ticked but never finalised: land on the last step so the action
    // that writes setupCompletedAt stays reachable, instead of restarting at one.
    setCurrentStepIndex(firstIncomplete >= 0 ? firstIncomplete : STEP_KEYS.length - 1);
  }, [data, isSetupFinalised, steps, setCurrentStepIndex]);

  if (isLoading) {
    return <PageLoader />;
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-20">
        <p className="text-sm" style={{ color: 'var(--hr-text)' }}>
          {t('setup.statusError')}
        </p>
        <button
          type="button"
          className="min-h-[44px] rounded-lg px-4 py-2 text-sm font-medium"
          style={{ backgroundColor: 'var(--accent)', color: 'var(--hr-card-bg)' }}
          onClick={(): void => {
            refetch();
          }}
        >
          {t('common.retry')}
        </button>
      </div>
    );
  }

  // The effect above owns the redirect; hold the loader until it lands.
  if (isSetupFinalised) {
    return <PageLoader />;
  }

  const currentKey = STEP_KEYS[currentStepIndex];
  const isFirst = currentStepIndex === 0;
  const isLast = currentStepIndex === STEP_KEYS.length - 1;

  const goBack = (): void => {
    if (currentStepIndex > 0) setCurrentStepIndex(currentStepIndex - 1);
  };

  const advanceAfterComplete = (): void => {
    if (isLast) {
      upsertConfig.mutate(
        { setupCompletedAt: String(Math.floor(Date.now() / 1000)) },
        {
          onSuccess: () => {
            navigate({ to: '/', replace: true });
          },
        }
      );
    } else {
      setCurrentStepIndex(currentStepIndex + 1);
    }
  };

  const handleStepComplete = (stepKey: string, data?: Record<string, unknown>): void => {
    completeStep.mutate(
      { stepKey, data },
      {
        onSuccess: () => {
          advanceAfterComplete();
        },
      }
    );
  };

  const handleCompanyProfile = (data: CompanyProfileFormData): void => {
    handleStepComplete('company_profile', data as Record<string, unknown>);
  };

  const handleWorkSchedule = (data: WorkScheduleFormData): void => {
    handleStepComplete('work_schedule', data as Record<string, unknown>);
  };

  const handleOrgStructure = (departments: OrgStructureFormData['departments']): void => {
    handleStepComplete('org_structure', { departments });
  };

  const handleLocations = (data: LocationsFormData): void => {
    handleStepComplete('locations', data as Record<string, unknown>);
  };

  const handleRolesPermissions = (): void => {
    handleStepComplete('roles_permissions');
  };

  const handlePayrollConfig = (data: PayrollConfigFormData): void => {
    handleStepComplete('payroll', data as Record<string, unknown>);
  };

  const handleNotifications = (data: NotificationsFormData): void => {
    handleStepComplete('notifications', data as Record<string, unknown>);
  };

  const handleEmployees = (): void => {
    handleStepComplete('employees');
  };

  const isSubmitting = completeStep.isPending || upsertConfig.isPending;

  if (showNewEmployee) {
    const NewEmployee = React.lazy(() => import('../employees/NewEmployee'));
    return (
      <React.Suspense fallback={<PageLoader />}>
        <NewEmployee onCancel={() => setShowNewEmployee(false)} isWizardMode />
      </React.Suspense>
    );
  }

  return (
    <>
      <HelmetContainer title="Setup wizard · rtylr" />
      <div>
        <div className="mb-6 text-center">
          <h1 className="text-xl font-bold" style={{ color: 'var(--hr-text)' }}>
            {t('setup.title')}
          </h1>
          <p className="mt-1 text-sm" style={{ color: 'var(--hr-text-muted)' }}>
            {t('setup.subtitle')}
          </p>
        </div>

        <WizardProgress
          steps={steps}
          currentStepIndex={currentStepIndex}
          onStepClick={setCurrentStepIndex}
        />

        {currentKey === 'company_profile' && (
          <CompanyProfileStep
            onNext={handleCompanyProfile}
            isFirst={isFirst}
            isSubmitting={isSubmitting}
          />
        )}
        {currentKey === 'work_schedule' && (
          <WorkScheduleStep
            onNext={handleWorkSchedule}
            onBack={goBack}
            isSubmitting={isSubmitting}
          />
        )}
        {currentKey === 'org_structure' && (
          <OrgStructureStep
            onNext={handleOrgStructure}
            onBack={goBack}
            isSubmitting={isSubmitting}
          />
        )}
        {currentKey === 'locations' && (
          <LocationsStep onNext={handleLocations} onBack={goBack} isSubmitting={isSubmitting} />
        )}
        {currentKey === 'roles_permissions' && (
          <RolesPermissionsStep
            onNext={handleRolesPermissions}
            onBack={goBack}
            isSubmitting={isSubmitting}
          />
        )}
        {currentKey === 'payroll' && (
          <PayrollConfigStep
            onNext={handlePayrollConfig}
            onBack={goBack}
            isSubmitting={isSubmitting}
          />
        )}
        {currentKey === 'notifications' && (
          <NotificationsStep
            onNext={handleNotifications}
            onBack={goBack}
            isSubmitting={isSubmitting}
          />
        )}
        {currentKey === 'employees' && (
          <EmployeesStep
            onNext={handleEmployees}
            onBack={goBack}
            onAddEmployee={() => setShowNewEmployee(true)}
            isLast={isLast}
            isSubmitting={isSubmitting}
          />
        )}
      </div>
    </>
  );
};

export default SetupWizard;
