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
} from '../../../../../shared/components/ui/dialog';
import { Button } from '../../../../../shared/components/ui/button';
import FlowSelect from '../../../shared/components/FlowSelect';
import type { ApiCycle, ApiCycleTaskDecision } from '../../../shared/services/flowApi';
import type { Task } from '../../../shared/stores/flowStore';
import { CYCLE_LENGTH_WEEKS } from './backlogModel';
import { endDateForLength } from '../../cycles/cycleModel';

export const fieldStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};

// The Portal dialog pads its header and footer but not its body.
const bodyClass = 'space-y-3 px-6';

interface StartCycleDialogProps {
  cycle: ApiCycle | null;
  /** The project's active cycle, if any: a second one cannot be started. */
  activeCycle?: ApiCycle;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onStart: (input: { lengthWeeks: number; goal: string }) => void;
}

export const StartCycleDialog: React.FC<StartCycleDialogProps> = ({
  cycle,
  activeCycle,
  pending,
  onOpenChange,
  onStart,
}) => {
  const { t } = useTranslation();
  const goalId = useId();
  const [lengthWeeks, setLengthWeeks] = useState('2');
  const [goal, setGoal] = useState('');

  useEffect(() => {
    if (cycle) {
      setLengthWeeks('2');
      setGoal(cycle.goal ?? '');
    }
  }, [cycle]);

  const blocked = activeCycle !== undefined && activeCycle.id !== cycle?.id;

  return (
    <Dialog open={cycle !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('flow.backlog.start.title', { name: cycle?.name ?? '' })}</DialogTitle>
          <DialogDescription>{t('flow.backlog.start.description')}</DialogDescription>
        </DialogHeader>
        {blocked ? (
          <p role="alert" className="px-6 text-flow-meta" style={{ color: 'var(--flow-warn-fg)' }}>
            {t('flow.backlog.start.alreadyActive', { name: activeCycle?.name ?? '' })}
          </p>
        ) : (
          <div className={bodyClass}>
            <div className="space-y-1">
              <span className="text-flow-meta font-medium" style={{ color: 'var(--flow-text)' }}>
                {t('flow.backlog.start.length')}
              </span>
              <FlowSelect
                label={t('flow.backlog.start.length')}
                value={lengthWeeks}
                options={CYCLE_LENGTH_WEEKS.map((weeks) => ({
                  value: String(weeks),
                  label: t('flow.backlog.start.weeks', { count: weeks }),
                }))}
                onChange={setLengthWeeks}
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
        )}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('flow.backlog.cancel')}
          </Button>
          <Button
            disabled={blocked || pending}
            onClick={() => onStart({ lengthWeeks: Number(lengthWeeks), goal: goal.trim() })}
          >
            {pending ? t('flow.backlog.start.starting') : t('flow.backlog.start.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

type RowChoice = 'next' | 'backlog' | 'cancel';

interface CompleteCycleDialogProps {
  cycle: ApiCycle | null;
  /** The cycle's unfinished tasks, in plan order. */
  unfinished: Task[];
  /** Where "next cycle" sends a task; none means the option is not offered. */
  nextCycle?: ApiCycle;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onComplete: (decisions: ApiCycleTaskDecision[]) => void;
}

/**
 * Lists every unfinished task with a choice: the next cycle (the default when
 * there is one), the backlog, or cancel. The server applies them in one write.
 */
export const CompleteCycleDialog: React.FC<CompleteCycleDialogProps> = ({
  cycle,
  unfinished,
  nextCycle,
  pending,
  onOpenChange,
  onComplete,
}) => {
  const { t } = useTranslation();
  const [choices, setChoices] = useState<Record<string, RowChoice>>({});
  const fallback: RowChoice = nextCycle ? 'next' : 'backlog';

  useEffect(() => {
    if (cycle) setChoices({});
  }, [cycle]);

  const options = [
    ...(nextCycle
      ? [{ value: 'next', label: t('flow.backlog.complete.toCycle', { name: nextCycle.name }) }]
      : []),
    { value: 'backlog', label: t('flow.backlog.complete.toBacklog') },
    { value: 'cancel', label: t('flow.backlog.complete.cancelTask') },
  ];

  const submit = () => {
    onComplete(
      unfinished.map((task) => {
        const choice = choices[task.id] ?? fallback;
        if (choice === 'next' && nextCycle) {
          return { taskId: task.id, decision: 'move', toCycleId: nextCycle.id };
        }
        return { taskId: task.id, decision: choice === 'cancel' ? 'cancel' : 'backlog' };
      })
    );
  };

  return (
    <Dialog open={cycle !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('flow.backlog.complete.title', { name: cycle?.name ?? '' })}</DialogTitle>
          <DialogDescription>
            {unfinished.length === 0
              ? t('flow.backlog.complete.allDone')
              : t('flow.backlog.complete.description', { count: unfinished.length })}
          </DialogDescription>
        </DialogHeader>
        {unfinished.length > 0 && (
          <ul className="max-h-80 space-y-1.5 overflow-y-auto px-6">
            {unfinished.map((task) => (
              <li key={task.id} className="flex items-center gap-2">
                <span
                  dir="auto"
                  className="min-w-0 flex-1 truncate text-flow-meta"
                  style={{ color: 'var(--flow-text)' }}
                  title={task.title}
                >
                  {task.title}
                </span>
                <FlowSelect
                  label={t('flow.backlog.complete.rowLabel', { title: task.title })}
                  value={choices[task.id] ?? fallback}
                  options={options}
                  onChange={(value) =>
                    setChoices((current) => ({ ...current, [task.id]: value as RowChoice }))
                  }
                  className="w-44 shrink-0"
                />
              </li>
            ))}
          </ul>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('flow.backlog.cancel')}
          </Button>
          <Button disabled={pending} onClick={submit}>
            {pending ? t('flow.backlog.complete.completing') : t('flow.backlog.complete.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

interface CycleLengthPresetsProps {
  /** The preset in use, or null when the dates were picked by hand. */
  weeks: number | null;
  /** Show "Custom dates" when no preset is in use. */
  showCustom: boolean;
  onPick: (weeks: number) => void;
}

/** 1 to 4 week buttons that set a cycle's end date from its start. */
export const CycleLengthPresets: React.FC<CycleLengthPresetsProps> = ({
  weeks,
  showCustom,
  onPick,
}) => {
  const { t } = useTranslation();
  return (
    <div
      className="flex flex-wrap items-center gap-1.5"
      role="group"
      aria-label={t('flow.cycles.length')}
    >
      <span className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.cycles.length')}
      </span>
      {CYCLE_LENGTH_WEEKS.map((option) => (
        <button
          key={option}
          type="button"
          aria-pressed={weeks === option}
          onClick={() => onPick(option)}
          className="flow-focus px-2 py-1 rounded-sm text-flow-meta"
          style={{
            border: '1px solid var(--flow-border-control)',
            background: weeks === option ? 'var(--accent-muted)' : 'transparent',
            color: 'var(--flow-text)',
          }}
        >
          {t('flow.backlog.start.weeks', { count: option })}
        </button>
      ))}
      {weeks === null && showCustom && (
        <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.cycles.customLength')}
        </span>
      )}
    </div>
  );
};

/** The preset a pair of dates matches, or null for a custom length. */
export function presetWeeks(startDate: string, endDate: string): number | null {
  return CYCLE_LENGTH_WEEKS.find((weeks) => endDateForLength(startDate, weeks) === endDate) ?? null;
}

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

export const EditCycleDialog: React.FC<EditCycleDialogProps> = ({
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
  const [weeks, setWeeks] = useState<number | null>(null);
  const [goal, setGoal] = useState('');

  useEffect(() => {
    if (cycle) {
      setName(cycle.name);
      setStartDate(cycle.startDate);
      setEndDate(cycle.endDate);
      setWeeks(presetWeeks(cycle.startDate, cycle.endDate));
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
        <div className={bodyClass}>
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
              onChange={(value) => {
                setStartDate(value);
                if (weeks !== null) setEndDate(endDateForLength(value, weeks));
              }}
              locale={i18n.language}
              aria-label={t('flow.cycles.startDate')}
              labels={{ placeholder: t('flow.cycles.startDate'), ...datePickerLabels }}
              className="text-flow-meta"
            />
            <DatePicker
              value={endDate}
              onChange={(value) => {
                setEndDate(value);
                setWeeks(presetWeeks(startDate, value));
              }}
              minDate={startDate || undefined}
              locale={i18n.language}
              aria-label={t('flow.cycles.endDate')}
              labels={{ placeholder: t('flow.cycles.endDate'), ...datePickerLabels }}
              className="text-flow-meta"
            />
          </div>
          <CycleLengthPresets
            weeks={weeks}
            showCustom={Boolean(startDate && endDate)}
            onPick={(picked) => {
              setWeeks(picked);
              if (startDate) setEndDate(endDateForLength(startDate, picked));
            }}
          />
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
