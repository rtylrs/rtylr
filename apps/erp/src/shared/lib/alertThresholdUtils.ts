/**
 * Alert threshold configuration and evaluation helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

export type ThresholdSeverity = 'info' | 'warning' | 'critical';

export interface Threshold {
  name: string;
  warningValue: number;
  criticalValue: number;
  unit?: string;
}

export interface ThresholdEvaluation {
  threshold: Threshold;
  currentValue: number;
  severity: ThresholdSeverity;
  exceedsPct: number;
  message: string;
}

/** Evaluate whether a value exceeds a threshold and return severity. */
export function evaluateThreshold(threshold: Threshold, currentValue: number): ThresholdSeverity {
  if (currentValue <= threshold.criticalValue) return 'critical';
  if (currentValue <= threshold.warningValue) return 'warning';
  return 'info';
}

/** Compute how much the current value exceeds the warning threshold (as %). */
export function exceedsWarningPct(threshold: Threshold, currentValue: number): number {
  if (threshold.warningValue === 0) return 0;
  return Math.max(0, ((currentValue - threshold.warningValue) / threshold.warningValue) * 100);
}

/** Build a human-readable evaluation message. */
export function buildThresholdMessage(
  threshold: Threshold,
  currentValue: number,
  severity: ThresholdSeverity
): string {
  const unit = threshold.unit ? ` ${threshold.unit}` : '';
  switch (severity) {
    case 'critical':
      return `${threshold.name} is CRITICAL at ${currentValue}${unit} (critical: ${threshold.criticalValue}${unit}).`;
    case 'warning':
      return `${threshold.name} is WARNING at ${currentValue}${unit} (warning: ${threshold.warningValue}${unit}).`;
    case 'info':
      return `${threshold.name} is normal at ${currentValue}${unit}.`;
  }
}

/** Evaluate a full ThresholdEvaluation object. */
export function evaluate(threshold: Threshold, currentValue: number): ThresholdEvaluation {
  const severity = evaluateThreshold(threshold, currentValue);
  return {
    threshold,
    currentValue,
    severity,
    exceedsPct: exceedsWarningPct(threshold, currentValue),
    message: buildThresholdMessage(threshold, currentValue, severity),
  };
}

/** Evaluate multiple thresholds and return only those that are warning or critical. */
export function activeAlerts(evaluations: ThresholdEvaluation[]): ThresholdEvaluation[] {
  return evaluations.filter((e) => e.severity !== 'info');
}

/** Return the highest severity across a list of evaluations. */
export function highestSeverity(evaluations: ThresholdEvaluation[]): ThresholdSeverity {
  if (evaluations.some((e) => e.severity === 'critical')) return 'critical';
  if (evaluations.some((e) => e.severity === 'warning')) return 'warning';
  return 'info';
}

/** Return a badge color for a threshold severity. */
export function severityColor(severity: ThresholdSeverity): string {
  switch (severity) {
    case 'critical':
      return 'red';
    case 'warning':
      return 'yellow';
    case 'info':
      return 'green';
  }
}
