import React from 'react';
import { useTranslation } from 'react-i18next';
import { Check } from 'lucide-react';
import type { SetupWizardStepPayload } from '@voxire/voxie/hr/setup/v1/types';

const STEP_KEYS = [
  'company_profile',
  'work_schedule',
  'org_structure',
  'locations',
  'roles_permissions',
  'payroll',
  'notifications',
  'employees',
] as const;

const STEP_I18N_MAP: Record<string, string> = {
  company_profile: 'hrSetup.steps.companyProfile',
  work_schedule: 'hrSetup.steps.workSchedule',
  org_structure: 'hrSetup.steps.orgStructure',
  locations: 'hrSetup.steps.locations',
  roles_permissions: 'hrSetup.steps.rolesPermissions',
  payroll: 'hrSetup.steps.payrollConfig',
  notifications: 'hrSetup.steps.notifications',
  employees: 'hrSetup.steps.employees',
};

interface WizardProgressProps {
  steps: SetupWizardStepPayload[];
  currentStepIndex: number;
  onStepClick: (index: number) => void;
}

const WizardProgress = ({
  steps,
  currentStepIndex,
  onStepClick,
}: WizardProgressProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');

  const isStepCompleted = (stepKey: string): boolean => {
    return steps.some((s) => s.stepKey === stepKey && s.completedAt != null);
  };

  return (
    <nav className="mb-8" aria-label={t('hrSetup.progress')}>
      <ol className="flex flex-wrap gap-2">
        {STEP_KEYS.map((key, index) => {
          const completed = isStepCompleted(key);
          const isCurrent = index === currentStepIndex;

          return (
            <li key={key} className="flex items-center gap-2">
              {/* Completed steps are not editable yet. Re-entering one would
                  remount its create-only form and duplicate persisted setup data. */}
              <button
                type="button"
                onClick={() => {
                  if (isCurrent) onStepClick(index);
                }}
                disabled={!isCurrent}
                className="flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-medium transition-colors"
                style={{
                  background: isCurrent
                    ? 'var(--accent)'
                    : completed
                      ? 'rgba(34,197,94,0.15)'
                      : 'var(--hr-hover)',
                  color: isCurrent
                    ? 'var(--hr-on-accent)'
                    : completed
                      ? '#22c55e'
                      : 'var(--hr-text-muted)',
                  border: isCurrent ? '1px solid var(--accent)' : '1px solid var(--hr-border)',
                  cursor: isCurrent ? 'pointer' : 'not-allowed',
                  opacity: isCurrent || completed ? 1 : 0.5,
                }}
                aria-current={isCurrent ? 'step' : undefined}
              >
                {completed ? (
                  <Check size={12} />
                ) : (
                  <span
                    className="flex h-4 w-4 items-center justify-center rounded-full text-[10px] font-bold"
                    style={{
                      background: isCurrent ? 'rgba(255,255,255,0.2)' : 'var(--hr-border)',
                    }}
                  >
                    {index + 1}
                  </span>
                )}
                <span className="hidden sm:inline">{t(STEP_I18N_MAP[key])}</span>
              </button>
              {index < STEP_KEYS.length - 1 && (
                <div
                  className="hidden h-px w-4 sm:block"
                  style={{ background: 'var(--hr-border)' }}
                />
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
};

export { STEP_KEYS };
export default WizardProgress;
