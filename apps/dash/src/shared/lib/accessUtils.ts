/**
 * Permission and access-control helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type AccessLevel = 'none' | 'read' | 'write' | 'admin';

const ACCESS_RANK: Record<AccessLevel, number> = {
  none: 0,
  read: 1,
  write: 2,
  admin: 3,
};

/** Returns true if the given access level meets or exceeds the required level. */
export function hasAccess(level: AccessLevel, required: AccessLevel): boolean {
  return ACCESS_RANK[level] >= ACCESS_RANK[required];
}

/** Returns the higher of two access levels. */
export function mergeAccess(a: AccessLevel, b: AccessLevel): AccessLevel {
  return ACCESS_RANK[a] >= ACCESS_RANK[b] ? a : b;
}

/** Returns a human-readable label for an access level. */
export function accessLabel(level: AccessLevel): string {
  const labels: Record<AccessLevel, string> = {
    none: 'No Access',
    read: 'Read Only',
    write: 'Read & Write',
    admin: 'Administrator',
  };
  return labels[level];
}

/** Returns true if the user is an admin. */
export function isAdmin(level: AccessLevel): boolean {
  return level === 'admin';
}

/** Returns true if the user can write (includes admin). */
export function canWrite(level: AccessLevel): boolean {
  return hasAccess(level, 'write');
}

/** Returns true if the user can read (includes write and admin). */
export function canRead(level: AccessLevel): boolean {
  return hasAccess(level, 'read');
}

/**
 * Resolves effective access for a resource given user-level and resource-level overrides.
 * Resource override takes precedence if provided.
 */
export function resolveAccess(
  userLevel: AccessLevel,
  resourceOverride?: AccessLevel | null
): AccessLevel {
  if (resourceOverride != null) return resourceOverride;
  return userLevel;
}

/** Returns all access levels sorted from lowest to highest. */
export function allAccessLevels(): AccessLevel[] {
  return ['none', 'read', 'write', 'admin'];
}
