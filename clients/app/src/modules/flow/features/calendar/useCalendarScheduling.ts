import { useEffect, useMemo, useRef, useState } from 'react';
import { useUpdateTask } from '../tasks/useTasks';
import { reschedule, type DragKind, type SchedulePatch } from '../../shared/lib/calendarSpans';

/** A drag in progress: which task, which part of it, and the day it was picked up on. */
export interface ActiveDrag {
  taskId: string;
  kind: DragKind;
  grabDay?: string;
}

export interface CalendarScheduling<T> {
  /** The tasks with any drop still on its way to the server already applied. */
  tasks: T[];
  /** The drag in progress, or null. */
  active: ActiveDrag | null;
  begin: (drag: ActiveDrag) => void;
  end: () => void;
  /** Drops the active drag on `day`. Returns whether it sent an update. */
  drop: (day: string) => boolean;
}

type Schedulable = { id: string | number; dueDate?: string | null; startDate?: string | null };

/**
 * Drag to reschedule, shared by the grid and the Unscheduled panel.
 *
 * The drag lives in state here rather than in `dataTransfer`, which a browser
 * only lets a drop handler read, so a day cell cannot tell during dragover what
 * is coming. Permission is the caller's: a task the user cannot edit is never
 * made draggable, so it never reaches `begin`.
 *
 * A drop is drawn at once. The new dates sit over the task until the task list
 * refetches (the update invalidates it), and come off straight away if the
 * update fails, which useUpdateTask already reports with a toast.
 */
export function useCalendarScheduling<T extends Schedulable>(
  source: readonly T[]
): CalendarScheduling<T> {
  const updateTask = useUpdateTask();
  const [active, setActive] = useState<ActiveDrag | null>(null);
  // Read in drop, which can fire in the same tick as dragstart's state update.
  const activeRef = useRef<ActiveDrag | null>(null);
  const [overrides, setOverrides] = useState<ReadonlyMap<string, SchedulePatch>>(new Map());

  // A fresh task list carries the server's dates, so the drawn ones step aside.
  useEffect(() => {
    setOverrides((current) => (current.size === 0 ? current : new Map()));
  }, [source]);

  const tasks = useMemo(
    () =>
      overrides.size === 0
        ? [...source]
        : source.map((task) => {
            const patch = overrides.get(String(task.id));
            return patch ? { ...task, ...patch } : task;
          }),
    [source, overrides]
  );

  const begin = (drag: ActiveDrag): void => {
    activeRef.current = drag;
    setActive(drag);
  };

  const end = (): void => {
    activeRef.current = null;
    setActive(null);
  };

  const drop = (day: string): boolean => {
    const drag = activeRef.current;
    end();
    if (!drag) return false;
    const task = tasks.find((candidate) => String(candidate.id) === drag.taskId);
    if (!task) return false;
    const patch = reschedule(task, drag.kind, day, drag.grabDay);
    if (!patch) return false;
    const id = drag.taskId;
    setOverrides((current) => new Map(current).set(id, patch));
    updateTask.mutate(
      { id, data: patch },
      {
        onError: () =>
          setOverrides((current) => {
            const next = new Map(current);
            next.delete(id);
            return next;
          }),
      }
    );
    return true;
  };

  return { tasks, active, begin, end, drop };
}
