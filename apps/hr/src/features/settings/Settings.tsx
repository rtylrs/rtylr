import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useGetHRConfig, useUpsertHRConfig } from '../../shared/hooks/queries/useSetup';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Input,
  Select,
  Checkbox,
  Skeleton,
  EmptyState,
  useIsMobile,
} from '../../shared/components/ui';
import type { HRConfig } from '../../shared/types/hrConfig';

type SettingsForm = Pick<
  HRConfig,
  | 'companyName'
  | 'defaultCurrency'
  | 'defaultTimezone'
  | 'fiscalYearStartMonth'
  | 'overtimeEnabled'
  | 'overtimeMultiplier'
>;

const toSettingsForm = (config: HRConfig): SettingsForm => ({
  companyName: config.companyName ?? '',
  defaultCurrency: config.defaultCurrency ?? 'USD',
  defaultTimezone: config.defaultTimezone ?? 'UTC',
  fiscalYearStartMonth: config.fiscalYearStartMonth ?? 1,
  overtimeEnabled: config.overtimeEnabled ?? false,
  overtimeMultiplier: config.overtimeMultiplier ?? 1.5,
});

const Settings: React.FC = () => {
  const isMobile = useIsMobile(900);
  const { data, isLoading, isError, isFetching, refetch } = useGetHRConfig();
  const update = useUpsertHRConfig();

  const [form, setForm] = useState<SettingsForm | null>(null);

  useEffect(() => {
    if (data && !form) setForm(toSettingsForm(data));
  }, [data, form]);

  if (isError && !form) {
    return (
      <div role="alert">
        <EmptyState
          title="Unable to load HR settings"
          description="Please try again."
          action={
            <Button
              variant="primary"
              onClick={() => {
                void refetch();
              }}
              disabled={isFetching}
            >
              Retry
            </Button>
          }
        />
      </div>
    );
  }

  if (isLoading || !form) {
    return <Skeleton height={280} />;
  }

  const save = () => {
    update.mutate({
      companyName: form.companyName?.trim() ?? '',
      defaultCurrency: form.defaultCurrency,
      defaultTimezone: form.defaultTimezone,
      fiscalYearStartMonth: form.fiscalYearStartMonth,
      overtimeEnabled: form.overtimeEnabled,
      overtimeMultiplier: form.overtimeMultiplier,
    });
  };

  return (
    <>
      <HelmetContainer title="HR settings · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 720 }}>
        <SectionHeader
          title="HR Settings"
          subtitle="Company profile, payroll defaults, and overtime policy"
        />
        <Card>
          <div
            style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16 }}
          >
            <div>
              <Label htmlFor="companyName">Company name</Label>
              <Input
                id="companyName"
                value={form.companyName}
                onChange={(e) => setForm({ ...form, companyName: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="defaultCurrency">Currency</Label>
              <Select
                id="defaultCurrency"
                value={form.defaultCurrency}
                onChange={(e) => setForm({ ...form, defaultCurrency: e.target.value })}
              >
                <option value="USD">USD - US Dollar</option>
                <option value="EUR">EUR - Euro</option>
                <option value="GBP">GBP - Pound</option>
              </Select>
            </div>
            <div>
              <Label htmlFor="defaultTimezone">Timezone</Label>
              <Input
                id="defaultTimezone"
                value={form.defaultTimezone}
                onChange={(e) => setForm({ ...form, defaultTimezone: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="fiscalYearStartMonth">Fiscal year start month</Label>
              <Input
                id="fiscalYearStartMonth"
                type="number"
                min={1}
                max={12}
                value={form.fiscalYearStartMonth}
                onChange={(e) =>
                  setForm({ ...form, fiscalYearStartMonth: Number(e.target.value) })
                }
              />
            </div>
            <div>
              <Label htmlFor="overtimeMultiplier">Overtime multiplier</Label>
              <Input
                id="overtimeMultiplier"
                type="number"
                min={0}
                step={0.05}
                value={form.overtimeMultiplier}
                onChange={(e) =>
                  setForm({ ...form, overtimeMultiplier: Number(e.target.value) })
                }
              />
            </div>
            <div>
              <Label>Overtime</Label>
              <Checkbox
                checked={form.overtimeEnabled ?? false}
                onChange={(overtimeEnabled) => setForm({ ...form, overtimeEnabled })}
                label="Enable overtime tracking"
              />
            </div>
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 18 }}>
            <Button variant="primary" onClick={save}>
              Save settings
            </Button>
          </div>
        </Card>
      </div>
    </>
  );
};

export default Settings;
