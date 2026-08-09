// Domain constants for the recruit app.
// Centralizes magic strings and configuration used across components and hooks.

// Submission status values that can be set via the admin UI.
export const SUBMISSION_STATUSES = [
  'pending',
  'reviewing',
  'shortlisted',
  'rejected',
  'hired',
] as const;

export type SubmissionStatus = (typeof SUBMISSION_STATUSES)[number];

// Labels shown in the admin submissions table.
export const SUBMISSION_STATUS_LABELS: Record<SubmissionStatus, string> = {
  pending: 'Pending',
  reviewing: 'Reviewing',
  shortlisted: 'Shortlisted',
  rejected: 'Rejected',
  hired: 'Hired',
};

// Hex colors for status badges in the admin table.
export const SUBMISSION_STATUS_COLORS: Record<SubmissionStatus, string> = {
  pending: '#f59e0b',
  reviewing: '#3b82f6',
  shortlisted: '#8b5cf6',
  rejected: '#ef4444',
  hired: '#10b981',
};

// Known seniority levels in ascending order.
export const SENIORITY_LEVELS = [
  'intern',
  'junior',
  'mid',
  'senior',
  'lead',
  'principal',
  'executive',
] as const;

export type SeniorityLevel = (typeof SENIORITY_LEVELS)[number];

export const SENIORITY_LABELS: Record<SeniorityLevel, string> = {
  intern: 'Intern',
  junior: 'Junior',
  mid: 'Mid-level',
  senior: 'Senior',
  lead: 'Lead',
  principal: 'Principal',
  executive: 'Executive',
};

// Available work modes.
export const WORK_MODES = ['remote', 'hybrid', 'onsite'] as const;
export type WorkMode = (typeof WORK_MODES)[number];

export const WORK_MODE_LABELS: Record<WorkMode, string> = {
  remote: 'Remote',
  hybrid: 'Hybrid',
  onsite: 'On-site',
};

// Available employment types.
export const EMPLOYMENT_TYPES = [
  'full-time',
  'part-time',
  'contract',
  'freelance',
  'internship',
] as const;

export type EmploymentType = (typeof EMPLOYMENT_TYPES)[number];

// Available career page templates.
export const TEMPLATE_IDS = ['minimal', 'editorial', 'bold'] as const;

// Max file size (in bytes) for resume uploads displayed in the UI.
export const MAX_RESUME_SIZE_BYTES = 5 * 1024 * 1024; // 5 MB

// Accepted MIME types for resume uploads.
export const ACCEPTED_RESUME_TYPES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
] as const;
