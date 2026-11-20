import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.caret.v1;

model CaretTool {
  name: string;
  label: string;
  description: string;
  category: string;
  requires_approval: boolean;

  /**
   * Access scope of this tool: "read" or "write". Mirrors the server-side
   * gating already applied in ask/analyst modes; exposed so the UI can
   * surface it too (e.g. counting write-scoped tools).
   */
  scope: string;
}

model ListCaretToolsRequest {
  mode: string;
}

model ListCaretToolsResponse {
  tools: CaretTool[];
}

/**
 * Image inputs accepted by the authenticated Caret SSE endpoint. At most four
 * raster images, two MiB combined; SVG and remote URLs are not accepted.
 */
model CaretImageInput {
  name: string;
  data_url: string;
}

/** A tool contract as the runtime registry describes it. */
model CaretToolDef {
  name: string;
  description: string;
  module: string;
  scope: string;
  requires_approval: boolean;
  input_hint: string;
  output_hint: string;
  freshness: string;
  side_effect: string;
  approval_reason: string;
}

model ListToolsRequest {
  @query("mode")
  mode: string;
}

model ListToolsResponse {
  tools: CaretToolDef[];
}

model GetProviderStatusRequest {}

model GetProviderStatusResponse {
  available: boolean;
  primary_provider: string;

  /** Provider name to availability. */
  providers: unknown;
}

/** Approve or deny a pending tool call. input is the tool's JSON input. */
model ToolDecisionRequest {
  tool_call_id: string;
  name: string;
  input: unknown;
  decision: string;
  mode: string;

  /**
   * The chat the proposal came from. The outcome is recorded there so
   * Caret knows in later turns what was done or skipped.
   */
  session_id: string;
}

model ToolDecisionResponse {
  tool_call_id: string;
  status: string;
  message: string;
  executed: boolean;
  tool: CaretToolDef;
  preview: string;
  output: unknown;
  duration_ms: int64;
}

model StopReplyRequest {
  session_id: string;
}

model StopReplyResponse {
  /** False when no reply was generating in the chat. */
  stopped: boolean;
}

model RateReplyRequest {
  session_id: string;

  /** The stored message id of the reply; empty rates the chat's latest reply. */
  message_id: string;

  /** "up", "down", or "clear" to remove a verdict. */
  rating: string;

  /** What was wrong or right, in the person's words. Optional. */
  reason: string;
}

model RateReplyResponse {
  saved: boolean;
}

model BuddySettings {
  enabled: boolean;
  instructions: string;
  interval_minutes: int32;
  daily_run_cap: int32;
  run_as_user_id: int64;
}

model GetBuddySettingsRequest {}

model GetBuddySettingsResponse {
  settings: BuddySettings;
}

model PutBuddySettingsRequest {
  enabled: boolean;
  instructions: string;
  interval_minutes: int32;
  daily_run_cap: int32;
}

model PutBuddySettingsResponse {
  settings: BuddySettings;
}

model BuddyRun {
  id: int64;
  trigger_kind: string;
  status: string;
  report: string;
  error: string;
  tool_calls: int32;
  tokens: int64;
  created_at: string;
  started_at: string;
  finished_at: string;
}

model ListBuddyRunsRequest {}

model ListBuddyRunsResponse {
  runs: BuddyRun[];
}

model BuddyAction {
  id: int64;
  run_id: int64;
  tool_name: string;
  input: unknown;
  status: string;
  decided_by: int64;
  result: unknown;
  expires_at: string;
  created_at: string;
  decided_at: string;
}

model ListBuddyActionsRequest {
  @query("status")
  status: string;
}

model ListBuddyActionsResponse {
  actions: BuddyAction[];
}

model DecideBuddyActionRequest {
  action_id: int64;
  decision: string;
}

model DecideBuddyActionResponse {
  action: BuddyAction;
}

model GetBuddyThreadRequest {}

model GetBuddyThreadResponse {
  session_id: string;
  last_post_at: string;
}

model BuddyRoutine {
  id: int64;
  name: string;
  instructions: string;
  every_minutes: int32;
  enabled: boolean;
  next_run_at: string;
  last_run_at: string;
}

model ListBuddyRoutinesRequest {}

model ListBuddyRoutinesResponse {
  routines: BuddyRoutine[];
}

model SetBuddyRoutineEnabledRequest {
  id: int64;
  enabled: boolean;
}

model SetBuddyRoutineEnabledResponse {
  routine: BuddyRoutine;
}

model DeleteBuddyRoutineRequest {
  id: int64;
}

model DeleteBuddyRoutineResponse {}
