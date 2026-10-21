/**
 * The verification screen for an account that has not confirmed its email.
 * The destination the user arrived with rides along, so verifying finishes
 * where they were headed.
 */
export function verifyEmailPath({
  email,
  sent = false,
  redirect,
  returnTo,
}: {
  email?: string;
  sent?: boolean;
  redirect?: string | null;
  returnTo?: string | null;
}): string {
  const params = new URLSearchParams();
  if (email) params.set('email', email);
  if (sent) params.set('sent', '1');
  if (redirect) params.set('redirect', redirect);
  if (returnTo) params.set('returnTo', returnTo);
  return `/verify?${params.toString()}`;
}
