// Report helpers for the insights reports feature.

import type { ReportTemplate, ScheduledReport } from '../hooks/queries/types';

/**
 * activeScheduledReports returns only reports where isActive is true.
 */
export function activeScheduledReports(reports: ScheduledReport[]): ScheduledReport[] {
  return reports.filter((r) => r.isActive);
}

/**
 * reportsByTemplate groups scheduled reports by their templateId.
 */
export function reportsByTemplate(reports: ScheduledReport[]): Map<string, ScheduledReport[]> {
  const map = new Map<string, ScheduledReport[]>();
  for (const report of reports) {
    const existing = map.get(report.templateId) ?? [];
    map.set(report.templateId, [...existing, report]);
  }
  return map;
}

/**
 * cadenceLabel returns a human-readable label for a report cadence.
 */
export function cadenceLabel(cadence: string): string {
  const labels: Record<string, string> = {
    daily: 'Daily',
    weekly: 'Weekly',
    monthly: 'Monthly',
    quarterly: 'Quarterly',
  };
  return labels[cadence] ?? cadence;
}

/**
 * reportChannelLabel returns a human-readable label for a delivery channel.
 */
export function reportChannelLabel(channel: string): string {
  const labels: Record<string, string> = {
    email: 'Email',
    slack: 'Slack',
    webhook: 'Webhook',
  };
  return labels[channel] ?? channel;
}

/**
 * findTemplate returns the template matching the given id, or undefined.
 */
export function findTemplate(templates: ReportTemplate[], id: string): ReportTemplate | undefined {
  return templates.find((t) => t.id === id);
}

/**
 * sortScheduledReports returns reports sorted by createdAt descending.
 */
export function sortScheduledReports(reports: ScheduledReport[]): ScheduledReport[] {
  return [...reports].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}
