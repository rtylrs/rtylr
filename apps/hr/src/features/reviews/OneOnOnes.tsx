import React, { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { MessageSquare, Plus, Calendar, Clock } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useOneOnOnes, useCreateOneOnOne } from '../../shared/hooks/useHrExtras';
import { useStaff } from '../staff/useStaff';
import {
  Card,
  SectionHeader,
  Button,
  Skeleton,
  Pill,
  Modal,
  Label,
  Input,
  Select,
  Textarea,
  useIsMobile,
} from '../../shared/components/ui';

const OneOnOnes: React.FC = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);
  const { data, isLoading } = useOneOnOnes();
  const { data: staff = [] } = useStaff();
  const create = useCreateOneOnOne();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState({
    managerId: 0,
    reportId: 0,
    scheduledAt: '',
    duration: 30,
    topics: '',
  });

  const onCreate = () => {
    const manager = staff.find((s) => s.id === form.managerId);
    const report = staff.find((s) => s.id === form.reportId);
    if (!manager || !report || !form.scheduledAt) {
      toast.error(t('reviews.oneOnOnes.errorCompleteAll'));
      return;
    }
    create.mutate(
      {
        managerId: manager.id,
        manager: manager.name,
        reportId: report.id,
        report: report.name,
        scheduledAt: new Date(form.scheduledAt).toISOString(),
        duration: form.duration,
        status: 'Scheduled',
        topics: form.topics
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
      },
      {
        onSuccess: () => {
          toast.success('1-on-1 scheduled');
          setOpen(false);
          setForm({ managerId: 0, reportId: 0, scheduledAt: '', duration: 30, topics: '' });
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title={t('reviews.oneOnOnes.pageTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('reviews.oneOnOnes.title')}
          subtitle={t('reviews.oneOnOnes.subtitle')}
          right={
            <Button variant="primary" icon={<Plus size={13} />} onClick={() => setOpen(true)}>
              Schedule 1-on-1
            </Button>
          }
        />
        {isLoading ? (
          <Skeleton height={200} />
        ) : !data || data.length === 0 ? (
          <EmptyState
            icon={MessageSquare}
            title={t('reviews.oneOnOnes.emptyTitle')}
            description={t('reviews.oneOnOnes.emptyDescription')}
            action={{ label: t('reviews.oneOnOnes.schedule'), onClick: () => setOpen(true) }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {data.map((o) => (
              <Card key={o.id}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: isMobile ? 'flex-start' : 'center',
                    flexDirection: isMobile ? 'column' : 'row',
                    gap: 12,
                    marginBottom: 8,
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {o.manager}{' '}
                      <span style={{ color: 'var(--hr-text-muted)', fontWeight: 400 }}>&harr;</span>{' '}
                      {o.report}
                    </div>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        marginTop: 4,
                        fontSize: 11,
                        color: 'var(--hr-text-muted)',
                      }}
                    >
                      <Calendar size={11} /> {new Date(o.scheduledAt).toLocaleString()}
                      <span>•</span>
                      <Clock size={11} /> {o.duration} min
                    </div>
                  </div>
                  <Pill
                    color={
                      o.status === 'Completed'
                        ? '#10b981'
                        : o.status === 'Cancelled'
                          ? '#ef4444'
                          : '#3b82f6'
                    }
                  >
                    {o.status}
                  </Pill>
                </div>
                {o.topics.length > 0 && (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginTop: 6 }}>
                    {o.topics.map((t, i) => (
                      <Pill key={i}>{t}</Pill>
                    ))}
                  </div>
                )}
                {o.notes && (
                  <p
                    style={{
                      margin: '8px 0 0',
                      fontSize: 11,
                      color: 'var(--hr-text-muted)',
                      fontStyle: 'italic',
                    }}
                  >
                    "{o.notes}"
                  </p>
                )}
              </Card>
            ))}
          </div>
        )}

        <Modal open={open} onClose={() => setOpen(false)} title={t('reviews.oneOnOnes.schedule')}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr',
                gap: 10,
              }}
            >
              <div>
                <Label>{t('reviews.oneOnOnes.manager')}</Label>
                <Select
                  value={form.managerId}
                  onChange={(e) => setForm({ ...form, managerId: Number(e.target.value) })}
                >
                  <option value={0}>Select</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <Label>{t('reviews.oneOnOnes.report')}</Label>
                <Select
                  value={form.reportId}
                  onChange={(e) => setForm({ ...form, reportId: Number(e.target.value) })}
                >
                  <option value={0}>Select</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </Select>
              </div>
            </div>
            <div>
              <Label>{t('reviews.oneOnOnes.when')}</Label>
              <Input
                type="datetime-local"
                value={form.scheduledAt}
                onChange={(e) => setForm({ ...form, scheduledAt: e.target.value })}
              />
            </div>
            <div>
              <Label>{t('reviews.oneOnOnes.duration')}</Label>
              <Input
                type="number"
                value={form.duration}
                onChange={(e) => setForm({ ...form, duration: Number(e.target.value) })}
              />
            </div>
            <div>
              <Label>{t('reviews.oneOnOnes.topics')}</Label>
              <Textarea
                value={form.topics}
                onChange={(e) => setForm({ ...form, topics: e.target.value })}
                placeholder={t('reviews.oneOnOnes.topicsPlaceholder')}
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <Button variant="ghost" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button variant="primary" onClick={onCreate}>
                Schedule
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </>
  );
};

export default OneOnOnes;
