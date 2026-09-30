// Pure helpers for the POS sign-in roster. Kept separate from the component so
// the normalization (which tolerates both snake_case and camelCase, since the
// API layer camelCases responses) is unit-testable. The PIN length and format
// policy lives in ./pinPolicy, which the staff screen shares.

export interface RosterMember {
  userId: number;
  name: string;
  hasPin: boolean;
  isManager: boolean;
  posRole: string;
}

function positiveInt(value: unknown): number {
  const n = typeof value === 'string' ? Number(value.trim()) : Number(value);
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0;
}

function asBool(value: unknown): boolean {
  return value === true;
}

/** Extracts the roster array from the endpoint envelope ({ roster: [...] }). */
export function extractRoster(data: unknown): unknown[] {
  if (Array.isArray(data)) return data;
  const raw = data && typeof data === 'object' ? (data as Record<string, unknown>) : null;
  if (raw && Array.isArray(raw.roster)) return raw.roster as unknown[];
  return [];
}

/** Maps a raw roster response into typed members, dropping unusable rows. */
export function mapRoster(data: unknown): RosterMember[] {
  return extractRoster(data).flatMap((entry) => {
    const r = entry as Record<string, unknown>;
    const userId = positiveInt(r.userId ?? r.user_id);
    const name = typeof r.name === 'string' ? r.name.trim() : '';
    if (!userId || !name) return [];
    return [
      {
        userId,
        name,
        hasPin: asBool(r.hasPin ?? r.has_pin),
        isManager: asBool(r.isManager ?? r.is_manager ?? r.canManagePosUsers ?? r.can_manage_pos_users),
        posRole: typeof r.posRole === 'string' ? r.posRole : String(r.pos_role ?? ''),
      },
    ];
  });
}

