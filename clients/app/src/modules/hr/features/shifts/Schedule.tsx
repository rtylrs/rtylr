import React, { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { ChevronLeft, ChevronRight, Send } from 'lucide-react';
import {
  useShiftSchedule,
  usePublishWeek,
  useUpsertShift,
  useDeleteShift,
} from '../../shared/hooks/useHrExtras';
import { useStaff } from '../staff/useStaff';
import {
  Card,
  SectionHeader,
  Button,
  Modal,
  Input,
  Label,
  Select,
  useIsMobile,
} from '../../shared/components/ui';
import ShiftCell from '../../shared/components/ShiftCell';
import { ScheduleShift } from '../../shared/services/hrExtras';

function startOfWeek(date: Date) {
  const d = new Date(date);
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  return new Date(d.setDate(diff));
}

function addDays(d: Date, n: number) {
  const r = new Date(d);
  r.setDate(r.getDate() + n);
  return r;
}

function dayStr(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(
    date.getDate()
  ).padStart(2, '0')}`;
}

const ROLE_COLORS = ['#8b5cf6', '#10b981', '#3b82f6', '#f59e0b', '#ef4444'];

const Schedule: React.FC = () => {
  const { t } = useTranslation('hrCommon');
  const isMobile = useIsMobile(900);
  const [weekStart, setWeekStart] = useState(startOfWeek(new Date()));
  const weekStartStr = dayStr(weekStart);
  const { data: shifts = [] } = useShiftSchedule(weekStartStr);
  const { data: staff = [] } = useStaff();
  const publish = usePublishWeek();
  const upsert = useUpsertShift();
  const del = useDeleteShift();

  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Partial<ScheduleShift> | null>(null);

  const days = useMemo(
    () => Array.from({ length: 7 }, (_, i) => addDays(weekStart, i)),
    [weekStart]
  );

  const grid = useMemo(() => {
    const map = new Map<string, ScheduleShift[]>();
    days.forEach((d) => map.set(dayStr(d), []));
    shifts.forEach((s) => {
      if (map.has(s.day)) map.get(s.day)!.push(s);
    });
    return map;
  }, [days, shifts]);

  const openNew = (day: string) => {
    setEditing({ day, start: '09:00', end: '17:00', color: ROLE_COLORS[0], published: false });
    setModalOpen(true);
  };

  const saveShift = () => {
    if (!editing || !editing.staffId || !editing.day || !editing.start || !editing.end) {
      toast.error(t('hrShifts.schedule.fillAllFields'));
      return;
    }
    const who = staff.find((s) => s.id === editing.staffId);
    if (!who) return;
    upsert.mutate(
      {
        id: editing.id,
        staffId: who.id,
        staffName: who.name,
        role: who.role,
        day: editing.day!,
        start: editing.start!,
        end: editing.end!,
        color: editing.color ?? ROLE_COLORS[0],
        published: editing.published ?? false,
      },
      {
        onSuccess: () => {
          toast.success(t('hrShifts.schedule.shiftSaved'));
          setModalOpen(false);
          setEditing(null);
        },
      }
    );
  };

  const weekLabel = `${weekStart.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })} - ${addDays(weekStart, 6).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}`;

  return (
    <>
      <HelmetContainer title={t('hrShifts.schedule.metaTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('hrShifts.schedule.title')}
          subtitle={t('hrShifts.schedule.subtitle')}
          right={
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
              <Button
                icon={<ChevronLeft size={13} />}
                onClick={() => setWeekStart(addDays(weekStart, -7))}
              />
              <Button onClick={() => setWeekStart(startOfWeek(new Date()))}>Today</Button>
              <Button
                icon={<ChevronRight size={13} />}
                onClick={() => setWeekStart(addDays(weekStart, 7))}
              />
              <Button
                variant="primary"
                icon={<Send size={13} />}
                onClick={() =>
                  publish.mutate(dayStr(weekStart), {
                    onSuccess: () => toast.success(t('hrShifts.schedule.weekPublished')),
                  })
                }
              >
                {t('hrShifts.schedule.publishWeek')}
              </Button>
            </div>
          }
        />
        <p style={{ margin: 0, fontSize: 12, color: 'var(--hr-text-muted)' }}>{weekLabel}</p>

        <Card style={{ padding: 0, overflow: 'hidden' }}>
          <div style={{ overflowX: 'auto' }}>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(7, minmax(120px, 1fr))',
                minWidth: 840,
              }}
            >
              {days.map((d) => (
                <div key={d.toISOString()} style={{ borderRight: '1px solid var(--hr-border)' }}>
                  <div
                    style={{
                      padding: '10px 12px',
                      background: 'var(--hr-hover)',
                      borderBottom: '1px solid var(--hr-border)',
                    }}
                  >
                    <div
                      style={{
                        fontSize: 10,
                        color: 'var(--hr-text-muted)',
                        textTransform: 'uppercase',
                        fontWeight: 600,
                      }}
                    >
                      {d.toLocaleDateString('en-US', { weekday: 'short' })}
                    </div>
                    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {d.getDate()}
                    </div>
                  </div>
                  <div
                    style={{
                      padding: 8,
                      minHeight: 260,
                      display: 'flex',
                      flexDirection: 'column',
                      gap: 6,
                    }}
                  >
                    {(grid.get(dayStr(d)) ?? []).map((sh) => (
                      <ShiftCell
                        key={sh.id}
                        shift={sh}
                        onClick={() => {
                          toast.warning(
                            t('hrShifts.schedule.confirmDelete', { name: sh.staffName }),
                            {
                              action: {
                                label: t('hrCommon.delete'),
                                onClick: () =>
                                  del.mutate(sh.id, {
                                    onSuccess: () =>
                                      toast.success(t('hrShifts.schedule.shiftRemoved')),
                                  }),
                              },
                            }
                          );
                        }}
                      />
                    ))}
                    <button
                      type="button"
                      onClick={() => openNew(dayStr(d))}
                      style={{
                        padding: '6px',
                        borderRadius: 8,
                        border: '1px dashed var(--hr-border)',
                        background: 'transparent',
                        color: 'var(--hr-text-muted)',
                        fontSize: 10,
                        cursor: 'pointer',
                      }}
                    >
                      + add shift
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </Card>

        <Modal
          open={modalOpen}
          onClose={() => setModalOpen(false)}
          title={t('hrShifts.schedule.newShift')}
        >
          {editing && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div>
                <Label>{t('hrShifts.schedule.employee')}</Label>
                <Select
                  value={editing.staffId ?? ''}
                  onChange={(e) => setEditing({ ...editing, staffId: Number(e.target.value) })}
                >
                  <option value="">{t('hrShifts.schedule.selectPlaceholder')}</option>
                  {staff.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
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
                  <Label>{t('hrShifts.schedule.start')}</Label>
                  <Input
                    type="time"
                    value={editing.start ?? ''}
                    onChange={(e) => setEditing({ ...editing, start: e.target.value })}
                  />
                </div>
                <div>
                  <Label>{t('hrShifts.schedule.end')}</Label>
                  <Input
                    type="time"
                    value={editing.end ?? ''}
                    onChange={(e) => setEditing({ ...editing, end: e.target.value })}
                  />
                </div>
              </div>
              <div>
                <Label>{t('hrShifts.schedule.color')}</Label>
                <div style={{ display: 'flex', gap: 6 }}>
                  {ROLE_COLORS.map((c) => (
                    <button
                      key={c}
                      type="button"
                      onClick={() => setEditing({ ...editing, color: c })}
                      style={{
                        width: 28,
                        height: 28,
                        borderRadius: 8,
                        border:
                          editing.color === c ? `2px solid ${c}` : '1px solid var(--hr-border)',
                        background: c,
                        cursor: 'pointer',
                      }}
                    />
                  ))}
                </div>
              </div>
              <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                <Button variant="ghost" onClick={() => setModalOpen(false)}>
                  {t('hrShifts.schedule.cancel')}
                </Button>
                <Button variant="primary" onClick={saveShift}>
                  {t('hrShifts.schedule.saveShift')}
                </Button>
              </div>
            </div>
          )}
        </Modal>
      </div>
    </>
  );
};

export default Schedule;
