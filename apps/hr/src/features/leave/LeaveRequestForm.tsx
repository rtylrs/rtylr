import React, { useState, useMemo } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import {
  useListLeaveTypes,
  useCreateLeaveRequest,
  useListLeaveBalances,
  useListLeaveRequests,
} from '../../shared/hooks/queries/useLeave';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  Label,
  Input,
  Select,
  Textarea,
  Skeleton,
} from '../../shared/components/ui';
import { todayIso } from '../../shared/lib/dateUtils';

const leaveFormSchema = z.object({
  leaveTypeId: z.string().min(1),
  startDate: z.string().min(1),
  endDate: z.string().min(1),
  notes: z.string().optional(),
});

function daysBetween(start: string, end: string): number {
  const s = new Date(start);
  const e = new Date(end);
  const diff = e.getTime() - s.getTime();
  return Math.max(1, Math.round(diff / 86400000) + 1);
}

const LeaveRequestForm = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: myEmployeeId } = useMyEmployeeId();
  const employeeId = myEmployeeId !== undefined ? String(myEmployeeId) : undefined;
  const {
    data: leaveTypes,
    isLoading: typesLoading,
    isError: typesError,
    refetch: refetchLeaveTypes,
  } = useListLeaveTypes();
  const { data: balances } = useListLeaveBalances(employeeId);
  const { data: existingRequests } = useListLeaveRequests({
    employeeId,
    status: undefined,
  });
  const createRequest = useCreateLeaveRequest();

  const [form, setForm] = useState({
    leaveTypeId: '',
    startDate: '',
    endDate: '',
    notes: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const selectedBalance = useMemo(() => {
    if (!balances || form.leaveTypeId === '') return null;
    return (
      balances.find((b) => b.employeeId === employeeId && b.leaveTypeId === form.leaveTypeId) ??
      null
    );
  }, [balances, form.leaveTypeId, employeeId]);

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    const result = leaveFormSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.issues.forEach((issue) => {
        const field = issue.path[0] as string;
        fieldErrors[field] = t('common.required');
      });
      setErrors(fieldErrors);
      return;
    }
    setErrors({});
    const { leaveTypeId, startDate, endDate, notes } = result.data;

    const now = new Date();
    const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    if (startDate < today || endDate < today) {
      toast.error(t('leave.form.errorPastDate'));
      return;
    }

    const requestedDays = daysBetween(startDate, endDate);
    const remainingDays = selectedBalance?.remainingDays;
    if (remainingDays !== undefined && requestedDays > remainingDays) {
      toast.error(
        `Requested ${requestedDays} day(s) exceeds remaining balance of ${remainingDays} day(s)`
      );
      return;
    }

    const activeRequests = (existingRequests ?? []).filter((r) => {
      const st = (r.status || '').toLowerCase();
      return st === 'approved' || st === 'pending';
    });
    const hasOverlap = activeRequests.some((r) => {
      const rStart = r.startDate || '';
      const rEnd = r.endDate || r.startDate || '';
      return startDate <= rEnd && endDate >= rStart;
    });
    if (hasOverlap) {
      toast.error(t('leave.form.errorOverlap'));
      return;
    }

    createRequest.mutate(
      {
        employeeId: employeeId ?? '',
        leaveTypeId,
        startDate,
        endDate,
        totalDays: daysBetween(startDate, endDate),
        notes: notes || undefined,
      },
      { onSuccess: () => navigate({ to: '/leave' }) }
    );
  };

  const update = (field: string, value: string): void => {
    setForm((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => {
      const next = { ...prev };
      delete next[field];
      return next;
    });
  };

  if (typesLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={200} />
      </div>
    );
  }

  if (typesError) {
    return (
      <Card style={{ textAlign: 'center', padding: 32 }}>
        <p style={{ fontSize: 14, fontWeight: 600, color: 'var(--hr-error)', margin: '0 0 8px' }}>
          Unable to load leave policies
        </p>
        <p style={{ fontSize: 12, color: 'var(--hr-text-muted)', margin: '0 0 16px' }}>
          The leave policy service could not be reached. Try again.
        </p>
        <Button variant="secondary" onClick={() => refetchLeaveTypes()}>
          Retry
        </Button>
      </Card>
    );
  }

  if (!leaveTypes || leaveTypes.length === 0) {
    return (
      <Card style={{ textAlign: 'center', padding: 32 }}>
        <p style={{ fontSize: 14, fontWeight: 600, color: 'var(--hr-text)', margin: '0 0 8px' }}>
          {t('leave.noLeaveTypes')}
        </p>
        <p style={{ fontSize: 12, color: 'var(--hr-text-muted)', margin: '0 0 16px' }}>
          {t('leave.configureLeaveTypes')}
        </p>
        <Button variant="primary" onClick={() => navigate({ to: '/settings/leave-policy' })}>
          Configure leave policies
        </Button>
      </Card>
    );
  }

  return (
    <>
      <HelmetContainer title="Leave request · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 560 }}>
        <SectionHeader title={t('leave.form.title')} subtitle={t('leave.form.subtitle')} />

        <form onSubmit={handleSubmit}>
          <Card style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <Label>{t('leave.form.leaveType')}</Label>
              <Select
                value={form.leaveTypeId}
                onChange={(e) => update('leaveTypeId', e.target.value)}
                style={errors.leaveTypeId ? { borderColor: 'var(--hr-error)' } : undefined}
              >
                <option value="">{t('leave.form.selectType')}</option>
                {leaveTypes.map((lt) => (
                  <option key={lt.id} value={lt.id}>
                    {lt.name}
                  </option>
                ))}
              </Select>
              {errors.leaveTypeId && (
                <p style={{ fontSize: 11, color: 'var(--hr-error)', margin: '4px 0 0' }}>
                  {errors.leaveTypeId}
                </p>
              )}
            </div>

            {selectedBalance && (
              <div
                style={{
                  padding: '8px 12px',
                  borderRadius: 8,
                  background: 'var(--hr-hover)',
                  fontSize: 12,
                }}
              >
                <span style={{ color: 'var(--hr-text-muted)' }}>
                  {t('leave.form.remainingBalance')}:{' '}
                </span>
                <span style={{ fontWeight: 700, color: 'var(--hr-text)' }}>
                  {selectedBalance.remainingDays} {t('leave.balances.daysLeft')}
                </span>
              </div>
            )}

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <div>
                <Label>{t('leave.form.startDate')}</Label>
                <Input
                  type="date"
                  value={form.startDate}
                  onChange={(e) => update('startDate', e.target.value)}
                  min={todayIso()}
                  style={errors.startDate ? { borderColor: 'var(--hr-error)' } : undefined}
                />
                {errors.startDate && (
                  <p style={{ fontSize: 11, color: 'var(--hr-error)', margin: '4px 0 0' }}>
                    {errors.startDate}
                  </p>
                )}
              </div>
              <div>
                <Label>{t('leave.form.endDate')}</Label>
                <Input
                  type="date"
                  value={form.endDate}
                  onChange={(e) => update('endDate', e.target.value)}
                  min={form.startDate || todayIso()}
                  style={errors.endDate ? { borderColor: 'var(--hr-error)' } : undefined}
                />
                {errors.endDate && (
                  <p style={{ fontSize: 11, color: 'var(--hr-error)', margin: '4px 0 0' }}>
                    {errors.endDate}
                  </p>
                )}
              </div>
            </div>

            <div>
              <Label>{t('leave.form.notes')}</Label>
              <Textarea
                value={form.notes}
                onChange={(e) => update('notes', e.target.value)}
                placeholder={t('leave.form.notesPlaceholder')}
                style={{ minHeight: 60 }}
              />
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <Button type="button" variant="ghost" onClick={() => navigate({ to: '/leave' })}>
                {t('common.cancel')}
              </Button>
              <Button
                type="submit"
                variant="primary"
                disabled={createRequest.isPending || !employeeId}
              >
                {createRequest.isPending ? t('common.loading') : t('leave.form.submit')}
              </Button>
            </div>
          </Card>
        </form>
      </div>
    </>
  );
};

export default LeaveRequestForm;
