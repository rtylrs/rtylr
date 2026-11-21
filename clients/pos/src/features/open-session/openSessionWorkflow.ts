import type { OpenSessionRequest } from '@voxire/voxie/pos/session/v1/types';
import { positiveInteger } from '../close-session/closeSessionWorkflow';

export class OpenRegisterInputError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'OpenRegisterInputError';
  }
}

interface OpenRegisterSessionParams {
  floatInput: string;
  locationId: number | null;
  staffId: string | null;
  openSession: (payload: OpenSessionRequest) => Promise<unknown>;
  resolveSessionId: () => Promise<number>;
  openDrawer: (openingFloat: number) => void;
}

export type OpenRegisterSessionResult =
  | { status: 'opened'; openingFloat: number }
  | { status: 'already_open' };

// A blank field means an empty drawer, the common case for a card-only till.
export function parseOpeningFloat(input: string): number {
  const trimmed = input.trim();
  if (!trimmed) return 0;
  const amount = Number(trimmed);
  if (!Number.isFinite(amount) || amount < 0) {
    throw new OpenRegisterInputError('Enter a valid opening float.');
  }
  return amount;
}

export async function openRegisterSession({
  floatInput,
  locationId,
  staffId,
  openSession,
  resolveSessionId,
  openDrawer,
}: OpenRegisterSessionParams): Promise<OpenRegisterSessionResult> {
  const openingFloat = parseOpeningFloat(floatInput);
  const location = positiveInteger(locationId);
  const staff = positiveInteger(staffId);
  if (location == null || staff == null) {
    throw new OpenRegisterInputError('Select a location and sign in before opening the register.');
  }

  try {
    await openSession({
      locationId: String(location),
      staffId: String(staff),
      openingFloat,
    });
  } catch (error) {
    // The server allows one open session per staff member. If one is already
    // there the register is usable, and the local drawer keeps its movements.
    try {
      await resolveSessionId();
    } catch {
      throw error;
    }
    return { status: 'already_open' };
  }

  openDrawer(openingFloat);
  return { status: 'opened', openingFloat };
}

export function openRegisterErrorMessage(error: unknown): string {
  if (error instanceof OpenRegisterInputError) return error.message;
  const status =
    error && typeof error === 'object' && 'statusCode' in error
      ? (error as { statusCode?: unknown }).statusCode
      : undefined;
  if (status === 403) {
    return 'You do not have permission to open a register. Ask a manager.';
  }
  return 'Could not open the register. Check your connection and try again.';
}
