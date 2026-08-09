// Email notification template helpers for the recruit app.
// The notification system sends candidate confirmation and status
// change emails using these helpers to compose subject lines and
// preview text from the available submission/career data.

/**
 * applicationReceivedSubject returns the email subject line for a
 * candidate confirmation email.
 */
export function applicationReceivedSubject(careerTitle: string, brandName: string): string {
  return `Your application to ${brandName} - ${careerTitle}`;
}

/**
 * statusUpdateSubject returns the subject line for a status change
 * email sent to a candidate.
 */
export function statusUpdateSubject(careerTitle: string, brandName: string): string {
  return `Update on your application - ${careerTitle} at ${brandName}`;
}

/**
 * candidatePreviewText returns a short preview blurb (used in email
 * clients that show it below the subject) for a received application.
 */
export function candidatePreviewText(candidateName: string, brandName: string): string {
  const first = candidateName.trim().split(/\s+/)[0] ?? candidateName;
  return `Hi ${first}, we have received your application at ${brandName} and will be in touch soon.`;
}

/**
 * hrNotificationSubject returns the internal notification subject sent
 * to the HR inbox when a new application arrives.
 */
export function hrNotificationSubject(careerTitle: string, candidateName: string): string {
  return `New application: ${candidateName} for ${careerTitle}`;
}

/**
 * statusLabel returns the human-readable string for a submission status
 * suitable for use in email copy.
 */
export function emailStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: 'received',
    reviewing: 'under review',
    shortlisted: 'shortlisted',
    rejected: 'not moving forward',
    hired: 'selected',
  };
  return labels[status] ?? status;
}

/**
 * isNotificationStatus returns true for statuses that should trigger
 * an outbound candidate email. We skip 'pending' because the initial
 * confirmation is sent separately via applicationReceivedSubject.
 */
export function isNotificationStatus(status: string): boolean {
  return ['reviewing', 'shortlisted', 'rejected', 'hired'].includes(status);
}
