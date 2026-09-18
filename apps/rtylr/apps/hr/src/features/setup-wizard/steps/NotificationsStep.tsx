import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  notificationsSchema,
  type NotificationsFormData,
} from '../../../shared/schemas/wizardSchemas';
import StepContainer from '../components/StepContainer';

interface NotificationsStepProps {
  onNext: (data: NotificationsFormData) => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const NotificationsStep = ({
  onNext,
  onBack,
  isSubmitting,
}: NotificationsStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [form, setForm] = useState<NotificationsFormData>({
    emailEnabled: true,
    smsEnabled: false,
  });

  const handleSubmit = (): void => {
    const result = notificationsSchema.safeParse(form);
    if (!result.success) return;
    onNext(result.data);
  };

  return (
    <StepContainer
      title={t('setup.notifications.title')}
      description={t('setup.notifications.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <label
          className="flex items-center gap-3 cursor-pointer rounded-lg p-4"
          style={{ background: 'var(--hr-hover)', border: '1px solid var(--hr-border)' }}
        >
          <input
            type="checkbox"
            checked={form.emailEnabled}
            onChange={(e) => setForm((f) => ({ ...f, emailEnabled: e.target.checked }))}
            className="h-4 w-4 rounded"
          />
          <span className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.notifications.emailEnabled')}
          </span>
        </label>

        <label
          className="flex items-center gap-3 cursor-pointer rounded-lg p-4"
          style={{ background: 'var(--hr-hover)', border: '1px solid var(--hr-border)' }}
        >
          <input
            type="checkbox"
            checked={form.smsEnabled}
            onChange={(e) => setForm((f) => ({ ...f, smsEnabled: e.target.checked }))}
            className="h-4 w-4 rounded"
          />
          <span className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.notifications.smsEnabled')}
          </span>
        </label>

        <p className="text-xs" style={{ color: 'var(--hr-text-dim)' }}>
          {t('setup.notifications.channelNote')}
        </p>
      </div>
    </StepContainer>
  );
};

export default NotificationsStep;
