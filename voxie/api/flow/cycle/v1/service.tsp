import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.cycle.v1;

@route("/flow/cycle/v1")
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
