/** The stable codes the pager service sends next to an error message. */
export const PAGER_ERROR = {
  venueNotFound: 'venue_not_found',
  ticketNotFound: 'ticket_not_found',
  invalidLabel: 'invalid_label',
  labelTaken: 'label_taken',
  invalidTransition: 'invalid_transition',
  tooManyOpenTickets: 'too_many_open_tickets',
  joinByNumberOff: 'join_by_number_off',
  typeYourNumber: 'type_your_number',
  invalidReviewUrl: 'invalid_review_url',
  invalidPrepMinutes: 'invalid_prep_minutes',
} as const;

/** Generated clients throw an `ApiError` carrying `statusCode` and `code`. */
export function errorStatus(error: unknown): number | undefined {
  const status = (error as { statusCode?: unknown } | null)?.statusCode;
  return typeof status === 'number' ? status : undefined;
}

export function errorCode(error: unknown): string | undefined {
  const code = (error as { code?: unknown } | null)?.code;
  return typeof code === 'string' ? code : undefined;
}

/** Five tries 1, 2, 4 and 8 seconds apart: about 20 seconds with the requests. */
export const DEPLOY_GAP_ATTEMPTS = 5;
export const DEPLOY_GAP_FIRST_DELAY_MS = 1_000;
export const DEPLOY_GAP_PATIENCE_MS = 20_000;

/** The pager said the venue or ticket does not exist: retrying will not help. */
export function isGone(error: unknown): boolean {
  const code = errorCode(error);
  return (
    errorStatus(error) === 404 &&
    (code === PAGER_ERROR.venueNotFound || code === PAGER_ERROR.ticketNotFound)
  );
}

/**
 * A failure that a deploy causes for a few seconds and that goes away on its
 * own: no answer at all, a 5xx, or a 404 the pager service did not give a code
 * (the route is not served yet).
 */
export function isTransient(error: unknown): boolean {
  const status = errorStatus(error);
  if (status === undefined || status >= 500) return true;
  return status === 404 && !isGone(error);
}
