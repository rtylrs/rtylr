/**
 * Caret runtime limits. Centralized so the composer, the warning banners,
 * and any future server validators agree on the same numbers.
 */

export const CARET_LIMITS = {
  /** Max characters a single user message may contain before we soft-error. */
  maxMessageChars: 8_000,
  /** Soft-warn the user as they approach the message ceiling. */
  warnAtChars: 7_200,
  /** Max files a single message may attach. */
  maxAttachments: 5,
  /** Per-file size cap in bytes. */
  maxAttachmentBytes: 10 * 1024 * 1024,
  /** Max messages we keep in a single chat session client-side. */
  maxClientMessages: 500,
};
