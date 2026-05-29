import { describe, it, expect } from 'vitest';
import {
  evaluateThreshold,
  exceedsWarningPct,
  buildThresholdMessage,
  evaluate,
  activeAlerts,
  highestSeverity,
  severityColor,
} from './alertThresholdUtils';
import type { Threshold } from './alertThresholdUtils';

const threshold: Threshold = {
  name: 'Stock Level',
  warningValue: 20,
  criticalValue: 5,
  unit: 'units',
};

describe('evaluateThreshold', () => {
  it('returns critical when value meets criticalValue', () => {
    expect(evaluateThreshold(threshold, 5)).toBe('critical');
    expect(evaluateThreshold(threshold, 3)).toBe('critical');
  });

  it('returns warning when value meets warningValue but not critical', () => {
    expect(evaluateThreshold(threshold, 20)).toBe('warning');
    expect(evaluateThreshold(threshold, 10)).toBe('warning');
  });

  it('returns info when below warningValue', () => {
    expect(evaluateThreshold(threshold, 21)).toBe('info');
    expect(evaluateThreshold(threshold, 100)).toBe('info');
  });
});

describe('exceedsWarningPct', () => {
  it('returns 0 when below warning', () => {
    expect(exceedsWarningPct(threshold, 10)).toBe(0);
  });

  it('returns positive pct when above warning', () => {
    expect(exceedsWarningPct(threshold, 30)).toBeCloseTo(50, 0);
  });
});

describe('buildThresholdMessage', () => {
  it('includes CRITICAL label for critical severity', () => {
    expect(buildThresholdMessage(threshold, 3, 'critical')).toContain('CRITICAL');
  });

  it('includes WARNING label for warning severity', () => {
    expect(buildThresholdMessage(threshold, 15, 'warning')).toContain('WARNING');
  });

  it('includes normal label for info severity', () => {
    expect(buildThresholdMessage(threshold, 50, 'info')).toContain('normal');
  });

  it('includes units in message when unit is set', () => {
    expect(buildThresholdMessage(threshold, 3, 'critical')).toContain('units');
  });
});

describe('evaluate', () => {
  it('returns a full evaluation object', () => {
    const eval_ = evaluate(threshold, 3);
    expect(eval_.severity).toBe('critical');
    expect(eval_.currentValue).toBe(3);
    expect(eval_.message).toContain('CRITICAL');
  });
});

describe('activeAlerts', () => {
  it('filters out info evaluations', () => {
    const evals = [evaluate(threshold, 3), evaluate(threshold, 15), evaluate(threshold, 50)];
    const active = activeAlerts(evals);
    expect(active).toHaveLength(2);
    expect(active.every((e) => e.severity !== 'info')).toBe(true);
  });
});

describe('highestSeverity', () => {
  it('returns critical when any critical exists', () => {
    const evals = [evaluate(threshold, 3), evaluate(threshold, 15)];
    expect(highestSeverity(evals)).toBe('critical');
  });

  it('returns warning when only warnings exist', () => {
    const evals = [evaluate(threshold, 15), evaluate(threshold, 18)];
    expect(highestSeverity(evals)).toBe('warning');
  });

  it('returns info when all are info', () => {
    const evals = [evaluate(threshold, 100)];
    expect(highestSeverity(evals)).toBe('info');
  });
});

describe('severityColor', () => {
  it('returns correct colors', () => {
    expect(severityColor('critical')).toBe('red');
    expect(severityColor('warning')).toBe('yellow');
    expect(severityColor('info')).toBe('green');
  });
});
