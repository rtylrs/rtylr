/**
 * HR request statuses arrive from the API in lowercase. Keep comparisons
 * tolerant of older title-cased fixtures while treating the wire value as
 * the canonical form.
 */
export function normalizeHrRequestStatus(status: string | null | undefined): string {
  return status?.trim().toLowerCase() ?? '';
}

export function isHrRequestStatus(status: string | null | undefined, expected: string): boolean {
  return normalizeHrRequestStatus(status) === normalizeHrRequestStatus(expected);
}
