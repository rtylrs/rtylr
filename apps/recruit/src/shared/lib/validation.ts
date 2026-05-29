// Form validation helpers for the public apply form in the recruit app.
// These are pure functions that return error messages or null so they
// can be used with any form library or with plain controlled components.

/**
 * validateRequired returns an error message when value is blank,
 * or null when it passes.
 */
export function validateRequired(value: string, label = 'This field'): string | null {
  if (!value || !value.trim()) return `${label} is required.`;
  return null;
}

/**
 * validateEmail returns an error message when value is not a plausible
 * email address, or null when it passes. Uses a simple heuristic that
 * catches the most common mistakes without being over-prescriptive.
 */
export function validateEmail(value: string): string | null {
  if (!value || !value.trim()) return 'Email is required.';
  const at = value.indexOf('@');
  if (at <= 0) return 'Enter a valid email address.';
  const dot = value.lastIndexOf('.');
  if (dot <= at + 1 || dot >= value.length - 1) return 'Enter a valid email address.';
  return null;
}

/**
 * validatePhone returns an error message when value is non-empty but
 * does not contain enough digits to be a phone number, or null otherwise.
 * Phone is optional in the apply form; pass empty string to skip the check.
 */
export function validatePhone(value: string): string | null {
  if (!value || !value.trim()) return null;
  const digits = value.replace(/\D/g, '');
  if (digits.length < 7) return 'Enter a valid phone number.';
  return null;
}

/**
 * validateUrl returns an error message when value is non-empty but is
 * not a valid HTTP/HTTPS URL, or null otherwise.
 */
export function validateUrl(value: string): string | null {
  if (!value || !value.trim()) return null;
  try {
    const url = new URL(value);
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return 'Enter a valid URL starting with http:// or https://.';
    }
    return null;
  } catch {
    return 'Enter a valid URL starting with http:// or https://.';
  }
}

/**
 * validateMaxLength returns an error message when value exceeds max
 * characters, or null otherwise.
 */
export function validateMaxLength(value: string, max: number, label = 'This field'): string | null {
  if (value.length > max) return `${label} must be ${max} characters or fewer.`;
  return null;
}

/**
 * validateFileSize returns an error message when file size exceeds
 * maxBytes, or null otherwise.
 */
export function validateFileSize(file: File, maxBytes: number): string | null {
  if (file.size > maxBytes) {
    const maxMb = (maxBytes / (1024 * 1024)).toFixed(0);
    return `File must be smaller than ${maxMb} MB.`;
  }
  return null;
}

/**
 * validateFileType returns an error message when the file's MIME type
 * is not in the accepted list, or null otherwise.
 */
export function validateFileType(file: File, accepted: string[]): string | null {
  if (!accepted.includes(file.type)) {
    return `File type not accepted. Accepted types: ${accepted.join(', ')}.`;
  }
  return null;
}
