import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.realtime.v1;

/** RealtimeService defines the authenticated preflight for Flow WebSocket connections. */
@route("/flow")
interface RealtimeService {
  /** Exchanges the authenticated API session for a short-lived WebSocket ticket. */
  @post
  @route("/ws-ticket")
  issueWebSocketTicket(
    ...IssueWebSocketTicketRequest,
  ): IssueWebSocketTicketResponse;
}
