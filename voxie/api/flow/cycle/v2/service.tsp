import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.cycle.v2;

@route("/flow/cycle/v2")
interface CycleService {
  @get
  @route("/cycles")
  listCycles(...ListCyclesRequest): ListCyclesResponse;

  @post
  @route("/cycles")
  createCycle(...CreateCycleRequest): CreateCycleResponse;

  @put
  @route("/cycles/{id}")
  updateCycle(...UpdateCycleRequest): UpdateCycleResponse;
}
