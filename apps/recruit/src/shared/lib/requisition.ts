// Job requisition helpers for the recruit admin panel.

export type RequisitionStatus = 'draft' | 'open' | 'on_hold' | 'filled' | 'cancelled';
export type RequisitionPriority = 'low' | 'medium' | 'high' | 'urgent';

export interface JobRequisition {
  id: string;
  title: string;
  department: string;
  status: RequisitionStatus;
  priority: RequisitionPriority;
  openedAt: string;
  filledAt?: string;
}

/**
 * requisitionStatusLabel returns a display-friendly label for a requisition status.
 */
export function requisitionStatusLabel(status: RequisitionStatus | string): string {
  const labels: Record<string, string> = {
    draft: 'Draft',
    open: 'Open',
    on_hold: 'On Hold',
    filled: 'Filled',
    cancelled: 'Cancelled',
  };
  return labels[status] ?? status;
}

/**
 * requisitionPriorityLabel returns a display label for a priority level.
 */
export function requisitionPriorityLabel(priority: RequisitionPriority | string): string {
  const labels: Record<string, string> = {
    low: 'Low',
    medium: 'Medium',
    high: 'High',
    urgent: 'Urgent',
  };
  return labels[priority] ?? priority;
}

/**
 * openRequisitions returns requisitions with open status.
 */
export function openRequisitions(requisitions: JobRequisition[]): JobRequisition[] {
  return requisitions.filter((r) => r.status === 'open');
}

/**
 * requisitionsByDepartment groups requisitions by department.
 */
export function requisitionsByDepartment(
  requisitions: JobRequisition[]
): Record<string, JobRequisition[]> {
  return requisitions.reduce<Record<string, JobRequisition[]>>((acc, r) => {
    acc[r.department] = [...(acc[r.department] ?? []), r];
    return acc;
  }, {});
}

/**
 * urgentOpenRequisitions returns open requisitions with urgent priority.
 */
export function urgentOpenRequisitions(requisitions: JobRequisition[]): JobRequisition[] {
  return requisitions.filter((r) => r.status === 'open' && r.priority === 'urgent');
}

/**
 * timeToFill returns days between openedAt and filledAt for a filled requisition.
 */
export function timeToFill(requisition: JobRequisition): number | null {
  if (!requisition.filledAt) return null;
  const opened = new Date(requisition.openedAt).getTime();
  const filled = new Date(requisition.filledAt).getTime();
  return Math.round((filled - opened) / (1000 * 60 * 60 * 24));
}
