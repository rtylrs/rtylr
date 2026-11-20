import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.milestone.v1;

@route("/flow/milestone/v1")
interface MilestoneService {
  @get
  @route("/milestones")
  listMilestones(...ListMilestonesRequest): ListMilestonesResponse;

  @post
  @route("/milestones")
  createMilestone(...CreateMilestoneRequest): CreateMilestoneResponse;

  @put
  @route("/milestones/{id}")
  updateMilestone(...UpdateMilestoneRequest): UpdateMilestoneResponse;
}
