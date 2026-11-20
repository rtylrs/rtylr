import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.work_schedule.v1;

/** WorkScheduleService defines HR work schedule endpoints. */
@route("/hr/work-schedule/v1")
interface WorkScheduleService {
  /** Lists work schedules for the tenant. */
  @get
  @route("/work-schedules")
  listWorkSchedules(...ListWorkSchedulesRequest): ListWorkSchedulesResponse;

  /** Creates a new work schedule. */
  @post
  @route("/work-schedules")
  createWorkSchedule(...CreateWorkScheduleRequest): CreateWorkScheduleResponse;

  /** Updates an existing work schedule. */
  @put
  @route("/work-schedules/{id}")
  updateWorkSchedule(...UpdateWorkScheduleRequest): UpdateWorkScheduleResponse;
}
