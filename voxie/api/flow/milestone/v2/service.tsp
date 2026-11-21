import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.milestone.v2;

@route("/flow/milestone/v2")
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

  /**
   * deleteMilestone removes the milestone. Its tasks stay in the project and
   * simply stop belonging to a milestone.
   */
  @delete
  @route("/milestones/{id}")
  deleteMilestone(...DeleteMilestoneRequest): DeleteMilestoneResponse;
}
