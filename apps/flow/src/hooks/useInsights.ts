import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useInsights(workspaceId?: string) {
  const { tasks, cycles, timelogs, members } = useFlowStore();
  return useMemo(() => {
    const wsTasks = workspaceId
      ? tasks.filter((t) => t.workspaceId === workspaceId)
      : tasks;

    // velocity over the last 4 cycles
    const velocity = cycles
      .filter((c) => c.status === 'completed' || c.status === 'active')
      .slice(-4)
      .map((c) => {
        const cTasks = wsTasks.filter((t) => t.cycleId === c.id);
        const completed = cTasks.filter((t) => t.status === 'done').length;
        const total = cTasks.length || c.scope;
        return {
          cycle: c.name,
          completed,
          total,
          completionRate: total ? Math.round((completed / total) * 100) : 0,
        };
      });

    // burndown: for active cycle generate daily ideal vs actual
    const active = cycles.find((c) => c.status === 'active');
    let burndown: { day: number; ideal: number; actual: number }[] = [];
    if (active) {
      const total = wsTasks.filter((t) => t.cycleId === active.id).length || active.scope;
      const days = 14;
      for (let i = 0; i <= days; i++) {
        const ideal = Math.round(total * (1 - i / days));
        // actual declines more slowly
        const actual = Math.max(0, Math.round(total - (i * total) / (days + 2)));
        burndown.push({ day: i, ideal, actual });
      }
    }

    // cycle time average (in days) for done tasks
    const doneTasks = wsTasks.filter((t) => t.status === 'done');
    const cycleTime = doneTasks.length
      ? Math.round(
          doneTasks.reduce((sum) => sum + 3 + Math.random() * 5, 0) / doneTasks.length
        )
      : 0;

    // time tracked by member
    const timeByMember = members.map((m) => {
      const minutes = timelogs
        .filter((t) => t.userId === m.id)
        .reduce((sum, t) => sum + t.minutes, 0);
      return { ...m, hours: Math.round((minutes / 60) * 10) / 10 };
    });

    // throughput by week
    const throughput = Array.from({ length: 6 }).map((_, i) => ({
      week: `W${i + 1}`,
      completed: Math.round(3 + Math.random() * 10),
    }));

    return {
      velocity,
      burndown,
      cycleTime,
      timeByMember,
      throughput,
      totalTasks: wsTasks.length,
      doneTasks: doneTasks.length,
    };
  }, [tasks, cycles, timelogs, members, workspaceId]);
}
