import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import {
  useGetPayrollConfig,
  useUpsertPayrollConfig,
} from '../../../shared/hooks/queries/usePayroll';
import type { PayrollConfigData } from '../../../shared/hooks/queries/usePayroll';
import { payrollConfigSchema } from '../../../shared/schemas/payroll';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Select,
  Checkbox,
  Skeleton,
} from '../../../shared/components/ui';

type FormData = z.infer<typeof payrollConfigSchema>;

const GROSS_MODELS = ['hourly', 'salary', 'both'] as const;
const PAY_PERIODS = ['weekly', 'biweekly', 'monthly'] as const;

const PayrollConfigSettings = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading } = useGetPayrollConfig();
  const upsertConfig = useUpsertPayrollConfig();
  const [form, setForm] = useState<FormData | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (data && !form) {
      setForm({
        grossModel: data.grossModel,
        period: data.period,
        overtimeEnabled: data.overtimeEnabled,
        deductionsEnabled: data.deductionsEnabled,
        bonusesEnabled: data.bonusesEnabled,
      });
    }
  }, [data, form]);

  if (isLoading || !form) {
    return <Skeleton height={320} />;
  }

  const handleSave = (): void => {
    const result = payrollConfigSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.issues.forEach((issue) => {
        const key = issue.path.join('.');
        fieldErrors[key] = issue.message;
      });
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    upsertConfig.mutate(result.data);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 640 }}>
      <SectionHeader title={t('payroll.config.title')} subtitle={t('payroll.config.subtitle')} />

      <Card>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* Gross Model */}
          <div>
            <Label htmlFor="grossModel">{t('payroll.config.grossModel')}</Label>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {GROSS_MODELS.map((model) => (
                <label
                  key={model}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    padding: '10px 16px',
                    borderRadius: 8,
                    border: `1px solid ${form.grossModel === model ? '#8b5cf6' : 'var(--hr-border)'}`,
                    background: form.grossModel === model ? 'rgba(139,92,246,0.08)' : 'transparent',
                    cursor: 'pointer',
                    fontSize: 13,
                    fontWeight: 500,
                    color: form.grossModel === model ? '#8b5cf6' : 'var(--hr-text)',
                    minHeight: 44,
                  }}
                >
                  <input
                    type="radio"
                    name="grossModel"
                    value={model}
                    checked={form.grossModel === model}
                    onChange={() => setForm({ ...form, grossModel: model })}
                    style={{ display: 'none' }}
                  />
                  {t(`payroll.config.grossModels.${model}`)}
                </label>
              ))}
            </div>
            {errors.grossModel && (
              <span style={{ fontSize: 11, color: '#ef4444', marginTop: 4 }}>
                {errors.grossModel}
              </span>
            )}
          </div>

          {/* Pay Period */}
          <div>
            <Label htmlFor="period">{t('payroll.config.period')}</Label>
            <Select
              id="period"
              value={form.period}
              onChange={(e) => setForm({ ...form, period: e.target.value as FormData['period'] })}
              style={{ maxWidth: 280 }}
            >
              {PAY_PERIODS.map((p) => (
                <option key={p} value={p}>
                  {t(`payroll.config.periods.${p}`)}
                </option>
              ))}
            </Select>
            {errors.period && (
              <span style={{ fontSize: 11, color: '#ef4444', marginTop: 4 }}>{errors.period}</span>
            )}
          </div>

          {/* Toggles */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Checkbox
              checked={form.overtimeEnabled}
              onChange={(v) => setForm({ ...form, overtimeEnabled: v })}
              label={t('payroll.config.enableOvertime')}
            />
            <Checkbox
              checked={form.deductionsEnabled}
              onChange={(v) => setForm({ ...form, deductionsEnabled: v })}
              label={t('payroll.config.enableDeductions')}
            />
            <Checkbox
              checked={form.bonusesEnabled}
              onChange={(v) => setForm({ ...form, bonusesEnabled: v })}
              label={t('payroll.config.enableBonuses')}
            />
          </div>
        </div>

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 24 }}>
          <Button variant="primary" onClick={handleSave} disabled={upsertConfig.isPending}>
            {upsertConfig.isPending ? t('common.loading') : t('common.save')}
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default PayrollConfigSettings;
