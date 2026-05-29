// Candidate notification helpers for the recruit admin panel.

export type NotificationEvent =
  | 'application_received'
  | 'status_changed'
  | 'interview_scheduled'
  | 'offer_sent'
  | 'offer_accepted'
  | 'offer_rejected';

const EVENT_LABELS: Record<NotificationEvent, string> = {
  application_received: 'Application Received',
  status_changed: 'Status Changed',
  interview_scheduled: 'Interview Scheduled',
  offer_sent: 'Offer Sent',
  offer_accepted: 'Offer Accepted',
  offer_rejected: 'Offer Rejected',
};

const EVENT_TEMPLATES: Record<NotificationEvent, string> = {
  application_received: 'Thank you for applying to {{position}}. We will review your application.',
  status_changed: 'Your application status has been updated to {{status}}.',
  interview_scheduled: 'Your interview for {{position}} is scheduled for {{date}}.',
  offer_sent: 'Congratulations! We have sent you an offer for {{position}}.',
  offer_accepted: 'You have accepted our offer for {{position}}. Welcome to the team!',
  offer_rejected: 'We understand you have declined our offer for {{position}}.',
};

/**
 * notificationLabel returns a human-readable label for a notification event.
 */
export function notificationLabel(event: NotificationEvent | string): string {
  return (EVENT_LABELS as Record<string, string>)[event] ?? event;
}

/**
 * notificationTemplate returns the default template for a notification event.
 */
export function notificationTemplate(event: NotificationEvent | string): string {
  return (EVENT_TEMPLATES as Record<string, string>)[event] ?? '';
}

/**
 * interpolateTemplate replaces {{key}} placeholders with values from the variables map.
 */
export function interpolateTemplate(
  template: string,
  variables: Record<string, string>
): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => variables[key] ?? `{{${key}}}`);
}

/**
 * allNotificationEvents returns all supported notification events.
 */
export function allNotificationEvents(): NotificationEvent[] {
  return [
    'application_received',
    'status_changed',
    'interview_scheduled',
    'offer_sent',
    'offer_accepted',
    'offer_rejected',
  ];
}
