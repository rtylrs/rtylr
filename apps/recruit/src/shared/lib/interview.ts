// Interview scheduling helpers for the recruit admin panel.

export type InterviewType = 'phone' | 'video' | 'onsite' | 'technical' | 'panel';

const TYPE_LABELS: Record<InterviewType, string> = {
  phone: 'Phone Screen',
  video: 'Video Interview',
  onsite: 'On-site Interview',
  technical: 'Technical Assessment',
  panel: 'Panel Interview',
};

const TYPE_DURATIONS: Record<InterviewType, number> = {
  phone: 30,
  video: 60,
  onsite: 90,
  technical: 90,
  panel: 60,
};

/**
 * interviewTypeLabel returns a human-readable label for the interview type.
 */
export function interviewTypeLabel(type: InterviewType | string): string {
  return (TYPE_LABELS as Record<string, string>)[type] ?? type;
}

/**
 * defaultDuration returns the default duration in minutes for an interview type.
 */
export function defaultDuration(type: InterviewType | string): number {
  return (TYPE_DURATIONS as Record<string, number>)[type] ?? 60;
}

/**
 * allInterviewTypes returns all available interview types.
 */
export function allInterviewTypes(): InterviewType[] {
  return ['phone', 'video', 'onsite', 'technical', 'panel'];
}

/**
 * formatDuration returns a human-readable duration string, e.g. '1h 30m' or '45m'.
 */
export function formatDuration(minutes: number): string {
  if (minutes < 60) return `${minutes}m`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return m === 0 ? `${h}h` : `${h}h ${m}m`;
}

/**
 * isValidMeetingUrl returns true when the URL appears to be a valid video
 * conference link (Zoom, Meet, Teams, or Calendly).
 */
export function isValidMeetingUrl(url: string): boolean {
  if (!url) return false;
  return /^https?:\/\/([\w-]+\.)?(zoom\.us|meet\.google\.com|teams\.microsoft\.com|calendly\.com)\//i.test(url);
}
