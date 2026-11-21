import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.floor.v1;

/** FloorService defines POS floor plan endpoints. */
@route("/pos/floor/v1")
interface FloorService {
  /** Returns floors, optionally filtered by branch. */
  @get
  @route("/floors")
  listFloors(...ListFloorsRequest): ListFloorsResponse;

  /** Creates a new floor. */
  @post
  @route("/floors")
  createFloor(...CreateFloorRequest): CreateFloorResponse;

  /** Updates an existing floor. */
  @put
  @route("/floors/{id}")
  updateFloor(...UpdateFloorRequest): UpdateFloorResponse;

  /** Removes a floor by ID. */
  @delete
  @route("/floors/{id}")
  deleteFloor(...DeleteFloorRequest): DeleteFloorResponse;
}
