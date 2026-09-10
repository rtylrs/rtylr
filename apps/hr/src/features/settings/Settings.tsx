import React, { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Settings as SettingsIcon } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
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
  useIsMobile,
} from '../../shared/components/ui';
import { HrSettings } from '../../shared/services/hrExtras';

const Settings: React.FC = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);
  const { data, isLoading, isError, refetch } = useHrSettings();
  const update = useUpdateHrSettings();

  const [form, setForm] = useState<HrSettings | null>(null);

  useEffect(() => {
    if (data && !form) setForm(data);
  }, [data, form]);

  if (isError && !form) {
    return (
      <div role="alert">
        <EmptyState
          icon={SettingsIcon}
          title={t('errors.hrSettingsLoadFailed')}
          description={t('errors.generic')}
          action={{ label: t('common.retry'), onClick: () => void refetch() }}
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
              <Label>{t('setup.hrSettings.workWeekStart')}</Label>
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
              <Label>{t('setup.hrSettings.weeklyHours')}</Label>
              <Input
                type="number"
                value={form.weeklyHours}
                onChange={(e) => setForm({ ...form, weeklyHours: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>{t('setup.hrSettings.overtimeAfter')}</Label>
              <Input
                type="number"
                value={form.overtimeAfterHours}
                onChange={(e) => setForm({ ...form, overtimeAfterHours: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>{t('setup.hrSettings.ptoPolicy')}</Label>
              <Input
                type="number"
                value={form.ptoPolicyDays}
                onChange={(e) => setForm({ ...form, ptoPolicyDays: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>{t('setup.hrSettings.currency')}</Label>
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
              <Label>{t('setup.hrSettings.industry')}</Label>
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
