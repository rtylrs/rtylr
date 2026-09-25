import React, { useEffect, useState } from 'react';
import { pageTitle } from '../../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { z } from 'zod';
import {
  useGetPayrollConfig,
  useUpsertPayrollConfig,
} from '../../../shared/hooks/queries/usePayroll';
import { payrollConfigSchema } from '../../../shared/schemas/payroll';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Select,
  Input,
  Checkbox,
  Skeleton,
} from '../../../shared/components/ui';

type FormData = z.infer<typeof payrollConfigSchema>;

const PAY_PERIODS = ['weekly', 'biweekly', 'monthly'] as const;

const PayrollConfigSettings = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading } = useGetPayrollConfig();
  const upsertConfig = useUpsertPayrollConfig();
  const [form, setForm] = useState<FormData | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [configNotice, setConfigNotice] = useState<string | null>(null);

  useEffect(() => {
    if (data && !form) {
      setForm({
        period: data.period,
        overtimeEnabled: data.overtimeEnabled,
        payDayOfMonth: data.payDayOfMonth,
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
      result.error.issues.forEach((issue: z.ZodIssue) => {
        const key = issue.path.join('.');
        fieldErrors[key] = issue.message;
      });
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    upsertConfig.mutate(result.data, {
      onSuccess: () => {
        setConfigNotice(null);
        toast.success(t('payroll.config.saved'));
      },
      onError: (error: unknown) => {
        const status = (error as { status?: number })?.status;
        if (status === 422) {
          setConfigNotice(t('payroll.config.notConfiguredNotice'));
        } else {
          toast.error(t('errors.generic'));
        }
      },
    });
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('common.pageTitles.payrollConfiguration'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 640 }}>
        <SectionHeader title={t('payroll.config.title')} subtitle={t('payroll.config.subtitle')} />

        <Card>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
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
                <span style={{ fontSize: 11, color: 'var(--hr-error)', marginTop: 4 }}>
                  {errors.period}
                </span>
              )}
            </div>

            <div>
              <Label htmlFor="payDayOfMonth">{t('payroll.config.payDayOfMonth')}</Label>
              <Input
                id="payDayOfMonth"
                type="number"
                min={1}
                max={28}
                value={form.payDayOfMonth}
                onChange={(e) => setForm({ ...form, payDayOfMonth: Number(e.target.value) })}
                style={{ maxWidth: 160 }}
              />
              {errors.payDayOfMonth && (
                <span style={{ fontSize: 11, color: 'var(--hr-error)', marginTop: 4 }}>
                  {errors.payDayOfMonth}
                </span>
              )}
            </div>

            {/* Toggles */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <Checkbox
                checked={form.overtimeEnabled}
                onChange={(v) => setForm({ ...form, overtimeEnabled: v })}
                label={t('payroll.config.enableOvertime')}
              />
            </div>
          </div>

          {configNotice && (
            <p style={{ color: 'var(--hr-status-draft)', marginBottom: 8, fontSize: 13 }}>
              {configNotice}
            </p>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 24 }}>
            <Button variant="primary" onClick={handleSave} disabled={upsertConfig.isPending}>
              {upsertConfig.isPending ? t('common.loading') : t('common.save')}
            </Button>
          </div>
        </Card>
      </div>
    </>
  );
};

export default PayrollConfigSettings;
