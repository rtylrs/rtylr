import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.caret.v1;

@route("/hub/caret/v1")
interface CaretService {
  @post
  @route("/tools")
  listCaretTools(...ListCaretToolsRequest): ListCaretToolsResponse;
}

/**
 * Runtime routes the Caret panel calls directly. Paths match the routes the
 * legacy handlers served; the SSE stream stays a raw handler.
 */
@route("/")
interface CaretRuntimeService {
  @get
  @route("/caret/tools")
  listTools(...ListToolsRequest): ListToolsResponse;

  @get
  @route("/ai/status")
  getProviderStatus(...GetProviderStatusRequest): GetProviderStatusResponse;

  @post
  @route("/caret/tools/decision")
  decideTool(...ToolDecisionRequest): ToolDecisionResponse;

  /**
   * Stops the reply generating in a chat. A reply whose client disconnects
   * otherwise keeps generating and is saved, so the Stop button calls this.
   */
  @post
  @route("/ai/generate-text-stream/stop")
  stopReply(...StopReplyRequest): StopReplyResponse;

  /**
   * Records a thumbs up or down on a reply. Caret learns lessons from these
   * verdicts and follows them in later replies.
   */
  @post
  @route("/caret/feedback")
  rateReply(...RateReplyRequest): RateReplyResponse;
}

@route("/")
interface BuddyService {
  @get
  @route("/caret/buddy/settings")
  getBuddySettings(...GetBuddySettingsRequest): GetBuddySettingsResponse;

  @put
  @route("/caret/buddy/settings")
  putBuddySettings(...PutBuddySettingsRequest): PutBuddySettingsResponse;

  @get
  @route("/caret/buddy/runs")
  listBuddyRuns(...ListBuddyRunsRequest): ListBuddyRunsResponse;

  @get
  @route("/caret/buddy/actions")
  listBuddyActions(...ListBuddyActionsRequest): ListBuddyActionsResponse;

  @post
  @route("/caret/buddy/actions/decision")
  decideBuddyAction(...DecideBuddyActionRequest): DecideBuddyActionResponse;

  @get
  @route("/caret/buddy/thread")
  getBuddyThread(...GetBuddyThreadRequest): GetBuddyThreadResponse;

  @get
  @route("/caret/buddy/routines")
  listBuddyRoutines(...ListBuddyRoutinesRequest): ListBuddyRoutinesResponse;

  @post
  @route("/caret/buddy/routines/enabled")
  setBuddyRoutineEnabled(
    ...SetBuddyRoutineEnabledRequest,
  ): SetBuddyRoutineEnabledResponse;

  @post
  @route("/caret/buddy/routines/delete")
  deleteBuddyRoutine(...DeleteBuddyRoutineRequest): DeleteBuddyRoutineResponse;
}
