import { createOneKitClient } from '@voxire/api';
import { SessionServiceClient } from '@voxire/voxie/pos/session/v1/client';
import type { CloseSessionRequest } from '@voxire/voxie/pos/session/v1/types';

const sessionClient = createOneKitClient(SessionServiceClient, 'pos');

export interface SessionClosePayload {
  closingFloat: number;
  expected: number;
  variance: number;
}

interface BuildCloseSessionPayloadParams {
  expected: number;
  actualInput: string;
}

interface CloseSessionWithFallbackParams {
  payload: SessionClosePayload;
  closeSession: (payload: CloseSessionRequest) => Promise<void>;
  enqueue: (payload: SessionClosePayload) => unknown;
  resolveSessionId?: () => Promise<number>;
}

type CloseSessionResult = { status: 'online' } | { status: 'queued' };

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function buildCloseSessionPayload({
  expected,
  actualInput,
}: BuildCloseSessionPayloadParams): SessionClosePayload {
  const actual = Number(actualInput);
  if (!Number.isFinite(actual) || actual < 0) {
    throw new Error('Enter a valid counted cash amount.');
  }

  const normalizedExpected = nonNegativeNumber(expected);
  return {
    closingFloat: actual,
    expected: normalizedExpected,
    variance: actual - normalizedExpected,
  };
}

// The backend closes a session by id (/pos/sessions/{id}/close) and has no
// "current" route, so the open session is resolved from the session list.
export async function resolveOpenSessionId(): Promise<number> {
  const res = await sessionClient.listSessions({});
  const open = (res.sessions ?? []).find((s) => s.status === 'open');
  if (!open || !open.id) {
    throw new Error('No open register session found.');
  }
  const sessionId = positiveInteger(open.id);
  if (sessionId == null) {
    throw new Error('No open register session found.');
  }
  return sessionId;
}

export async function closeSessionWithFallback({
  payload,
  closeSession,
  enqueue,
  resolveSessionId = resolveOpenSessionId,
}: CloseSessionWithFallbackParams): Promise<CloseSessionResult> {
  try {
    const sessionId = await resolveSessionId();
    await closeSession({ id: String(sessionId), closingFloat: payload.closingFloat });
    return { status: 'online' };
  } catch {
    enqueue(payload);
    return { status: 'queued' };
  }
}
