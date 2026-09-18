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

// The code is what gets stored; only the name is translated.
const SETTINGS_CURRENCIES = ['USD', 'EUR', 'GBP'] as const;

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
    update.mutate(form, { onSuccess: () => toast.success(t('setup.hrSettings.saved')) });
  };

  return (
    <>
      <HelmetContainer title={t('setup.hrSettings.metaTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 720 }}>
        <SectionHeader
          title={t('setup.hrSettings.pageTitle')}
          subtitle={t('setup.hrSettings.subtitle')}
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
                <option value="Monday">{t('setup.hrSettings.monday')}</option>
                <option value="Sunday">{t('setup.hrSettings.sunday')}</option>
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
                {SETTINGS_CURRENCIES.map((code) => (
                  <option key={code} value={code}>
                    {code}, {t(`common.currencies.${code}`)}
                  </option>
                ))}
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
              {t('setup.hrSettings.save')}
            </Button>
          </div>
        </Card>
      </div>
    </>
  );
};

export default Settings;
