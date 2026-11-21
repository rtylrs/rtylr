import type { MyLocation } from '@voxire/voxie/pos/access/v1/types';
import type { Location } from '@voxire/voxie/pos/location/v1/types';
import type { Terminal } from '@voxire/voxie/pos/terminal/v1/types';

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

export function mergeLocationOptions(...lists: LocationOption[][]): LocationOption[] {
  const seen = new Map<number, LocationOption>();
  for (const list of lists) {
    for (const loc of list) {
      if (!seen.has(loc.locationId)) seen.set(loc.locationId, loc);
    }
  }
  return [...seen.values()];
}

function branchLocationRawId(location: Location | MyLocation): unknown {
  if ('id' in location && location.id !== undefined) return location.id;
  if ('locationId' in location && location.locationId !== undefined) return location.locationId;
  return undefined;
}

export function mapBranchLocations(locations: Array<Location | MyLocation>): LocationOption[] {
  return locations.flatMap((location) => {
    const id = positiveInt(branchLocationRawId(location));
    if (!id) return [];

    return [
      {
        locationId: id,
        name: location.name ?? '',
        address: location.address ?? '',
        phone: location.phone ?? '',
      },
    ];
  });
}

export function filterTerminalsForLocation(
  rawTerminals: Terminal[],
  locationId: number
): TerminalOption[] {
  return rawTerminals.flatMap((terminal) => {
    const id = positiveInt(terminal.id);
    const branchId = positiveInt(terminal.branchId);
    if (!id || branchId !== locationId) return [];

    return [
      {
        id,
        name: terminal.name ?? '',
        type: 'pos',
        isActive: terminal.isActive ?? false,
      },
    ];
  });
}

export function createdLocationId(locationId: string): number {
  const id = positiveInt(locationId);

  if (!id) {
    throw new Error(CREATED_LOCATION_ID_ERROR);
  }

  return id;
}

export function createdTerminalId(terminalId: string): number {
  const id = positiveInt(terminalId);

  if (!id) {
    throw new Error(CREATED_TERMINAL_ID_ERROR);
  }

  return id;
}

const TERMINAL_DEVICE_ID_KEY = 'rtylr.pos.deviceId';

export function getOrCreateTerminalDeviceId(storage?: Pick<Storage, 'getItem' | 'setItem'>): string {
  const store = storage ?? (typeof window !== 'undefined' ? window.localStorage : undefined);
  try {
    const existing = store?.getItem(TERMINAL_DEVICE_ID_KEY)?.trim();
    if (existing) return existing;
  } catch {
    // Ignore storage read failures in restricted environments.
  }

  const generated =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? `pos-${crypto.randomUUID()}`
      : `pos-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

  try {
    store?.setItem(TERMINAL_DEVICE_ID_KEY, generated);
  } catch {
    // Ignore storage write failures in restricted environments.
  }

  return generated;
}
