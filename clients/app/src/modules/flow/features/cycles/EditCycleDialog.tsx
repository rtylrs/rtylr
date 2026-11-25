import React, { useEffect, useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { DatePicker } from '@voxire/ui';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../../shared/components/ui/dialog';
import { Button } from '../../../../shared/components/ui/button';
import type { ApiCycle } from '../../shared/services/flowApi';

const fieldStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};

export interface CycleEdit {
  name: string;
  startDate: string;
  endDate: string;
  goal: string;
}

interface EditCycleDialogProps {
  cycle: ApiCycle | null;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (edit: CycleEdit) => void;
}

const EditCycleDialog: React.FC<EditCycleDialogProps> = ({
  cycle,
  pending,
  onOpenChange,
  onSave,
}) => {
  const { t, i18n } = useTranslation();
  const nameId = useId();
  const goalId = useId();
  const [name, setName] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [goal, setGoal] = useState('');

  useEffect(() => {
    if (cycle) {
      setName(cycle.name);
      setStartDate(cycle.startDate);
      setEndDate(cycle.endDate);
      setGoal(cycle.goal ?? '');
    }
  }, [cycle]);

  const invalid = !name.trim() || !startDate || !endDate || endDate < startDate;
  const datePickerLabels = {
    previousMonth: t('flowCommon.datePicker.previousMonth'),
    nextMonth: t('flowCommon.datePicker.nextMonth'),
  };

  return (
    <Dialog open={cycle !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('flow.cycles.edit.title')}</DialogTitle>
          <DialogDescription>{t('flow.cycles.edit.description')}</DialogDescription>
        </DialogHeader>
        <div className="space-y-3 px-6">
          <div className="space-y-1">
            <label
              htmlFor={nameId}
              className="text-flow-meta font-medium"
              style={{ color: 'var(--flow-text)' }}
            >
              {t('flow.cycles.edit.name')}
            </label>
            <input
              id={nameId}
              value={name}
              onChange={(e) => setName(e.target.value)}
              dir="auto"
              className="flow-focus w-full rounded-md px-3 py-2 text-flow-meta"
              style={fieldStyle}
            />
          </div>
          <div className="flex gap-2">
            <DatePicker
              value={startDate}
              onChange={setStartDate}
              locale={i18n.language}
              aria-label={t('flow.cycles.startDate')}
              labels={{ placeholder: t('flow.cycles.startDate'), ...datePickerLabels }}
              className="text-flow-meta"
            />
            <DatePicker
              value={endDate}
              onChange={setEndDate}
              minDate={startDate || undefined}
              locale={i18n.language}
              aria-label={t('flow.cycles.endDate')}
              labels={{ placeholder: t('flow.cycles.endDate'), ...datePickerLabels }}
              className="text-flow-meta"
            />
          </div>
          <div className="space-y-1">
            <label
              htmlFor={goalId}
              className="text-flow-meta font-medium"
              style={{ color: 'var(--flow-text)' }}
            >
              {t('flow.backlog.start.goal')}
            </label>
            <textarea
              id={goalId}
              value={goal}
              onChange={(e) => setGoal(e.target.value)}
              rows={3}
              dir="auto"
              placeholder={t('flow.backlog.start.goalPlaceholder')}
              className="flow-focus w-full rounded-md px-3 py-2 text-flow-meta"
              style={fieldStyle}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('flow.cycles.cancel')}
          </Button>
          <Button
            disabled={invalid || pending}
            onClick={() => onSave({ name: name.trim(), startDate, endDate, goal: goal.trim() })}
          >
            {pending ? t('flow.cycles.edit.saving') : t('flow.cycles.edit.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default EditCycleDialog;
