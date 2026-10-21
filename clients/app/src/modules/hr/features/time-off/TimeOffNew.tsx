import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useStaff } from '../staff/useStaff';
import { useSubmitTimeOff } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Input,
  Textarea,
  Select,
  Breadcrumbs,
  useIsMobile,
} from '../../shared/components/ui';

const TYPES = ['Vacation', 'Sick', 'Personal', 'Bereavement', 'Unpaid'] as const;

const TimeOffNew: React.FC = () => {
  const { t } = useTranslation('hrCommon');
  const isMobile = useIsMobile(900);
  const nav = useNavigate();
  const { data: staff } = useStaff();
  const submit = useSubmitTimeOff();

  const [staffId, setStaffId] = useState<number>(0);
  const [type, setType] = useState<(typeof TYPES)[number]>('Vacation');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [reason, setReason] = useState('');

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const who = (staff ?? []).find((s) => s.id === staffId);
    if (!who) {
      toast.error(t('hrTimeOff.errorSelectEmployee'));
      return;
    }
    if (!startDate || !endDate) {
      toast.error(t('hrTimeOff.errorSelectRange'));
      return;
    }
    if (endDate < startDate) {
      toast.error(t('hrTimeOff.errorEndBeforeStart'));
      return;
    }
    submit.mutate(
      { staffId, staffName: who.name, type, startDate, endDate, reason },
      {
        onSuccess: () => {
          toast.success(t('hrTimeOff.submitted'));
          nav({ to: '/hr/time-off' });
        },
        onError: () => toast.error(t('hrTimeOff.submitFailed')),
      }
    );
  };

  return (
    <>
      <HelmetContainer title={t('hrTimeOff.pageTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 600 }}>
        <Breadcrumbs
          items={[
            { label: t('hrNav.timeOff'), href: '/hr/time-off' },
            { label: t('hrTimeOff.page.newRequest') },
          ]}
        />
        <SectionHeader title={t('hrTimeOff.formTitle')} subtitle={t('hrTimeOff.formSubtitle')} />
        <Card>
          <form onSubmit={onSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div>
              <Label>{t('hrLeave.employee')}</Label>
              <Select value={staffId} onChange={(e) => setStaffId(Number(e.target.value))}>
                <option value={0}>{t('hrTimeOff.selectEmployee')}</option>
                {(staff ?? []).map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.role})
                  </option>
                ))}
              </Select>
            </div>
            <div>
              <Label>{t('hrTimeOff.type')}</Label>
              <Select
                value={type}
                onChange={(e) => setType(e.target.value as (typeof TYPES)[number])}
              >
                {TYPES.map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </Select>
            </div>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr',
                gap: 10,
              }}
            >
              <div>
                <Label>{t('hrTimeOff.startDate')}</Label>
                <Input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                />
              </div>
              <div>
                <Label>{t('hrTimeOff.endDate')}</Label>
                <Input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
              </div>
            </div>
            <div>
              <Label>{t('hrTimeOff.reason')}</Label>
              <Textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder={t('hrTimeOff.reasonPlaceholder')}
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', flexWrap: 'wrap' }}>
              <Link to="/hr/time-off">
                <Button type="button" variant="ghost">
                  {t('hrCommon.cancel')}
                </Button>
              </Link>
              <Button type="submit" variant="primary" disabled={submit.isPending}>
                {submit.isPending ? 'Submitting...' : 'Submit request'}
              </Button>
            </div>
          </form>
        </Card>
      </div>
    </>
  );
};

export default TimeOffNew;
