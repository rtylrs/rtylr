// Interview scheduling helpers for the recruit admin panel.

export type InterviewType = 'phone' | 'video' | 'onsite' | 'technical' | 'panel';

const INTERVIEW_TYPE_LABELS: Record<InterviewType, string> = {
  phone: 'Phone Screen',
  video: 'Video Call',
  onsite: 'On-site',
  technical: 'Technical',
  panel: 'Panel',
};

const DEFAULT_DURATIONS: Record<InterviewType, number> = {
  phone: 30,
  video: 45,
  onsite: 60,
  technical: 90,
  panel: 60,
};

/**
 * interviewTypeLabel returns a human-readable label for an interview type.
 */
export function interviewTypeLabel(type: InterviewType | string): string {
  return (INTERVIEW_TYPE_LABELS as Record<string, string>)[type] ?? type;
}

/**
 * defaultDuration returns the default duration in minutes for an interview type.
 */
export function defaultDuration(type: InterviewType | string): number {
  return (DEFAULT_DURATIONS as Record<string, number>)[type] ?? 60;
}

/**
 * allInterviewTypes returns all supported interview types.
 */
export function allInterviewTypes(): InterviewType[] {
  return ['phone', 'video', 'onsite', 'technical', 'panel'];
}

/**
 * formatDuration returns a human-readable duration string.
 */
export function formatDuration(minutes: number): string {
  if (minutes < 60) return `${minutes}min`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return m === 0 ? `${h}h` : `${h}h ${m}min`;
}

/**
 * isValidMeetingUrl returns true when the URL looks like a valid video meeting URL.
 */
export function isValidMeetingUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    return parsed.protocol === 'https:' && parsed.hostname.length > 0;
  } catch {
    return false;
  }
}
