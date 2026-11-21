import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.realtime.v1;

model IssueWebSocketTicketRequest {}

model IssueWebSocketTicketResponse {
  /** Short-lived, single-use ticket accepted by the Flow WebSocket handshake. */
  ticket: string;
}
