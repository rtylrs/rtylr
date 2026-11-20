import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.terminal.v1;

/** TerminalService defines POS terminal device endpoints. */
@route("/pos/terminal/v1")
interface TerminalService {
  /** Returns terminals for the active business. */
  @get
  @route("/terminals")
  listTerminals(...ListTerminalsRequest): ListTerminalsResponse;

  /** Registers a new terminal. */
  @post
  @route("/terminals")
  createTerminal(...CreateTerminalRequest): CreateTerminalResponse;

  /** Updates an existing terminal. */
  @put
  @route("/terminals/{id}")
  updateTerminal(...UpdateTerminalRequest): UpdateTerminalResponse;

  /** Records a terminal heartbeat. */
  @post
  @route("/terminals/heartbeat")
  heartbeatTerminal(...HeartbeatTerminalRequest): HeartbeatTerminalResponse;
}
