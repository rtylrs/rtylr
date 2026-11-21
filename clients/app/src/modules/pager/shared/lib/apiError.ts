/** The stable codes the pager service sends next to an error message. */
export const PAGER_ERROR = {
  venueNotFound: 'venue_not_found',
  ticketNotFound: 'ticket_not_found',
  invalidLabel: 'invalid_label',
  labelTaken: 'label_taken',
  invalidTransition: 'invalid_transition',
  tooManyOpenTickets: 'too_many_open_tickets',
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
