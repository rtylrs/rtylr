import React from 'react';
import { useTranslation } from 'react-i18next';
import StepContainer from '../components/StepContainer';

const DEFAULT_ROLE_KEYS = [
  'employee',
  'manager',
  'hrAdmin',
  'payrollManager',
  'executive',
] as const;

interface RolesPermissionsStepProps {
  onNext: () => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const RolesPermissionsStep = ({
  onNext,
  onBack,
  isSubmitting,
}: RolesPermissionsStepProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');

  const handleNext = (): void => {
    onNext();
  };

  return (
    <StepContainer
      title={t('hrSetup.rolesPermissions.title')}
      description={t('hrSetup.rolesPermissions.description')}
      onBack={onBack}
      onNext={handleNext}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-3">
        {DEFAULT_ROLE_KEYS.map((key) => (
          <div
            key={key}
            className="rounded-lg px-4 py-3"
            style={{ border: '1px solid var(--hr-border)', background: 'var(--hr-card)' }}
          >
            <p className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
              {t(`hrSetup.rolesPermissions.defaultRoles.${key}.name`)}
            </p>
            <p className="mt-0.5 text-xs" style={{ color: 'var(--hr-text-dim)' }}>
              {t(`hrSetup.rolesPermissions.defaultRoles.${key}.description`)}
            </p>
          </div>
        ))}
        <p className="mt-1 text-xs" style={{ color: 'var(--hr-text-muted)' }}>
          {t('hrSetup.rolesPermissions.previewNotice')}
        </p>
      </div>
    </StepContainer>
  );
};

export default RolesPermissionsStep;
