import React, { useEffect, useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
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

const fieldStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};

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
          <p role="alert" className="text-flow-meta" style={{ color: 'var(--flow-warn-fg)' }}>
            {t('flow.backlog.start.alreadyActive', { name: activeCycle?.name ?? '' })}
          </p>
        ) : (
          <div className="space-y-3">
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
          <ul className="max-h-80 space-y-1.5 overflow-y-auto">
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
