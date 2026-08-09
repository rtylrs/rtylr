import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  payrollConfigSchema,
  type PayrollConfigFormData,
} from '../../../shared/schemas/wizardSchemas';
import StepContainer from '../components/StepContainer';

interface PayrollConfigStepProps {
  onNext: (data: PayrollConfigFormData) => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const PayrollConfigStep = ({
  onNext,
  onBack,
  isSubmitting,
}: PayrollConfigStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [form, setForm] = useState<PayrollConfigFormData>({
    payrollGrossModel: 'salary',
    payrollPeriod: 'monthly',
    enableOvertime: true,
    enableDeductions: true,
    enableBonuses: true,
  });

  const handleSubmit = (): void => {
    const result = payrollConfigSchema.safeParse(form);
    if (!result.success) return;
    onNext(result.data);
  };

  const grossModels = [
    { value: 'hourly' as const, label: t('setup.payrollConfig.hourly') },
    { value: 'salary' as const, label: t('setup.payrollConfig.salary') },
    { value: 'both' as const, label: t('setup.payrollConfig.both') },
  ];

  const periods = [
    { value: 'weekly' as const, label: t('setup.payrollConfig.weekly') },
    { value: 'biweekly' as const, label: t('setup.payrollConfig.biweekly') },
    { value: 'monthly' as const, label: t('setup.payrollConfig.monthly') },
  ];

  return (
    <StepContainer
      title={t('setup.payrollConfig.title')}
      description={t('setup.payrollConfig.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-6">
        <div>
          <label className="mb-2 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.payrollConfig.grossModel')}
          </label>
          <div className="flex gap-2">
            {grossModels.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => setForm((f) => ({ ...f, payrollGrossModel: opt.value }))}
                className="flex-1 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors"
                style={{
                  background:
                    form.payrollGrossModel === opt.value ? 'var(--accent)' : 'var(--hr-hover)',
                  color: form.payrollGrossModel === opt.value ? '#fff' : 'var(--hr-text-muted)',
                  border: `1px solid ${form.payrollGrossModel === opt.value ? 'var(--accent)' : 'var(--hr-border)'}`,
                }}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('setup.payrollConfig.period')}
          </label>
          <div className="flex gap-2">
            {periods.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => setForm((f) => ({ ...f, payrollPeriod: opt.value }))}
                className="flex-1 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors"
                style={{
                  background:
                    form.payrollPeriod === opt.value ? 'var(--accent)' : 'var(--hr-hover)',
                  color: form.payrollPeriod === opt.value ? '#fff' : 'var(--hr-text-muted)',
                  border: `1px solid ${form.payrollPeriod === opt.value ? 'var(--accent)' : 'var(--hr-border)'}`,
                }}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        <div className="flex flex-col gap-3">
          {[
            { key: 'enableOvertime' as const, label: t('setup.payrollConfig.enableOvertime') },
            { key: 'enableDeductions' as const, label: t('setup.payrollConfig.enableDeductions') },
            { key: 'enableBonuses' as const, label: t('setup.payrollConfig.enableBonuses') },
          ].map(({ key, label }) => (
            <label key={key} className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={form[key]}
                onChange={(e) => setForm((f) => ({ ...f, [key]: e.target.checked }))}
                className="h-4 w-4 rounded"
              />
              <span className="text-sm" style={{ color: 'var(--hr-text)' }}>
                {label}
              </span>
            </label>
          ))}
        </div>
      </div>
    </StepContainer>
  );
};

export default PayrollConfigStep;
