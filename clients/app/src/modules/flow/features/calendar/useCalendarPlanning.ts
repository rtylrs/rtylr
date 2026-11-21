import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import type { ApiCycle, ApiMilestone } from '../../shared/services/flowApi';
import { cyclesQueryOptions } from '../cycles/useCycles';
import { useMilestones } from '../milestones/useMilestones';
import { useApprovedLeaveInRange } from '../../shared/hooks/queries/useApprovedLeaveInRange';
import { daysInRange, isoDay, type DayRange } from '../../shared/lib/calendarSpans';

/** A cycle drawn as a bar behind the calendar's tasks. */
export interface PlanningCycle {
  id: string;
  projectId: string;
  name: string;
  range: DayRange;
  status: ApiCycle['status'];
}

/** A milestone drawn as a marker on its target day. */
export interface PlanningMilestone {
  id: string;
  name: string;
  day: string;
  status: ApiMilestone['status'];
}

/** Approved leave on one day, as far as HR lets the caller see it. */
export interface LeaveDay {
  /** Distinct people on leave that day. */
  people: number;
}

/**
 * Cycles are project scoped on v2 and there is no business-wide listing, so the
 * Plan calendar asks once per project. Capped so a business with a hundred
 * projects does not fire a hundred requests on every month change; the caller
 * passes the projects that actually have work on screen first.
 */
export const MAX_CYCLE_PROJECTS = 20;

/**
 * Module level so useQueries can keep the combined array stable between
 * renders: an inline combine is a new function every render and re-runs.
 */
const pickData = (results: { data?: ApiCycle[] }[]): (ApiCycle[] | undefined)[] =>
  results.map((result) => result.data);

export function toPlanningCycles(projectId: string, cycles: readonly ApiCycle[]): PlanningCycle[] {
  const out: PlanningCycle[] = [];
  for (const cycle of cycles) {
    const start = isoDay(cycle.startDate);
    const end = isoDay(cycle.endDate);
    if (!start || !end) continue;
    out.push({
      id: String(cycle.id),
      projectId,
      name: cycle.name,
      range: start <= end ? { start, end } : { start: end, end: start },
      status: cycle.status,
    });
  }
  return out;
}

export function toPlanningMilestones(milestones: readonly ApiMilestone[]): PlanningMilestone[] {
  const out: PlanningMilestone[] = [];
  for (const milestone of milestones) {
    const day = isoDay(milestone.targetDate);
    if (!day) continue;
    out.push({ id: String(milestone.id), name: milestone.name, day, status: milestone.status });
  }
  return out;
}

/** Folds approved leave requests into one entry per day of `[from, to]`. */
export function leaveByDay(
  requests: readonly { employeeId?: string; startDate?: string; endDate?: string }[],
  from: string,
  to: string
): Map<string, LeaveDay> {
  const people = new Map<string, Set<string>>();
  for (const request of requests) {
    const start = isoDay(request.startDate);
    const end = isoDay(request.endDate) ?? start;
    if (!start || !end || end < start) continue;
    for (const day of daysInRange({ start, end }, from, to)) {
      const bucket = people.get(day) ?? new Set<string>();
      bucket.add(String(request.employeeId ?? ''));
      people.set(day, bucket);
    }
  }
  const out = new Map<string, LeaveDay>();
  for (const [day, bucket] of people) {
    out.set(day, { people: bucket.size });
  }
  return out;
}

/**
 * Everything the calendars draw behind the tasks for `[from, to]`: cycle bars,
 * milestone markers and approved leave.
 *
 * `cycleProjectIds` are the projects whose cycles to fetch. `milestoneProjectId`
 * narrows milestones to one project; undefined lists them business-wide, which
 * the v2 milestones endpoint supports in one request.
 *
 * Every source is optional to the screen: a failed request leaves its layer
 * out rather than failing the calendar.
 */
export function useCalendarPlanning({
  cycleProjectIds,
  milestoneProjectId,
  from,
  to,
}: {
  cycleProjectIds: readonly string[];
  milestoneProjectId: string | undefined;
  from: string;
  to: string;
}): { cycles: PlanningCycle[]; milestones: PlanningMilestone[]; leave: Map<string, LeaveDay> } {
  const ids = useMemo(
    () => [...new Set(cycleProjectIds.filter(Boolean))].slice(0, MAX_CYCLE_PROJECTS),
    [cycleProjectIds]
  );
  const cycleData = useQueries({
    queries: ids.map((id) => cyclesQueryOptions(id)),
    combine: pickData,
  });
  const milestonesQuery = useMilestones(milestoneProjectId);
  // The read TaskDetail already makes for its time-off warning. HR narrows it
  // server side: a leave approver gets the team, anyone else only their own,
  // and a business without HR (or a caller with no employee row) gets null.
  const leaveQuery = useApprovedLeaveInRange(from, to, from !== '' && to !== '');

  const cycles = useMemo(
    () =>
      ids
        .flatMap((id, index) => toPlanningCycles(id, cycleData[index] ?? []))
        .filter((cycle) => cycle.range.start <= to && cycle.range.end >= from),
    [ids, cycleData, from, to]
  );

  const milestones = useMemo(
    () =>
      toPlanningMilestones(milestonesQuery.data ?? []).filter(
        (milestone) => milestone.day >= from && milestone.day <= to
      ),
    [milestonesQuery.data, from, to]
  );

  const leave = useMemo(
    () => leaveByDay(leaveQuery.data ?? [], from, to),
    [leaveQuery.data, from, to]
  );

  return { cycles, milestones, leave };
}
