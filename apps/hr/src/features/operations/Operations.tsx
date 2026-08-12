import React, { useState, useMemo } from 'react';
import { useUserStore } from '@voxire/auth';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { Calculator, RefreshCw, Clock, Coffee, AlertTriangle, MapPin } from 'lucide-react';
import {
  useSwaps,
  useCreateSwap,
  useResolveSwap,
  useAvailability,
  useBreaks,
  useStartBreak,
  useEndBreak,
  useOvertimeAlerts,
  useHrSettings,
} from '../../shared/hooks/useHrExtras';
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
  EmptyState,
  ACCENT,
  useIsMobile,
} from '../../shared/components/ui';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import { canHrPermission } from '../../shared/lib/hrPermissions';

const TABS = ['Labor Cost', 'Swap Shifts', 'Availability', 'Breaks', 'Overtime'] as const;

const Operations: React.FC = () => {
  const isMobile = useIsMobile(900);
  const [tab, setTab] = useState<(typeof TABS)[number]>('Labor Cost');
  const { data: staff = [] } = useStaff();
  const { data: payroll = [] } = usePayroll();
  const { data: swaps = [] } = useSwaps();
  const createSwap = useCreateSwap();
  const resolveSwap = useResolveSwap();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canManageShifts = canHrPermission(role, 'hr:shifts:manage');
  const { data: availability = [] } = useAvailability();
  const { data: breaks = [] } = useBreaks();
  const startBreak = useStartBreak();
  const endBreak = useEndBreak();
  const { data: overtimeAlerts = [] } = useOvertimeAlerts();
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
    return payroll.filter((p) => ids.has(p.id));
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
      toast.error('Geolocation unavailable');
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setGpsStatus(`${pos.coords.latitude.toFixed(3)}, ${pos.coords.longitude.toFixed(3)}`);
        toast.success('Clocked in with location');
      },
      () => {
        setGpsStatus('Permission denied');
        toast.error('Location permission denied');
      }
    );
  };

  return (
    <>
      <HelmetContainer title="Operations · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title="Operations"
          subtitle="Labor cost, shift swaps, availability, breaks, overtime"
        />
        <Tabs tabs={[...TABS]} active={tab} onChange={(t) => setTab(t as typeof tab)} />

        {tab === 'Labor Cost' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div
              style={{
                display: 'flex',
                alignItems: isMobile ? 'stretch' : 'center',
                flexDirection: isMobile ? 'column' : 'row',
                gap: 12,
              }}
            >
              <Label style={{ marginBottom: 0 }}>Department</Label>
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
                    Weekly labor
                  </span>
                </div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
                  ${weeklyCost.toLocaleString()}
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
                  ${monthlyCost.toLocaleString()}
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
                  Weekly hours
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
                  Blended rate
                </div>
                <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hr-text)' }}>
                  ${blendedRate.toFixed(2)}/h
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
                  GPS clock-in demo
                </div>
                <div style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
                  Status: {gpsStatus}
                </div>
              </div>
              <Button variant="primary" onClick={checkGps}>
                Clock in with GPS
              </Button>
            </Card>
          </div>
        )}

        {tab === 'Swap Shifts' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Card>
              <SectionHeader title="Request a swap" />
              <div
                style={{
                  marginTop: 10,
                  display: 'grid',
                  gridTemplateColumns: isMobile ? '1fr' : 'repeat(2, 1fr)',
                  gap: 10,
                }}
              >
                <div>
                  <Label>Requester</Label>
                  <Select
                    value={swapForm.requesterId}
                    onChange={(e) =>
                      setSwapForm({ ...swapForm, requesterId: Number(e.target.value) })
                    }
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
                  <Label>Target</Label>
                  <Select
                    value={swapForm.targetId}
                    onChange={(e) => setSwapForm({ ...swapForm, targetId: Number(e.target.value) })}
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
                  <Label>Shift day</Label>
                  <Input
                    type="date"
                    value={swapForm.shiftDay}
                    onChange={(e) => setSwapForm({ ...swapForm, shiftDay: e.target.value })}
                  />
                </div>
                <div>
                  <Label>Reason</Label>
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
                      toast.error('Complete all fields');
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
                          toast.success('Swap requested');
                          setSwapForm({ requesterId: 0, targetId: 0, shiftDay: '', reason: '' });
                        },
                      }
                    );
                  }}
                >
                  Request swap
                </Button>
              </div>
            </Card>
            {swaps.length === 0 ? (
              <EmptyState
                icon={<RefreshCw size={20} />}
                title="No swap requests"
                description="Shift swap requests will show up here."
              />
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {swaps.map((s) => (
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
                        s.status === 'Accepted'
                          ? '#10b981'
                          : s.status === 'Rejected'
                            ? '#ef4444'
                            : '#f59e0b'
                      }
                    >
                      {s.status}
                    </Pill>
                    {canManageShifts && s.status === 'Pending' && (
                      <ApprovalButtons
                        onApprove={() =>
                          resolveSwap.mutate(
                            { id: s.id, accepted: true },
                            { onSuccess: () => toast.success('Swap accepted') }
                          )
                        }
                        onReject={() =>
                          resolveSwap.mutate(
                            { id: s.id, accepted: false },
                            { onSuccess: () => toast.success('Swap rejected') }
                          )
                        }
                      />
                    )}
                  </Card>
                ))}
              </div>
            )}
          </div>
        )}

        {tab === 'Availability' && (
          <Card style={{ padding: 0, overflow: 'hidden' }}>
            {availability.length === 0 ? (
              <EmptyState
                icon={<Clock size={20} />}
                title="No availability set"
                description="Ask team to mark their weekly availability."
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
                        Employee
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
                        Day
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
                        Can work
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
                        Hours
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {availability.map((a) => (
                      <tr key={a.id} style={{ borderTop: '1px solid var(--hr-border)' }}>
                        <td
                          style={{ padding: '10px 12px', color: 'var(--hr-text)', fontWeight: 600 }}
                        >
                          {a.staffName}
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {a.day}
                        </td>
                        <td style={{ padding: '10px 12px' }}>
                          <Pill color={a.canWork ? '#10b981' : '#ef4444'}>
                            {a.canWork ? 'Yes' : 'No'}
                          </Pill>
                        </td>
                        <td style={{ padding: '10px 12px', color: 'var(--hr-text-muted)' }}>
                          {a.canWork ? `${a.startTime} - ${a.endTime}` : '-'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        )}

        {tab === 'Breaks' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Card>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <Coffee size={14} style={{ color: ACCENT }} />
                <div style={{ flex: 1, fontSize: 12, color: 'var(--hr-text-muted)' }}>
                  Track breaks for compliance and overtime calculations
                </div>
                <Button
                  variant="primary"
                  onClick={() =>
                    startBreak.mutate(
                      {
                        staffId: staff[0]?.id ?? 1,
                        staffName: staff[0]?.name ?? 'Alex Chen',
                        type: 'Lunch',
                      },
                      { onSuccess: () => toast.success('Break started') }
                    )
                  }
                >
                  Start lunch break
                </Button>
              </div>
            </Card>
            <Card style={{ padding: 0 }}>
              {breaks.length === 0 ? (
                <EmptyState
                  icon={<Coffee size={20} />}
                  title="No breaks logged"
                  description="Start a break to track time."
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
                        Employee
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
                        Type
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
                        Started
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
                        Duration
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
                                  onSuccess: () => toast.success('Break ended'),
                                })
                              }
                            >
                              End
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

        {tab === 'Overtime' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {overtimeAlerts.length === 0 ? (
              <EmptyState
                icon={<AlertTriangle size={20} />}
                title="No overtime alerts"
                description="Everyone is within policy limits."
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
