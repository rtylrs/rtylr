import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@voxire/ui';
import type { LeaveTypePayload } from './useLeave';
import { useCreateLeaveRequest } from './useLeave';

const labelClass = 'mb-1 block text-[10px] font-semibold uppercase tracking-wider';
const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border)',
  color: 'var(--flow-text)',
};

function daysBetween(start: string, end: string): number {
  const s = new Date(start);
  const e = new Date(end);
  const diff = e.getTime() - s.getTime();
  return Math.max(1, Math.round(diff / 86400000) + 1);
}

interface RequestLeaveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  employeeId: string;
  leaveTypes: LeaveTypePayload[];
}

const RequestLeaveDialog: React.FC<RequestLeaveDialogProps> = ({
  open,
  onOpenChange,
  employeeId,
  leaveTypes,
}) => {
  const { t } = useTranslation();
  const createRequest = useCreateLeaveRequest();

  const [leaveTypeId, setLeaveTypeId] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [notes, setNotes] = useState('');

  const totalDays = useMemo(
    () => (startDate && endDate ? daysBetween(startDate, endDate) : 0),
    [startDate, endDate]
  );

  const canSubmit =
    leaveTypeId !== '' && startDate !== '' && endDate !== '' && endDate >= startDate;

  const reset = (): void => {
    setLeaveTypeId('');
    setStartDate('');
    setEndDate('');
    setNotes('');
  };

  const submit = (): void => {
    if (!canSubmit) return;
    createRequest.mutate(
      {
        employeeId,
        leaveTypeId,
        startDate,
        endDate,
        totalDays,
        notes: notes.trim() || undefined,
      },
      {
        onSuccess: () => {
          reset();
          onOpenChange(false);
        },
      }
    );
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) reset();
        onOpenChange(next);
      }}
    >
      <DialogContent
        className="max-w-lg"
        style={{ background: 'var(--flow-card)', color: 'var(--flow-text)' }}
      >
        <DialogHeader>
          <DialogTitle style={{ color: 'var(--flow-text)' }}>
            {t('flow.leave.newRequest')}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <label
              htmlFor="leave-type"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.leave.leaveType')}
            </label>
            <select
              id="leave-type"
              aria-label={t('flow.leave.leaveType')}
              value={leaveTypeId}
              onChange={(e) => setLeaveTypeId(e.target.value)}
              required
              className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
              style={inputStyle}
            >
              <option value="">{t('flow.leave.selectLeaveType')}</option>
              {leaveTypes.map((type) => (
                <option key={type.id} value={type.id}>
                  {type.name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <label
                htmlFor="leave-start-date"
                className={labelClass}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.leave.startDate')}
              </label>
              <input
                id="leave-start-date"
                aria-label={t('flow.leave.startDate')}
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                style={inputStyle}
              />
            </div>
            <div>
              <label
                htmlFor="leave-end-date"
                className={labelClass}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.leave.endDate')}
              </label>
              <input
                id="leave-end-date"
                aria-label={t('flow.leave.endDate')}
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                style={inputStyle}
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="leave-notes"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.leave.notes')}
            </label>
            <textarea
              id="leave-notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
              className="flow-focus w-full rounded px-2 py-1.5 text-[16px] md:text-[13px]"
              style={inputStyle}
            />
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="flow-focus min-h-[44px] rounded px-3 text-[13px] font-medium md:min-h-0 md:py-2"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
              onClick={submit}
              disabled={!canSubmit || createRequest.isPending}
              className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded px-4 text-[13px] font-semibold disabled:opacity-40 md:min-h-0 md:py-2.5"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {t('flow.leave.submit')}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default RequestLeaveDialog;
