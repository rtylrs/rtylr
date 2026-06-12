export interface LocationOption {
  locationId: number;
  name: string;
  address: string;
  phone: string;
}

export interface TerminalOption {
  id: number;
  name: string;
  type: string;
  isActive: boolean;
}

export const CREATED_TERMINAL_ID_ERROR =
  'Terminal registration did not return a valid terminal ID.';

export const CREATED_LOCATION_ID_ERROR =
  'Location creation did not return a valid location ID.';

function positiveInt(value: unknown): number | null {
  const number = Number(value);
  if (!Number.isFinite(number) || number <= 0) return null;
  return Math.floor(number);
}

function record(value: unknown): Record<string, unknown> | null {
  return value && typeof value === 'object' ? (value as Record<string, unknown>) : null;
}

export function extractLocationsList(data: unknown): unknown[] {
  if (Array.isArray(data)) return data;
  const raw = record(data);
  if (Array.isArray(raw?.locations)) return raw.locations as unknown[];
  return [];
}

export function mergeLocationOptions(...lists: LocationOption[][]): LocationOption[] {
  const seen = new Map<number, LocationOption>();
  for (const list of lists) {
    for (const loc of list) {
      if (!seen.has(loc.locationId)) seen.set(loc.locationId, loc);
    }
  }
  return [...seen.values()];
}

export function mapBranchLocations(rawBranches: unknown[]): LocationOption[] {
  return rawBranches.flatMap((branch) => {
    const raw = branch as Record<string, unknown>;
    const id = positiveInt(raw.id ?? raw.locationId ?? raw.location_id);
    if (!id) return [];

    return [
      {
        locationId: id,
        name: String(raw.name ?? ''),
        address: String(raw.address ?? ''),
        phone: String(raw.phone ?? ''),
      },
    ];
  });
}

export function filterTerminalsForLocation(
  rawTerminals: unknown[],
  locationId: number
): TerminalOption[] {
  return rawTerminals.flatMap((terminal) => {
    const raw = terminal as Record<string, unknown>;
    const id = positiveInt(raw.id);
    const branchId = positiveInt(raw.branchId ?? raw.branch_id);
    if (!id || branchId !== locationId) return [];

    return [
      {
        id,
        name: String(raw.name ?? ''),
        type: String(raw.type ?? 'pos'),
        isActive: raw.isActive !== false,
      },
    ];
  });
}

export function extractCreatedLocationId(response: unknown): number {
  const raw = record(response);
  const nestedData = record(raw?.data);
  const id =
    positiveInt(raw?.location_id ?? raw?.locationId ?? raw?.id) ??
    positiveInt(nestedData?.location_id ?? nestedData?.locationId ?? nestedData?.id);

  if (!id) {
    throw new Error(CREATED_LOCATION_ID_ERROR);
  }

  return id;
}

export function extractCreatedTerminalId(response: unknown): number {
  const raw = record(response);
  const nestedTerminal = record(raw?.terminal);
  const nestedData = record(raw?.data);
  const id =
    positiveInt(raw?.id) ??
    positiveInt(raw?.terminalId ?? raw?.terminal_id) ??
    positiveInt(nestedTerminal?.id) ??
    positiveInt(nestedData?.id ?? nestedData?.terminalId ?? nestedData?.terminal_id);

  if (!id) {
    throw new Error(CREATED_TERMINAL_ID_ERROR);
  }

  return id;
}
