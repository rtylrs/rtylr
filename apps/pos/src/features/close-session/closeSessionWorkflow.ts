import { unwrap } from '@voxire/api';
import { posService } from '../../shared/services/api/posService';
import type { CloseSessionPayload } from '../../shared/services/api/posService';

export interface SessionClosePayload extends CloseSessionPayload {
  expected: number;
  variance: number;
}

interface BuildCloseSessionPayloadParams {
  expected: number;
  actualInput: string;
}

interface CloseSessionWithFallbackParams {
  payload: SessionClosePayload;
  closeSession: (sessionId: number, payload: CloseSessionPayload) => Promise<void>;
  enqueue: (payload: SessionClosePayload) => unknown;
}

type CloseSessionResult = { status: 'online' } | { status: 'queued' };

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
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
    closing_float: actual,
    expected: normalizedExpected,
    variance: actual - normalizedExpected,
  };
}

// The backend closes a session by id (/pos/sessions/{id}/close) and has no
// "current" route, so the open session is resolved from the session list.
export async function resolveOpenSessionId(): Promise<number> {
  const res = (await unwrap(posService.getSessions())) as {
    sessions?: Array<{ id?: number; status?: string }>;
  };
  const open = (res.sessions ?? []).find((s) => s.status === 'open');
  if (!open || !open.id) {
    throw new Error('No open register session found.');
  }
  return open.id;
}

export async function closeSessionWithFallback({
  payload,
  closeSession,
  enqueue,
}: CloseSessionWithFallbackParams): Promise<CloseSessionResult> {
  try {
    const sessionId = await resolveOpenSessionId();
    await closeSession(sessionId, { closing_float: payload.closing_float });
    return { status: 'online' };
  } catch {
    enqueue(payload);
    return { status: 'queued' };
  }
}
