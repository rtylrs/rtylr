// Submission helpers for the recruit app.
// These are pure functions that operate on Submission and FormField
// objects from types.ts. They live outside the React component tree
// so they are testable without a DOM or query client.

import type { Submission, FormField, FormFieldKind } from './types';

// Submission status labels shown in the admin table and candidate
// emails. Keep in sync with the Go recruiter service constants.
export type SubmissionStatus = 'pending' | 'reviewing' | 'shortlisted' | 'rejected' | 'hired';

const STATUS_LABELS: Record<SubmissionStatus, string> = {
  pending: 'Pending',
  reviewing: 'Reviewing',
  shortlisted: 'Shortlisted',
  rejected: 'Rejected',
  hired: 'Hired',
};

const STATUS_ORDER: SubmissionStatus[] = [
  'pending',
  'reviewing',
  'shortlisted',
  'rejected',
  'hired',
];

/**
 * getStatusLabel returns a human-readable label for a submission status.
 * Unknown statuses fall back to capitalizing the raw value so legacy
 * data doesn't show as a blank cell.
 */
export function getStatusLabel(status: string): string {
  if (status in STATUS_LABELS) {
    return STATUS_LABELS[status as SubmissionStatus];
  }
  return status.charAt(0).toUpperCase() + status.slice(1);
}

/**
 * isTerminalStatus returns true when the submission has been definitively
 * resolved. Used to disable the "Change status" button in the admin UI.
 */
export function isTerminalStatus(status: string): boolean {
  return status === 'rejected' || status === 'hired';
}

/**
 * sortSubmissions sorts by status priority (pending first, terminal last)
 * then by createdAt descending within the same status bucket.
 */
export function sortSubmissions(submissions: Submission[]): Submission[] {
  return [...submissions].sort((a, b) => {
    const ai = STATUS_ORDER.indexOf(a.status as SubmissionStatus);
    const bi = STATUS_ORDER.indexOf(b.status as SubmissionStatus);
    const aIdx = ai === -1 ? STATUS_ORDER.length : ai;
    const bIdx = bi === -1 ? STATUS_ORDER.length : bi;
    if (aIdx !== bIdx) return aIdx - bIdx;
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
  });
}

/**
 * parseAnswersJson safely decodes the answersJson field on a Submission.
 * The wire sends a JSON-stringified object; bad data or a double-encoded
 * string both return an empty object so the UI degrades gracefully.
 */
export function parseAnswersJson(answersJson: string): Record<string, unknown> {
  if (!answersJson) return {};
  try {
    const parsed = JSON.parse(answersJson);
    if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>;
    }
    return {};
  } catch {
    return {};
  }
}

/**
 * formatAnswer renders a single answer value into a displayable string.
 * Handles the multi_select case (comma-join) and boolean case.
 */
export function formatAnswer(value: unknown, kind: FormFieldKind | string): string {
  if (value === null || value === undefined || value === '') return '-';
  if (kind === 'boolean') return value === true || value === 'true' ? 'Yes' : 'No';
  if (kind === 'multi_select' && Array.isArray(value)) return value.join(', ');
  return String(value);
}

/**
 * getRequiredFields returns FormField objects that are marked required.
 * Sorted by position so they appear in the order the form builder chose.
 */
export function getRequiredFields(fields: FormField[]): FormField[] {
  return [...fields].filter((f) => f.required).sort((a, b) => a.position - b.position);
}

/**
 * parseOptionsJson safely decodes the optionsJson field on a FormField.
 * Returns an empty array when the field is empty or malformed.
 */
export function parseOptionsJson(optionsJson: string): string[] {
  if (!optionsJson) return [];
  try {
    const parsed = JSON.parse(optionsJson);
    if (Array.isArray(parsed)) {
      return parsed.map(String);
    }
    return [];
  } catch {
    return [];
  }
}
