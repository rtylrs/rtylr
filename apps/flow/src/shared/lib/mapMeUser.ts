import type { FlowUser } from '../stores/userStore';

/**
 * Map `/user/me` (or equivalent) payload to Flow profile fields, same field fallbacks as Hub sidebar.
 */
export function mapMeResponseToFlowUser(
  data: Record<string, unknown> | null | undefined
): FlowUser | null {
  if (!data || typeof data !== 'object') return null;

  const email = String(data.email ?? data.emailAddress ?? '').trim();
  const id = String(data.id ?? data.userId ?? '').trim();

  const fullName =
    typeof data.fullName === 'string'
      ? data.fullName.trim()
      : typeof data.full_name === 'string'
        ? data.full_name.trim()
        : '';
  const firstName = typeof data.firstName === 'string' ? data.firstName.trim() : '';
  const lastName = typeof data.lastName === 'string' ? data.lastName.trim() : '';
  const singleName = typeof data.name === 'string' ? data.name.trim() : '';

  const nameFromParts = [firstName, lastName].filter(Boolean).join(' ').trim();

  const name = fullName || nameFromParts || singleName || firstName || email || 'User';

  return {
    id: id || email || 'user',
    name,
    email,
    avatar: typeof data.avatar === 'string' ? data.avatar : undefined,
    firstName: firstName || undefined,
  };
}
