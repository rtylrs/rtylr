import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.time.v2;

/**
 * TimeService is Flow's daily time logging: a work day per person per date, the
 * lines of work inside it, and the work-type tree those lines are classified by.
 *
 * Lines persist as HR time entries, so Flow and payroll read the same rows rather
 * than reconciling two stores. Approval lives in a separate service.
 */
@route("/flow/time/v2")
interface TimeService {
  @get
  @route("/work-days")
  getWorkDay(...GetWorkDayRequest): GetWorkDayResponse;

  @put
  @route("/work-days")
  upsertWorkDay(...UpsertWorkDayRequest): UpsertWorkDayResponse;

  @get
  @route("/work-days/range")
  listWorkDays(...ListWorkDaysRequest): ListWorkDaysResponse;

  /**
   * listTeamWorkDays replaces v1's workspace-scoped team view. The roster is
   * the caller's business team; visibility is governed by the business's
   * configured policy (who may see whose days), enforced server-side.
   */
  @get
  @route("/work-days/team")
  listTeamWorkDays(...ListTeamWorkDaysRequest): ListTeamWorkDaysResponse;

  @post
  @route("/work-days/{work_day_id}/lines")
  createWorkDayLine(...CreateWorkDayLineRequest): CreateWorkDayLineResponse;

  @patch
  @route("/work-days/lines/{id}")
  updateWorkDayLine(...UpdateWorkDayLineRequest): UpdateWorkDayLineResponse;

  @delete
  @route("/work-days/lines/{id}")
  deleteWorkDayLine(...DeleteWorkDayLineRequest): DeleteWorkDayLineResponse;

  @get
  @route("/work-types")
  listWorkTypes(...ListWorkTypesRequest): ListWorkTypesResponse;

  @post
  @route("/work-types")
  createWorkType(...CreateWorkTypeRequest): CreateWorkTypeResponse;

  @put
  @route("/work-types/{id}")
  updateWorkType(...UpdateWorkTypeRequest): UpdateWorkTypeResponse;

  @delete
  @route("/work-types/{id}")
  deleteWorkType(...DeleteWorkTypeRequest): DeleteWorkTypeResponse;
}
