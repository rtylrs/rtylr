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
  const { t } = useTranslation('hrCommon');
  const [form, setForm] = useState<PayrollConfigFormData>({
    payrollPeriod: 'monthly',
    enableOvertime: true,
    payDayOfMonth: 28,
  });

  const handleSubmit = (): void => {
    const result = payrollConfigSchema.safeParse(form);
    if (!result.success) return;
    onNext(result.data);
  };

  const periods = [
    { value: 'weekly' as const, label: t('hrSetup.payrollConfig.weekly') },
    { value: 'biweekly' as const, label: t('hrSetup.payrollConfig.biweekly') },
    { value: 'monthly' as const, label: t('hrSetup.payrollConfig.monthly') },
  ];

  return (
    <StepContainer
      title={t('hrSetup.payrollConfig.title')}
      description={t('hrSetup.payrollConfig.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-6">
        <div>
          <label className="mb-2 block text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
            {t('hrSetup.payrollConfig.period')}
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
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={form.enableOvertime}
              onChange={(e) => setForm((f) => ({ ...f, enableOvertime: e.target.checked }))}
              className="h-4 w-4 rounded"
            />
            <span className="text-sm" style={{ color: 'var(--hr-text)' }}>
              {t('hrSetup.payrollConfig.enableOvertime')}
            </span>
          </label>
          <label className="flex items-center gap-3 cursor-pointer">
            <span className="text-sm" style={{ color: 'var(--hr-text)' }}>
              {t('hrSetup.payrollConfig.payDayOfMonth')}
            </span>
            <input
              type="number"
              min={1}
              max={28}
              value={form.payDayOfMonth}
              onChange={(e) => setForm((f) => ({ ...f, payDayOfMonth: Number(e.target.value) }))}
              className="w-24 rounded-lg border px-3 py-2 text-sm"
              style={{ borderColor: 'var(--hr-border)', background: 'var(--hr-input-bg)' }}
            />
          </label>
        </div>
        {/* The remaining wizard options are intentionally not shown here: the
            live payroll contract stores only period, pay day, and overtime. */}
      </div>
    </StepContainer>
  );
};

export default PayrollConfigStep;
