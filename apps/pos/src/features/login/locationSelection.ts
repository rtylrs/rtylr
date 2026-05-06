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

function positiveInt(value: unknown): number | null {
  const number = Number(value);
  if (!Number.isFinite(number) || number <= 0) return null;
  return Math.floor(number);
}

function record(value: unknown): Record<string, unknown> | null {
  return value && typeof value === 'object' ? (value as Record<string, unknown>) : null;
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
