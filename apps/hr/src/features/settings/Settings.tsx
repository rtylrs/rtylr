import React, { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { INDUSTRIES } from '@voxire/config';
import { useHrSettings, useUpdateHrSettings } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Input,
  Select,
  Skeleton,
  EmptyState,
  useIsMobile,
} from '../../shared/components/ui';
import { HrSettings } from '../../shared/services/hrExtras';

const Settings: React.FC = () => {
  const isMobile = useIsMobile(900);
  const { data, isLoading, isError, isFetching, refetch } = useHrSettings();
  const update = useUpdateHrSettings();

  const [form, setForm] = useState<HrSettings | null>(null);

  useEffect(() => {
    if (data && !form) setForm(data);
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
    update.mutate(form, { onSuccess: () => toast.success('Settings saved') });
  };

  return (
    <>
      <HelmetContainer title="HR settings · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 720 }}>
        <SectionHeader
          title="HR Settings"
          subtitle="Work week, overtime, PTO policy, and currency"
        />
        <Card>
          <div
            style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16 }}
          >
            <div>
              <Label>Work week starts on</Label>
              <Select
                value={form.workWeekStart}
                onChange={(e) =>
                  setForm({ ...form, workWeekStart: e.target.value as HrSettings['workWeekStart'] })
                }
              >
                <option value="Monday">Monday</option>
                <option value="Sunday">Sunday</option>
              </Select>
            </div>
            <div>
              <Label>Weekly hours</Label>
              <Input
                type="number"
                value={form.weeklyHours}
                onChange={(e) => setForm({ ...form, weeklyHours: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>Overtime after (hrs/week)</Label>
              <Input
                type="number"
                value={form.overtimeAfterHours}
                onChange={(e) => setForm({ ...form, overtimeAfterHours: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>PTO policy (days/year)</Label>
              <Input
                type="number"
                value={form.ptoPolicyDays}
                onChange={(e) => setForm({ ...form, ptoPolicyDays: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>Currency</Label>
              <Select
                value={form.currency}
                onChange={(e) =>
                  setForm({ ...form, currency: e.target.value as HrSettings['currency'] })
                }
              >
                <option value="USD">USD - US Dollar</option>
                <option value="EUR">EUR - Euro</option>
                <option value="GBP">GBP - Pound</option>
              </Select>
            </div>
            <div>
              <Label>Industry</Label>
              <Select
                value={form.industry ?? ''}
                onChange={(e) => setForm({ ...form, industry: e.target.value })}
              >
                <option value="">-</option>
                {INDUSTRIES.map((i) => (
                  <option key={i.value} value={i.value}>
                    {i.label}
                  </option>
                ))}
              </Select>
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
