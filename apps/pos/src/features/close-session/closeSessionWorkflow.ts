import type { SessionPayload } from '../../shared/services/posService';

interface BuildCloseSessionPayloadParams {
  sessionId: string;
  expected: number;
  actualInput: string;
}

interface CloseSessionWithFallbackParams {
  payload: SessionPayload;
  closeSession: (payload: SessionPayload) => Promise<void>;
  enqueue: (payload: SessionPayload) => unknown;
}

type CloseSessionResult = { status: 'online' } | { status: 'queued' };

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

export function buildCloseSessionPayload({
  sessionId,
  expected,
  actualInput,
}: BuildCloseSessionPayloadParams): SessionPayload {
  const actual = Number(actualInput);
  if (!Number.isFinite(actual) || actual < 0) {
    throw new Error('Enter a valid counted cash amount.');
  }

  const normalizedExpected = nonNegativeNumber(expected);
  return {
    sessionId,
    expected: normalizedExpected,
    actual,
    variance: actual - normalizedExpected,
  };
}

export async function closeSessionWithFallback({
  payload,
  closeSession,
  enqueue,
}: CloseSessionWithFallbackParams): Promise<CloseSessionResult> {
  try {
    await closeSession(payload);
    return { status: 'online' };
  } catch {
    enqueue(payload);
    return { status: 'queued' };
  }
}
