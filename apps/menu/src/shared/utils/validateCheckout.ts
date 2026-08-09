/**
 * Validation helpers for the checkout form.
 */

export interface CheckoutFields {
  name: string;
  phone: string;
  email?: string;
}

export interface ValidationResult {
  valid: boolean;
  error: string | null;
}

/**
 * Validates that the customer name is non-empty after trimming.
 */
export function validateName(name: string): ValidationResult {
  if (!name.trim()) {
    return { valid: false, error: 'Please enter your name.' };
  }
  return { valid: true, error: null };
}

/**
 * Validates that the phone field is non-empty after trimming.
 */
export function validatePhone(phone: string): ValidationResult {
  if (!phone.trim()) {
    return { valid: false, error: 'Please enter your phone number.' };
  }
  return { valid: true, error: null };
}

/**
 * Validates an optional email field. Empty is accepted; if provided it must
 * contain an "@" symbol.
 */
export function validateEmail(email: string): ValidationResult {
  const trimmed = email.trim();
  if (!trimmed) return { valid: true, error: null };
  if (!trimmed.includes('@')) {
    return { valid: false, error: 'Please enter a valid email address.' };
  }
  return { valid: true, error: null };
}

/**
 * Runs all checkout validations and returns the first error found, or null
 * if all fields are valid.
 */
export function validateCheckoutFields(fields: CheckoutFields): string | null {
  const nameResult = validateName(fields.name);
  if (!nameResult.valid) return nameResult.error;

  const phoneResult = validatePhone(fields.phone);
  if (!phoneResult.valid) return phoneResult.error;

  if (fields.email !== undefined) {
    const emailResult = validateEmail(fields.email);
    if (!emailResult.valid) return emailResult.error;
  }

  return null;
}
