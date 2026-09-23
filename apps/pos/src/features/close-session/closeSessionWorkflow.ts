import { createOneKitClient } from "@voxire/api";
import { SessionServiceClient } from "@voxire/voxie/pos/session/v1/client";
import type { CloseSessionRequest } from "@voxire/voxie/pos/session/v1/types";
import { isRetryableTransportError } from "../../shared/utils/errorClassify";
import { useUserStore } from "../../shared/stores/userStore";

const sessionClient = createOneKitClient(SessionServiceClient, "pos");

export interface SessionClosePayload {
  closingFloat: number;
  expected: number;
  variance: number;
  /** Session the close belongs to, captured at enqueue time so replays hit
   *  the right register session instead of whatever happens to be open. */
  sessionId?: number;
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

type CloseSessionResult = { status: "online" } | { status: "queued" };

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

function positiveInteger(value: unknown): number | undefined {
  const numeric =
    typeof value === "string" ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function buildCloseSessionPayload({
  expected,
  actualInput,
}: BuildCloseSessionPayloadParams): SessionClosePayload {
  const actual = Number(actualInput);
  if (!Number.isFinite(actual) || actual < 0) {
    throw new Error("Enter a valid counted cash amount.");
  }

  const normalizedExpected = nonNegativeNumber(expected);
  return {
    closingFloat: actual,
    expected: normalizedExpected,
    variance: actual - normalizedExpected,
  };
}

export class NoOpenRegisterSessionError extends Error {
  constructor() {
    super("No open register session found.");
    this.name = "NoOpenRegisterSessionError";
  }
}

// The backend closes a session by id (/pos/sessions/{id}/close) and has no
// "current" route, so the open session is resolved from the business-wide list.
export async function resolveOpenSessionId(): Promise<number> {
  const res = await sessionClient.listSessions({});
  const staffId = positiveInteger(useUserStore.getState().userId);
  const open = (res.sessions ?? []).find(
    (session) =>
      staffId != null &&
      session.status === "open" &&
      positiveInteger(session.staffId) === staffId,
  );
  const sessionId = positiveInteger(open?.id);
  if (staffId == null || sessionId == null) {
    throw new NoOpenRegisterSessionError();
  }
  return sessionId;
}

export async function closeSessionWithFallback({
  payload,
  closeSession,
  enqueue,
  resolveSessionId = resolveOpenSessionId,
}: CloseSessionWithFallbackParams): Promise<CloseSessionResult> {
  // Resolve the session first: "no open session" is a permanent condition
  // that must surface now instead of becoming an unfixable queue entry.
  const sessionId = await resolveSessionId();
  try {
    await closeSession({
      id: String(sessionId),
      closingFloat: payload.closingFloat,
    });
    return { status: "online" };
  } catch (error) {
    // Only transport-class failures may be replayed later; server rejections
    // (4xx) would fail forever as poison entries.
    if (!isRetryableTransportError(error)) throw error;
    enqueue({ ...payload, sessionId });
    return { status: "queued" };
  }
}
