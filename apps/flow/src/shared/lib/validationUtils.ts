/** Returns true if the string is a valid non-empty project/workspace name. */
export function isValidName(name: string): boolean {
  return name.trim().length >= 1 && name.trim().length <= 100;
}

/** Returns true if the string is a non-empty task title. */
export function isValidTitle(title: string): boolean {
  return title.trim().length >= 1 && title.trim().length <= 255;
}

/** Returns true if the string is a valid ISO date (YYYY-MM-DD). */
export function isValidIsoDate(date: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(date) && !isNaN(Date.parse(date));
}

/** Returns true if the start date is before the end date. */
export function isValidDateRange(start: string, end: string): boolean {
  if (!isValidIsoDate(start) || !isValidIsoDate(end)) return false;
  return start <= end;
}

/** Returns true if the estimate is a positive finite number. */
export function isValidEstimate(estimate: unknown): boolean {
  return typeof estimate === 'number' && isFinite(estimate) && estimate > 0;
}

/** Returns true if the hex color is a valid 6-digit color. */
export function isValidHexColor(color: string): boolean {
  return /^#[0-9a-fA-F]{6}$/.test(color);
}

/** Returns true if the email address is plausibly valid (basic check). */
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

/** Returns true if the progress value is between 0 and 100 inclusive. */
export function isValidProgress(progress: number): boolean {
  return Number.isInteger(progress) && progress >= 0 && progress <= 100;
}

/** Returns a validation error message for a name, or null if valid. */
export function nameValidationError(name: string): string | null {
  if (name.trim().length === 0) return 'Name is required';
  if (name.trim().length > 100) return 'Name must be 100 characters or fewer';
  return null;
}

/** Returns a validation error message for a title, or null if valid. */
export function titleValidationError(title: string): string | null {
  if (title.trim().length === 0) return 'Title is required';
  if (title.trim().length > 255) return 'Title must be 255 characters or fewer';
  return null;
}
