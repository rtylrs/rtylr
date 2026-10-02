import React, { useState, useMemo, useCallback } from 'react';
import { useUserStore } from '../../../../shared/stores/userStore';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { Calculator, RefreshCw, Clock, Coffee, AlertTriangle, MapPin } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import {
  useSwaps,
  useCreateSwap,
  useResolveSwap,
  useBreaks,
  useStartBreak,
  useEndBreak,
  useOvertimeAlerts,
  useHrSettings,
} from '../../shared/hooks/useHrExtras';
import { useAvailability } from '../../shared/hooks/queries/useAvailability';
import type { Availability } from '../../shared/schemas/shifts';
import { usePayroll } from '../payroll/usePayroll';
import { useStaff } from '../staff/useStaff';
import {
  Card,
  SectionHeader,
  Button,
  Tabs,
  Label,
  Select,
  Input,
  Pill,
  ACCENT,
  useIsMobile,
} from '../../shared/components/ui';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import { formatCurrency } from '../../shared/lib/hrNumberUtils';
import { canHrPermission, canViewCompensation } from '../../shared/lib/hrPermissions';
import { normalizeHrRequestStatus } from '../../shared/lib/requestStatus';

const TABS = ['Labor Cost', 'Swap Shifts', 'Availability', 'Breaks', 'Overtime'] as const;

/**
 * Indexed by `AvailabilityPayload.day_of_week`, which is 0 = Monday through
 * 6 = Sunday so the value indexes the same Monday-first week the roster uses.
 * Deliberately not `Date.getDay()` order, where 0 is Sunday.
 */
const WEEKDAY_KEYS = [
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
  'sunday',
] as const;

const Operations: React.FC = () => {
  const { t } = useTranslation('hrCommon');
  const isMobile = useIsMobile(900);
  const showCompensation = canViewCompensation();
  const [tab, setTab] = useState<(typeof TABS)[number]>(
    showCompensation ? 'Labor Cost' : 'Swap Shifts'
  );
  const { data: staff = [] } = useStaff();
  const { data: payroll = [] } = usePayroll({ enabled: showCompensation });
  const { data: swaps = [], refetch: refetchSwaps } = useSwaps();
  const createSwap = useCreateSwap();
  const resolveSwap = useResolveSwap();
  const canManageShifts = canHrPermission('hr:schedules:manage');
  const visibleTabs = showCompensation ? TABS : TABS.filter((name) => name !== 'Labor Cost');
  const activeTab = !showCompensation && tab === 'Labor Cost' ? 'Swap Shifts' : tab;
  const { data: availability = [], refetch: refetchAvailability } = useAvailability();

  /**
   * A row is either a recurring weekly pattern or a single-date override. The
   * fallback covers a row carrying neither, which the handler is meant to reject
   * but which must not blank the table if one ever arrives.
   */
  const formatAvailabilityDay = (row: Availability): string => {
    if (row.specificDate) return row.specificDate;
    if (row.dayOfWeek !== undefined) {
      return t(`hrOperations.availability.weekdays.${WEEKDAY_KEYS[row.dayOfWeek]}`);
    }
    return '-';
  };
  const { data: breaks = [] } = useBreaks();
  const startBreak = useStartBreak();
  const endBreak = useEndBreak();
  const { data: overtimeAlerts = [], refetch: refetchOvertimeAlerts } = useOvertimeAlerts();
  const handleStartLunchBreak = useCallback((): void => {
    startBreak.mutate(
      { staffId: staff[0]?.id ?? 1, staffName: staff[0]?.name ?? 'Alex Chen', type: 'Lunch' },
      { onSuccess: () => toast.success(t('hrOperations.toasts.breakStarted')) }
    );
  }, [startBreak, staff, t]);
  const { data: settings } = useHrSettings();

  /* Labor cost */
  const [deptFilter, setDeptFilter] = useState('All');
  const departments = useMemo(
    () => ['All', ...Array.from(new Set(staff.map((s) => s.department)))],
    [staff]
  );
  const filteredPayroll = useMemo(() => {
    if (deptFilter === 'All') return payroll;
    const ids = new Set(staff.filter((s) => s.department === deptFilter).map((s) => s.id));
    return payroll.filter((p) => ids.has(p.staffId));
  }, [payroll, staff, deptFilter]);
  const weeklyCost = filteredPayroll.reduce(
    (sum, p) => sum + p.basePay + p.overtime - p.deductions,
    0
  );
  const monthlyCost = weeklyCost * (52 / 12);
  const weeklyHours = filteredPayroll.reduce((sum, p) => sum + p.hoursWorked, 0);
  const blendedRate = weeklyHours > 0 ? weeklyCost / weeklyHours : 0;

  /* Swap shift form */
  const [swapForm, setSwapForm] = useState({
    requesterId: 0,
    targetId: 0,
    shiftDay: '',
    reason: '',
  });

  /* GPS clock-in demo */
  const [gpsStatus, setGpsStatus] = useState<string>('Not checked');
  const checkGps = () => {
    if (!navigator.geolocation) {
      setGpsStatus('Geolocation not supported');
      toast.error(t('hrOperations.toasts.geolocationUnavailable'));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setGpsStatus(`${pos.coords.latitude.toFixed(3)}, ${pos.coords.longitude.toFixed(3)}`);
        toast.success(t('hrOperations.toasts.clockedInWithLocation'));
      },
      () => {
        setGpsStatus('Permission denied');
        toast.error(t('hrOperations.toasts.locationPermissionDenied'));
      }
    );
  };

  return (
    <>
      <HelmetContainer title={t('hrOperations.metaTitle')} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader title={t('hrOperations.title')} subtitle={t('hrOperations.subtitle')} />
        <Tabs
          tabs={[...visibleTabs]}
          active={activeTab}
          onChange={(t) => setTab(t as typeof tab)}
        />

        {activeTab === 'Labor Cost' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div
              style={{
                display: 'flex',
                alignItems: isMobile ? 'stretch' : 'center',
                flexDirection: isMobile ? 'column' : 'row',
                gap: 12,
              }}
            >
              <Label style={{ marginBottom: 0 }}>{t('hrOperations.department')}</Label>
              <div style={{ width: isMobile ? '100%' : 200 }}>
                <Select value={deptFilter} onChange={(e) => setDeptFilter(e.target.value)}>
                  {departments.map((d) => (
                    <option key={d}>{d}</option>
                  ))}
                </Select>
              </div>
            </div>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
                gap: 12,
              }}
            >
              <Card>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                  <Calculator size={13} style={{ color: ACCENT }} />
                  <span
                    style={{
                      fontSize: 11,
                      color: 'var(--hr-text-muted)',
                      textTransform: 'uppercase',
                      fontWeight: 600,
                    }}
                  >
                    {t('hrOperations.weeklyLabor')}
                  </span>
                </div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
                  {formatCurrency(weeklyCost, settings?.currency)}
                </div>
              </Card>
              <Card>
                <div
                  style={{
                    fontSize: 11,
                    color: 'var(--hr-text-muted)',
                    textTransform: 'uppercase',
                    fontWeight: 600,
                    marginBottom: 6,
                  }}
                >
                  Monthly (est)
                </div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
                  {formatCurrency(monthlyCost, settings?.currency)}
                </div>
              </Card>
              <Card>
                <div
                  style={{
                    fontSize: 11,
                    color: 'var(--hr-text-muted)',
                    textTransform: 'uppercase',
                    fontWeight: 600,
                    marginBottom: 6,
                  }}
                >
                  {t('hrOperations.weeklyHours')}
                </div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
                  {weeklyHours}h
                </div>
              </Card>
              <Card>
                <div
                  style={{
                    fontSize: 11,
                    color: 'var(--hr-text-muted)',
                    textTransform: 'uppercase',
                    fontWeight: 600,
                    marginBottom: 6,
                  }}
                >
                  {t('hrOperations.blendedRate')}
                </div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
                  {formatCurrency(blendedRate, settings?.currency)}/h
                </div>
              </Card>
            </div>
            <Card>
              <div style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
                Weekly hours policy:{' '}
                <strong style={{ color: 'var(--hr-text)' }}>{settings?.weeklyHours ?? 40}h</strong>{' '}
                • Overtime after:{' '}
                <strong style={{ color: 'var(--hr-text)' }}>
                  {settings?.overtimeAfterHours ?? 40}h
                </strong>
              </div>
            </Card>
            <Card
              style={{
                display: 'flex',
                alignItems: isMobile ? 'flex-start' : 'center',
                flexDirection: isMobile ? 'column' : 'row',
                gap: 12,
              }}
            >
              <MapPin size={14} style={{ color: ACCENT }} />
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                  {t('hrOperations.gpsTitle')}
                </div>
                <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                  Status: {gpsStatus}
                </div>
              </div>
              <Button variant="primary" onClick={checkGps}>
                {t('hrOperations.gpsClockIn')}
              </Button>
            </Card>
          </div>
        )}

        {activeTab === 'Swap Shifts' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Card>
              <SectionHeader title={t('hrOperations.requestSwapTitle')} />
              <div
                style={{
                  marginTop: 10,
                  display: 'grid',
                  gridTemplateColumns: isMobile ? '1fr' : 'repeat(2, 1fr)',
                  gap: 10,
                }}
              >
                <div>
                  <Label>{t('hrOperations.form.requester')}</Label>
                  <Select
                    value={swapForm.requesterId}
                    onChange={(e) =>
                      setSwapForm({ ...swapForm, requesterId: Number(e.target.value) })
                    }
                  >
                    <option value={0}>{t('hrOperations.selectPlaceholder')}</option>
                    {staff.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name}
                      </option>
                    ))}
                  </Select>
                </div>
                <div>
                  <Label>{t('hrOperations.form.target')}</Label>
                  <Select
                    value={swapForm.targetId}
                    onChange={(e) => setSwapForm({ ...swapForm, targetId: Number(e.target.value) })}
                  >
                    <option value={0}>{t('hrOperations.selectPlaceholder')}</option>
                    {staff.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name}
                      </option>
                    ))}
                  </Select>
                </div>
                <div>
                  <Label>{t('hrOperations.form.shiftDay')}</Label>
                  <Input
                    type="date"
                    value={swapForm.shiftDay}
                    onChange={(e) => setSwapForm({ ...swapForm, shiftDay: e.target.value })}
                  />
                </div>
                <div>
                  <Label>{t('hrOperations.form.reason')}</Label>
                  <Input
                    value={swapForm.reason}
                    onChange={(e) => setSwapForm({ ...swapForm, reason: e.target.value })}
                  />
                </div>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 12 }}>
                <Button
                  variant="primary"
                  icon={<RefreshCw size={13} />}
                  onClick={() => {
                    const requester = staff.find((s) => s.id === swapForm.requesterId);
                    const target = staff.find((s) => s.id === swapForm.targetId);
                    if (!requester || !target || !swapForm.shiftDay) {
                      toast.error(t('hrOperations.toasts.completeAllFields'));
                      return;
                    }
                    createSwap.mutate(
                      {
                        requesterId: requester.id,
                        requester: requester.name,
                        targetId: target.id,
                        target: target.name,
                        shiftDay: swapForm.shiftDay,
                        reason: swapForm.reason,
                      },
                      {
                        onSuccess: () => {
                          toast.success(t('hrOperations.toasts.swapRequested'));
                          setSwapForm({ requesterId: 0, targetId: 0, shiftDay: '', reason: '' });
                        },
                      }
                    );
                  }}
                >
                  {t('hrOperations.requestSwap')}
                </Button>
              </div>
            </Card>
            {swaps.length === 0 ? (
              <EmptyState
                icon={RefreshCw}
                title={t('hrOperations.emptySwapsTitle')}
                description={t('hrOperations.emptySwapsDescription')}
                action={{ label: t('hrCommon.refresh'), onClick: () => refetchSwaps() }}
              />
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {swaps.map((s) => {
                  const status = normalizeHrRequestStatus(s.status);
                  return (
                    <Card key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <div style={{ flex: 1 }}>
                        <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                          {s.requester}{' '}
                          <span style={{ color: 'var(--hr-text-muted)' }}>wants to swap with</span>{' '}
                          {s.target}
                        </div>
                        <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                          {s.shiftDay} • {s.reason}
                        </div>
                      </div>
                      <Pill
                        color={
                          status === 'accepted'
                            ? '#10b981'
                            : status === 'rejected'
                              ? '#ef4444'
                              : '#f59e0b'
                        }
                      >
                        {s.status}
                      </Pill>
                      {canManageShifts && status === 'pending' && (
                        <ApprovalButtons
                          onApprove={() =>
                            resolveSwap.mutate(
                              { id: s.id, accepted: true },
                              {
                                onSuccess: () =>
                                  toast.success(t('hrOperations.toasts.swapAccepted')),
                              }
                            )
                          }
                          onReject={() =>
                            resolveSwap.mutate(
                              { id: s.id, accepted: false },
                              {
                                onSuccess: () =>
                                  toast.success(t('hrOperations.toasts.swapRejected')),
                              }
                            )
                          }
                        />
                      )}
                    </Card>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {activeTab === 'Availability' && (
          <Card style={{ padding: 0, overflow: 'hidden' }}>
            {availability.length === 0 ? (
              <EmptyState
                icon={Clock}
                title={t('hrOperations.availability.emptyTitle')}
                description={t('hrOperations.availability.emptyDescription')}
                action={{ label: t('hrCommon.refresh'), onClick: () => refetchAvailability() }}
              />
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
                  <thead>
                    <tr style={{ background: 'var(--hr-hover)' }}>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.availability.columns.employee')}
                      </th>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.availability.columns.day')}
                      </th>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.availability.columns.available')}
                      </th>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.availability.columns.hours')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {availability.map((a) => {
                      const statusColor = a.isAvailable
                        ? 'var(--hr-status-active)'
                        : 'var(--hr-status-inactive)';
                      return (
                        <tr key={a.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                          <td
                            style={{
                              padding: '10px 12px',
                              color: 'var(--hr-text)',
                              fontWeight: 600,
                            }}
                          >
                            {a.staffName}
                          </td>
                          <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                            {formatAvailabilityDay(a)}
                          </td>
                          <td style={{ padding: '10px 12px' }}>
                            {/* Pill builds its background as `${color}22`, which a
                                var() cannot be concatenated into, so the tinted
                                background and border come through color-mix here. */}
                            <Pill
                              color={statusColor}
                              style={{
                                background: `color-mix(in srgb, ${statusColor} 13%, transparent)`,
                                border: `1px solid color-mix(in srgb, ${statusColor} 20%, transparent)`,
                              }}
                            >
                              {a.isAvailable
                                ? t('hrOperations.availability.available')
                                : t('hrOperations.availability.blocked')}
                            </Pill>
                          </td>
                          <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                            {a.startTime && a.endTime ? `${a.startTime} - ${a.endTime}` : '-'}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        )}

        {activeTab === 'Breaks' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Card>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <Coffee size={14} style={{ color: ACCENT }} />
                <div style={{ flex: 1, fontSize: 12, color: 'var(--hr-text-muted)' }}>
                  {t('hrOperations.breaksSubtitle')}
                </div>
                <Button variant="primary" onClick={handleStartLunchBreak}>
                  {t('hrOperations.startLunchBreak')}
                </Button>
              </div>
            </Card>
            <Card style={{ padding: 0 }}>
              {breaks.length === 0 ? (
                <EmptyState
                  icon={Coffee}
                  title={t('hrOperations.emptyBreaksTitle')}
                  description={t('hrOperations.emptyBreaksDescription')}
                  action={{
                    label: t('hrOperations.startLunchBreak'),
                    onClick: handleStartLunchBreak,
                  }}
                />
              ) : (
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
                  <thead>
                    <tr style={{ background: 'var(--hr-hover)' }}>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.breakColumns.employee')}
                      </th>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.breakColumns.type')}
                      </th>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.breakColumns.started')}
                      </th>
                      <th
                        style={{
                          padding: '10px 12px',
                          textAlign: 'left',
                          fontSize: 10,
                          color: 'var(--hr-text-muted)',
                          textTransform: 'uppercase',
                        }}
                      >
                        {t('hrOperations.breakColumns.duration')}
                      </th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {breaks.map((b) => (
                      <tr key={b.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                        <td
                          style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}
                        >
                          {b.staffName}
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {b.type}
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {new Date(b.start).toLocaleString()}
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {b.durationMin !== undefined ? `${b.durationMin}m` : 'In progress'}
                        </td>
                        <td style={{ padding: '10px 12px' }}>
                          {!b.end && (
                            <Button
                              variant="primary"
                              style={{ padding: '4px 10px' }}
                              onClick={() =>
                                endBreak.mutate(b.id, {
                                  onSuccess: () =>
                                    toast.success(t('hrOperations.toasts.breakEnded')),
                                })
                              }
                            >
                              {t('hrOperations.endBreak')}
                            </Button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </Card>
          </div>
        )}

        {activeTab === 'Overtime' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {overtimeAlerts.length === 0 ? (
              <EmptyState
                icon={AlertTriangle}
                title={t('hrOperations.emptyOvertimeTitle')}
                description={t('hrOperations.emptyOvertimeDescription')}
                action={{ label: t('hrCommon.refresh'), onClick: () => refetchOvertimeAlerts() }}
              />
            ) : (
              overtimeAlerts.map((a) => (
                <Card
                  key={a.id}
                  style={{ borderColor: 'rgba(239,68,68,0.4)', background: 'rgba(239,68,68,0.06)' }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <AlertTriangle size={16} style={{ color: '#ef4444' }} />
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                        {a.staffName}
                      </div>
                      <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                        {a.hoursThisWeek}h this week (threshold {a.threshold}h) • triggered{' '}
                        {a.triggeredAt}
                      </div>
                    </div>
                    <Pill color="#ef4444">Over {a.hoursThisWeek - a.threshold}h</Pill>
                  </div>
                </Card>
              ))
            )}
          </div>
        )}
      </div>
    </>
  );
};

export default Operations;
