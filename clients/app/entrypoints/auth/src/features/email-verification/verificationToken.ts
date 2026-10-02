/**
 * The email verification token's shape, mirrored from onion so the form can
 * check what it is about to send without a round trip.
 *
 * onion issues `"<userID>.<10 digits>"` and
 * `parseEmailVerificationToken`
 * (services/onion/internal/modules/user/internal/infrastructure/email_verification.go)
 * accepts nothing else: the id must parse as a positive int64 and the code half
 * must be exactly ten digits. The id half is the point of the format — it is
 * what the anonymous endpoint needs to recompute the account-bound digest — so
 * it is carried through here and never trimmed off as noise.
 *
 * The six-digit code that predates the user binding is a third case, and it is
 * forwarded rather than rewritten: onion refuses it with its own "request a new
 * one" message, which is the only copy of that in any language.
 */

/** Nineteen digits of int64, the separator, and the ten-digit code. */
export const VERIFICATION_TOKEN_MAX_LENGTH = 30;

/** The ten digits onion generates. */
const CODE_DIGITS = 10;

/**
 * Whitespace, plus the zero-width and bidirectional marks mail clients insert
 * around a copied token. `\s` already covers the non-breaking space.
 */
const INVISIBLE = /[\s\u200b\u200c\u200d\u2060\u200e\u200f]/g;

const INT64_MAX = BigInt('9223372036854775807');

/**
 * Reduce a typed or pasted value to the token it can contain, so what is on
 * screen is what gets submitted.
 *
 * A pasted verification *link* is reduced to its `token` parameter, because the
 * link is the whole of what the email offers: the bare code beside it carries no
 * user id, so no client can turn it into something onion accepts.
 */
export function normalizeVerificationToken(raw: string): string {
  const cleaned = raw.replace(INVISIBLE, '');
  const param = cleaned.indexOf('token=');
  const scoped =
    param === -1 ? cleaned : (cleaned.slice(param + 'token='.length).split(/[&#]/)[0] ?? '');
  const separator = scoped.indexOf('.');
  if (separator === -1) return scoped.replace(/\D/g, '');

  // A second separator ends the code: digits either side of it must not be
  // joined into a code that was never issued.
  const code = scoped.slice(separator + 1).split('.')[0] ?? '';
  return `${scoped.slice(0, separator).replace(/\D/g, '')}.${code.replace(/\D/g, '')}`;
}

/** True for exactly the tokens onion will accept. */
export function isVerificationToken(value: string): boolean {
  const separator = value.indexOf('.');
  if (separator <= 0) return false;

  const userId = value.slice(0, separator);
  if (!/^\d+$/.test(userId)) return false;
  if (!new RegExp(`^\\d{${CODE_DIGITS}}$`).test(value.slice(separator + 1))) return false;

  const parsed = BigInt(userId);
  return parsed > BigInt(0) && parsed <= INT64_MAX;
}

/**
 * True for the pre-binding six-digit code.
 *
 * onion answers it with a distinct "this code is too old, request a new one"
 * error before it spends any of the attempt budget, so it is worth sending
 * rather than answering here with copy the client cannot localize.
 */
export function isLegacyVerificationCode(value: string): boolean {
  return /^\d{6}$/.test(value);
}
