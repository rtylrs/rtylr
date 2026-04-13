import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
  const { data, isLoading } = useGetSetupStatus();
  const completeStep = useCompleteSetupStep();
  const upsertConfig = useUpsertHRConfig();
  const { currentStepIndex, setCurrentStepIndex } = useHRSetupStore();
  const [showNewEmployee, setShowNewEmployee] = useState(false);

  const steps = data?.steps ?? [];

  useEffect(() => {
    if (data && !data.isComplete) {
      const firstIncomplete = STEP_KEYS.findIndex(
        (key) => !steps.some((s) => s.stepKey === key && s.completedAt != null)
      );
      if (firstIncomplete >= 0) {
        setCurrentStepIndex(firstIncomplete);
      }
    }
  }, [data, steps, setCurrentStepIndex]);

  if (isLoading) {
    return <PageLoader />;
  }

  if (data?.isComplete) {
    navigate('/', { replace: true });
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
        { setupCompletedAt: Math.floor(Date.now() / 1000) },
        {
          onSuccess: () => {
            navigate('/', { replace: true });
          },
        }
      );
    } else {
      setCurrentStepIndex(currentStepIndex + 1);
    }
  };

  const handleStepComplete = (stepKey: string): void => {
    completeStep.mutate(stepKey, {
      onSuccess: () => {
        advanceAfterComplete();
      },
    });
  };

  const handleCompanyProfile = (_data: CompanyProfileFormData): void => {
    handleStepComplete('company_profile');
  };

  const handleWorkSchedule = (_data: WorkScheduleFormData): void => {
    handleStepComplete('work_schedule');
  };

  const handleOrgStructure = (): void => {
    handleStepComplete('org_structure');
  };

  const handleLocations = (): void => {
    handleStepComplete('locations');
  };

  const handleRolesPermissions = (): void => {
    handleStepComplete('roles_permissions');
  };

  const handlePayrollConfig = (_data: PayrollConfigFormData): void => {
    handleStepComplete('payroll');
  };

  const handleNotifications = (_data: NotificationsFormData): void => {
    handleStepComplete('notifications');
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
        <WorkScheduleStep onNext={handleWorkSchedule} onBack={goBack} isSubmitting={isSubmitting} />
      )}
      {currentKey === 'org_structure' && (
        <OrgStructureStep onNext={handleOrgStructure} onBack={goBack} isSubmitting={isSubmitting} />
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
  );
};

export default SetupWizard;
