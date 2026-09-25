import React from 'react';
import { useTranslation } from 'react-i18next';
import StepContainer from '../components/StepContainer';

interface NotificationsStepProps {
  onNext: () => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const NotificationsStep = ({
  onNext,
  onBack,
  isSubmitting,
}: NotificationsStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const handleNext = (): void => {
    onNext();
  };

  return (
    <StepContainer
      title={t('setup.notifications.title')}
      description={t('setup.notifications.description')}
      onBack={onBack}
      onNext={handleNext}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <p className="text-sm" style={{ color: 'var(--hr-text-muted)' }}>
          {t('setup.notifications.channelNote')}
        </p>
      </div>
    </StepContainer>
  );
};

export default NotificationsStep;
