// Business goal tracking helpers for the insights analytics feature.

export type GoalStatus = 'on_track' | 'at_risk' | 'off_track' | 'achieved';

export interface Goal {
  id: string;
  name: string;
  target: number;
  current: number;
  deadline: string;
}

/**
 * goalProgress returns the percentage of the target reached.
 */
export function goalProgress(goal: Goal): number {
  if (goal.target === 0) return 0;
  return Math.min(100, Math.round((goal.current / goal.target) * 1000) / 10);
}

/**
 * goalStatus returns the status of a goal based on progress and remaining time.
 * Expects deadline in YYYY-MM-DD format and referenceDate as epoch milliseconds.
 */
export function goalStatus(goal: Goal, referenceDate: number = Date.now()): GoalStatus {
  const progress = goalProgress(goal);
  if (progress >= 100) return 'achieved';

  const deadline = new Date(goal.deadline).getTime();
  const totalTime = deadline - referenceDate;
  if (totalTime <= 0) return 'off_track';

  if (progress >= 70) return 'on_track';
  if (progress >= 40) return 'at_risk';
  return 'off_track';
}

/**
 * remainingToTarget returns how much more is needed to reach the target.
 */
export function remainingToTarget(goal: Goal): number {
  return Math.max(0, goal.target - goal.current);
}

/**
 * goalStatusLabel returns a display label for a goal status.
 */
export function goalStatusLabel(status: GoalStatus): string {
  const labels: Record<GoalStatus, string> = {
    on_track: 'On Track',
    at_risk: 'At Risk',
    off_track: 'Off Track',
    achieved: 'Achieved',
  };
  return labels[status];
}

/**
 * goalsAchieved returns goals that have been fully achieved.
 */
export function goalsAchieved(goals: Goal[]): Goal[] {
  return goals.filter((g) => goalProgress(g) >= 100);
}

/**
 * goalsAtRisk returns goals that are at risk or off track.
 */
export function goalsAtRisk(goals: Goal[]): Goal[] {
  return goals.filter((g) => {
    const status = goalStatus(g);
    return status === 'at_risk' || status === 'off_track';
  });
}
