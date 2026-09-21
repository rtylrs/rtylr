import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useCreateTimeEntry } from '../../shared/hooks/queries/useTimeEntries';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { timeEntryFormSchema } from '../../shared/schemas/timeEntries';
import { Modal, Button, Input, Textarea, Label, Checkbox } from '../../shared/components/ui';
import { todayIso } from '../../shared/lib/dateUtils';

interface TimeEntryFormProps {
  open: boolean;
  onClose: () => void;
  prefill?: {
    durationMinutes?: number;
    description?: string;
    source?: 'manual' | 'timer';
  };
}

const TimeEntryForm = ({
  open,
  onClose,
  prefill,
}: TimeEntryFormProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const createEntry = useCreateTimeEntry();
  const { data: myEmployeeId, isError: employeeLookupFailed } = useMyEmployeeId();

  const today = todayIso();
  const [date, setDate] = useState(today);
  const [hours, setHours] = useState(
    prefill?.durationMinutes ? String(Math.floor(prefill.durationMinutes / 60)) : ''
  );
  const [minutes, setMinutes] = useState(
    prefill?.durationMinutes ? String(prefill.durationMinutes % 60) : ''
  );
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [useStartEnd, setUseStartEnd] = useState(false);
  const [description, setDescription] = useState(prefill?.description ?? '');
  const [isBillable, setIsBillable] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // The form stays mounted while the modal is closed, so the state initialisers
  // run only once. Re-seed on every open, or the timer's prefill never reaches
  // the fields and the previous entry's values reappear on the next open.
  const prefillMinutes = prefill?.durationMinutes;
  const prefillDescription = prefill?.description;
  useEffect(() => {
    if (!open) return;
    setDate(todayIso());
    setHours(prefillMinutes ? String(Math.floor(prefillMinutes / 60)) : '');
    setMinutes(prefillMinutes ? String(prefillMinutes % 60) : '');
    setStartTime('');
    setEndTime('');
    setUseStartEnd(false);
    setDescription(prefillDescription ?? '');
    setIsBillable(false);
    setErrors({});
  }, [open, prefillMinutes, prefillDescription]);

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();

    let durationMinutes: number;
    if (useStartEnd && startTime && endTime) {
      const [sh, sm] = startTime.split(':').map(Number);
      const [eh, em] = endTime.split(':').map(Number);
      durationMinutes = eh * 60 + em - (sh * 60 + sm);
      if (durationMinutes <= 0) {
        setErrors({ endTime: t('timeLogging.form.invalidTimeRange') });
        return;
      }
    } else {
      durationMinutes = parseInt(hours || '0', 10) * 60 + parseInt(minutes || '0', 10);
    }

    const formData = {
      date,
      durationMinutes,
      startTime: useStartEnd ? startTime : undefined,
      endTime: useStartEnd ? endTime : undefined,
      description,
      isBillable,
      source: prefill?.source ?? 'manual',
      employeeId: myEmployeeId ?? 0,
    };

    const result = timeEntryFormSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      // zod v4 renamed `.errors` → `.issues`.
      result.error.issues.forEach((err) => {
        const key = err.path[0]?.toString() ?? 'form';
        fieldErrors[key] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    createEntry.mutate(
      { ...formData, employeeId: String(formData.employeeId) },
      {
        onSuccess: () => {
          onClose();
        },
      }
    );
  };

  if (!open) return null;

  return (
    <Modal open={open} onClose={onClose} title={t('timeLogging.form.title')} width={440}>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {/* Date */}
        <div>
          <Label htmlFor="te-date">{t('timeLogging.form.date')}</Label>
          <Input
            id="te-date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            required
            style={{ fontSize: 16 }}
          />
        </div>

        {/* Duration mode toggle */}
        <Checkbox
          checked={useStartEnd}
          onChange={setUseStartEnd}
          label={t('timeLogging.form.useStartEnd')}
        />

        {useStartEnd ? (
          <div style={{ display: 'flex', gap: 10 }}>
            <div style={{ flex: 1 }}>
              <Label htmlFor="te-start">{t('timeLogging.form.startTime')}</Label>
              <Input
                id="te-start"
                type="time"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                required
                style={{ fontSize: 16 }}
              />
            </div>
            <div style={{ flex: 1 }}>
              <Label htmlFor="te-end">{t('timeLogging.form.endTime')}</Label>
              <Input
                id="te-end"
                type="time"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                required
                style={{ fontSize: 16 }}
              />
              {errors.endTime && (
                <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
                  {errors.endTime}
                </p>
              )}
            </div>
          </div>
        ) : (
          <div style={{ display: 'flex', gap: 10 }}>
            <div style={{ flex: 1 }}>
              <Label htmlFor="te-hours">{t('timeLogging.form.hours')}</Label>
              <Input
                id="te-hours"
                type="number"
                min={0}
                max={23}
                value={hours}
                onChange={(e) => setHours(e.target.value)}
                placeholder="0"
                style={{ fontSize: 16 }}
              />
            </div>
            <div style={{ flex: 1 }}>
              <Label htmlFor="te-minutes">{t('timeLogging.form.minutes')}</Label>
              <Input
                id="te-minutes"
                type="number"
                min={0}
                max={59}
                value={minutes}
                onChange={(e) => setMinutes(e.target.value)}
                placeholder="0"
                style={{ fontSize: 16 }}
              />
            </div>
          </div>
        )}

        {/* Description */}
        <div>
          <Label htmlFor="te-desc">{t('timeLogging.form.description')}</Label>
          <Textarea
            id="te-desc"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('timeLogging.form.descriptionPlaceholder')}
            rows={3}
          />
        </div>

        {/* Billable */}
        <Checkbox
          checked={isBillable}
          onChange={setIsBillable}
          label={t('timeLogging.form.billable')}
        />

        {/* Validation errors */}
        {employeeLookupFailed && (
          <p style={{ color: 'var(--hr-error)', fontSize: 12 }}>
            {t('timeLogging.form.employeeLookupFailed')}
          </p>
        )}
        {errors.durationMinutes && (
          <p style={{ color: 'var(--hr-error)', fontSize: 12 }}>
            {t('timeLogging.form.durationRequired')}
          </p>
        )}

        {/* Actions */}
        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
          <Button variant="ghost" type="button" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="primary"
            type="submit"
            disabled={createEntry.isPending || myEmployeeId === undefined}
          >
            {createEntry.isPending ? t('common.loading') : t('common.save')}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default TimeEntryForm;
