import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.cycle.v2;

/**
 * CycleService plans a project's cycles (sprints): what is in each one, in what
 * order, and when each starts and ends.
 */
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

  /**
   * deleteCycle removes the cycle. Its tasks stay in the project and go back to
   * the unplanned backlog.
   */
  @delete
  @route("/cycles/{id}")
  deleteCycle(...DeleteCycleRequest): DeleteCycleResponse;

  @get
  @route("/cycles/{id}")
  getCycle(...GetCycleRequest): GetCycleResponse;

  /** startCycle starts an upcoming cycle. A project has one active cycle at a time. */
  @post
  @route("/cycles/{id}/start")
  startCycle(...StartCycleRequest): StartCycleResponse;

  /** completeCycle completes a cycle and settles every unfinished task in one write. */
  @post
  @route("/cycles/{id}/complete")
  completeCycle(...CompleteCycleRequest): CompleteCycleResponse;

  /**
   * moveCycleTask places a task in a cycle or the unplanned backlog at a
   * position. It is the one write behind drag and drop on the Backlog screen.
   */
  @put
  @route("/cycles/tasks/order")
  moveCycleTask(...MoveCycleTaskRequest): MoveCycleTaskResponse;

  /** getCycleCapacity returns each assignee's available and committed time. */
  @get
  @route("/cycles/{id}/capacity")
  getCycleCapacity(...GetCycleCapacityRequest): GetCycleCapacityResponse;
}
