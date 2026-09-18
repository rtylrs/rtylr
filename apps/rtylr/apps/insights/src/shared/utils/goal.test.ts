import { describe, it, expect } from 'vitest';
import {
  goalProgress,
  goalStatus,
  remainingToTarget,
  goalStatusLabel,
  goalsAchieved,
  goalsAtRisk,
} from './goal';
import type { Goal } from './goal';

function makeGoal(name: string, target: number, current: number, deadline = '2030-12-31'): Goal {
  return { id: name, name, target, current, deadline };
}

describe('goalProgress', () => {
  it('returns percentage of target reached', () => {
    expect(goalProgress(makeGoal('A', 100, 80))).toBe(80);
  });

  it('caps at 100 when exceeded', () => {
    expect(goalProgress(makeGoal('A', 100, 120))).toBe(100);
  });

  it('returns 0 when target is 0', () => {
    expect(goalProgress(makeGoal('A', 0, 10))).toBe(0);
  });
});

describe('goalStatus', () => {
  it('returns achieved when at 100%', () => {
    expect(goalStatus(makeGoal('A', 100, 100))).toBe('achieved');
  });

  it('returns on_track when above 70%', () => {
    expect(goalStatus(makeGoal('A', 100, 80))).toBe('on_track');
  });

  it('returns at_risk when between 40% and 70%', () => {
    expect(goalStatus(makeGoal('A', 100, 55))).toBe('at_risk');
  });

  it('returns off_track when below 40%', () => {
    expect(goalStatus(makeGoal('A', 100, 20))).toBe('off_track');
  });
});

describe('remainingToTarget', () => {
  it('returns difference to target', () => {
    expect(remainingToTarget(makeGoal('A', 100, 60))).toBe(40);
  });

  it('returns 0 when target is met', () => {
    expect(remainingToTarget(makeGoal('A', 100, 100))).toBe(0);
  });

  it('returns 0 when exceeded', () => {
    expect(remainingToTarget(makeGoal('A', 100, 110))).toBe(0);
  });
});

describe('goalStatusLabel', () => {
  it('returns On Track label', () => {
    expect(goalStatusLabel('on_track')).toBe('On Track');
  });

  it('returns Achieved label', () => {
    expect(goalStatusLabel('achieved')).toBe('Achieved');
  });
});

describe('goalsAchieved', () => {
  it('returns goals at 100% or above', () => {
    const goals = [makeGoal('A', 100, 100), makeGoal('B', 100, 80)];
    expect(goalsAchieved(goals)).toHaveLength(1);
  });
});

describe('goalsAtRisk', () => {
  it('returns at_risk and off_track goals', () => {
    const goals = [makeGoal('A', 100, 80), makeGoal('B', 100, 55), makeGoal('C', 100, 20)];
    expect(goalsAtRisk(goals)).toHaveLength(2);
  });
});
