import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  DatePicker,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@voxire/ui';
import GateNotice from '../../shared/components/GateNotice';
import type { LeaveTypePayload } from './useLeave';
import { useCreateLeaveRequest } from './useLeave';

const labelClass = 'mb-1 block text-flow-meta font-semibold uppercase tracking-wider';
const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
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
  const { t, i18n } = useTranslation();
  const createRequest = useCreateLeaveRequest();

  const [leaveTypeId, setLeaveTypeId] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [notes, setNotes] = useState('');

  const totalDays = useMemo(
    () => (startDate && endDate ? daysBetween(startDate, endDate) : 0),
    [startDate, endDate]
  );

  // A workspace that has not configured any leave types yet leaves the Type
  // field with nothing real to select - the placeholder is its only option.
  // Root CLAUDE.md rule 12 ("a control that cannot store its value is either
  // absent or disabled with the reason beside it") applies here: the field
  // and the submit button both surface this specific reason rather than
  // rendering as though a normal, fillable form.
  const hasLeaveTypes = leaveTypes.length > 0;
  const noLeaveTypesReason = hasLeaveTypes ? '' : t('flow.leave.noLeaveTypesReason');

  const canSubmit =
    hasLeaveTypes &&
    leaveTypeId !== '' &&
    startDate !== '' &&
    endDate !== '' &&
    endDate >= startDate;

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
          <DialogDescription style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.leave.newRequestDescription')}
          </DialogDescription>
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
              disabled={!hasLeaveTypes}
              aria-describedby={noLeaveTypesReason ? 'leave-type-gate' : undefined}
              className="flow-focus min-h-[44px] w-full rounded-sm px-2 text-flow-title disabled:cursor-not-allowed md:min-h-0 md:py-1.5 md:text-flow-body"
              style={{
                ...inputStyle,
                background: hasLeaveTypes ? inputStyle.background : 'var(--flow-hover)',
                color: hasLeaveTypes ? inputStyle.color : 'var(--flow-text-muted)',
              }}
            >
              <option value="">{t('flow.leave.selectLeaveType')}</option>
              {leaveTypes.map((type) => (
                <option key={type.id} value={type.id}>
                  {type.name}
                </option>
              ))}
            </select>
            <GateNotice id="leave-type-gate" className="mt-1" reason={noLeaveTypesReason} />
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
              <DatePicker
                id="leave-start-date"
                value={startDate}
                onChange={setStartDate}
                maxDate={endDate || undefined}
                locale={i18n.language}
                labels={{
                  previousMonth: t('flowCommon.datePicker.previousMonth'),
                  nextMonth: t('flowCommon.datePicker.nextMonth'),
                }}
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
              <DatePicker
                id="leave-end-date"
                value={endDate}
                onChange={setEndDate}
                minDate={startDate || undefined}
                locale={i18n.language}
                labels={{
                  previousMonth: t('flowCommon.datePicker.previousMonth'),
                  nextMonth: t('flowCommon.datePicker.nextMonth'),
                }}
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
              className="flow-focus w-full rounded-sm px-2 py-1.5 text-flow-title md:text-flow-body"
              style={inputStyle}
            />
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="flow-focus min-h-[44px] rounded-sm px-3 text-flow-body font-medium md:min-h-0 md:py-2"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flowCommon.cancel')}
            </button>
            <button
              type="button"
              onClick={submit}
              disabled={!canSubmit || createRequest.isPending}
              title={noLeaveTypesReason || undefined}
              aria-describedby={noLeaveTypesReason ? 'leave-type-gate' : undefined}
              className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded-sm px-4 text-flow-body font-semibold disabled:cursor-not-allowed disabled:opacity-40 md:min-h-0 md:py-2.5"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
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
