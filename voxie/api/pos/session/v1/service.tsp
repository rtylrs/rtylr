import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.session.v1;

/** SessionService defines POS register session endpoints. */
@route("/pos/session/v1")
interface SessionService {
  /** Returns register sessions for the active business. */
  @get
  @route("/sessions")
  listSessions(...ListSessionsRequest): ListSessionsResponse;

  /** Opens a new register session. */
  @post
  @route("/sessions/open")
  openSession(...OpenSessionRequest): OpenSessionResponse;

  /** Closes an open register session. */
  @post
  @route("/sessions/{id}/close")
  closeSession(...CloseSessionRequest): CloseSessionResponse;
}
