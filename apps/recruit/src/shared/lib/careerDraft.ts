// Career draft helpers for the recruit admin panel.

export interface CareerDraftValidation {
  isValid: boolean;
  errors: string[];
}

/**
 * validateCareerDraft returns validation errors for a career draft.
 */
export function validateCareerDraft(draft: {
  title?: string;
  department?: string;
  location?: string;
  details?: string;
  salaryMin?: number;
  salaryMax?: number;
}): CareerDraftValidation {
  const errors: string[] = [];
  if (!draft.title?.trim()) errors.push('Title is required');
  if (!draft.department?.trim()) errors.push('Department is required');
  if (!draft.location?.trim()) errors.push('Location is required');
  if (!draft.details?.trim()) errors.push('Job details are required');
  if (
    draft.salaryMin !== undefined &&
    draft.salaryMax !== undefined &&
    draft.salaryMin > draft.salaryMax
  ) {
    errors.push('Minimum salary cannot exceed maximum salary');
  }
  return { isValid: errors.length === 0, errors };
}

/**
 * draftCompletionPct returns a percentage (0-100) of how complete the draft is.
 */
export function draftCompletionPct(draft: {
  title?: string;
  department?: string;
  location?: string;
  details?: string;
  requirements?: string;
  salaryMin?: number;
  salaryMax?: number;
}): number {
  const fields = [
    Boolean(draft.title?.trim()),
    Boolean(draft.department?.trim()),
    Boolean(draft.location?.trim()),
    Boolean(draft.details?.trim()),
    Boolean(draft.requirements?.trim()),
    draft.salaryMin !== undefined,
    draft.salaryMax !== undefined,
  ];
  const filled = fields.filter(Boolean).length;
  return Math.round((filled / fields.length) * 100);
}

/**
 * salaryRangeLabel returns a formatted salary range string.
 */
export function salaryRangeLabel(
  min: number | undefined,
  max: number | undefined,
  currency = 'USD'
): string {
  if (min === undefined && max === undefined) return 'Competitive';
  if (min !== undefined && max !== undefined)
    return `${currency} ${min.toLocaleString()} - ${max.toLocaleString()}`;
  if (min !== undefined) return `${currency} ${min.toLocaleString()}+`;
  return `Up to ${currency} ${max!.toLocaleString()}`;
}
