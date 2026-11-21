import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.cycle.v2;

/**
 * Cycle is a sprint-style time box scoped to ONE project in v2. v1 attached
 * cycles to workspaces with an optional project filter; the removal of
 * workspaces makes the project the only sensible owner.
 */
model Cycle {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /**
   * project_id is the project this cycle belongs to. Always set in v2;
   * workspace-era cycles were migrated onto their filtered project or archived.
   */
  @encode(string)
  project_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** start_date is the start date value for this contract. */
  start_date: string;

  /** end_date is the end date value for this contract. */
  end_date: string;

  /** status is the current lifecycle state. */
  status: string;

  /** scope is the total task count in the cycle. Display projection. */
  scope: int32;

  /** completed is how many of those tasks are done. Display projection. */
  completed: int32;

  /** goal is what the cycle is meant to achieve. Empty when none is set. */
  goal: string;

  /** started_at is when the cycle was started, RFC 3339. Empty until then. */
  started_at: string;

  /** completed_at is when the cycle was completed, RFC 3339. Empty until then. */
  completed_at: string;

  /** totals counts the cycle's tasks by status bucket. */
  totals: CycleTotals;

  /**
   * estimate_minutes is the summed estimate of the cycle's tasks: a task's
   * effort_minutes, or its estimated_hours when it has no effort_minutes.
   */
  estimate_minutes: int32;

  /**
   * carried_over lists the tasks this cycle received from cycles that were
   * completed with a move decision, grouped by the cycle they came from.
   */
  carried_over: CycleCarryOver[];
}

/**
 * CycleTotals counts tasks by the bucket of their status (flow_statuses.bucket,
 * or the built-in status when the task has no custom status).
 */
model CycleTotals {
  backlog: int32;
  active: int32;
  done: int32;
  cancelled: int32;
}

model CycleCarryOver {
  /** from_cycle_id is the completed cycle the tasks came from. */
  @encode(string)
  from_cycle_id: int64;

  from_cycle_name: string;

  /** task_ids are the tasks carried over that are still in this cycle. */
  @Onkt.encode("string")
  task_ids: int64[];
}

model ListCyclesRequest {
  /** project_id lists the cycles of one project. */
  @Onkt.required
  @minValueExclusive(0)
  @query("project_id")
  @encode(string)
  project_id?: int64;

  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListCyclesResponse {
  cycles: Cycle[];
  pagination: PaginationResponse;
}

model CreateCycleRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  start_date: string;
  end_date: string;

  /** goal is what the cycle is meant to achieve. Optional. */
  goal: string;
}

model CreateCycleResponse {
  cycle: Cycle;
}

model UpdateCycleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** name renames the cycle. Empty keeps the current name. */
  name: string;

  /**
   * status is kept for older clients. Use startCycle and completeCycle to move
   * a cycle through its lifecycle; they enforce one active cycle per project
   * and settle unfinished tasks.
   */
  status: string;

  /** start_date as YYYY-MM-DD. Empty keeps the current date. */
  start_date: string;

  /** end_date as YYYY-MM-DD. Empty keeps the current date. */
  end_date: string;

  /** goal replaces the goal. Absent keeps it, an empty string clears it. */
  goal?: string;
}

model UpdateCycleResponse {}

model GetCycleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetCycleResponse {
  cycle: Cycle;
}

model StartCycleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** start_date as YYYY-MM-DD. Empty starts the cycle today. */
  start_date: string;

  /**
   * length_weeks is the cycle length, 1 to 4 weeks, counted from start_date.
   * Zero keeps the cycle's stored length, else two weeks.
   */
  @minValue(0)
  @maxValue(4)
  length_weeks: int32;

  /** goal replaces the goal. Absent keeps it, an empty string clears it. */
  goal?: string;
}

/**
 * Answers 409 when the project already has an active cycle, or when the cycle
 * is not upcoming.
 */
model StartCycleResponse {
  cycle: Cycle;
}

/**
 * CycleTaskDecision says what happens to one unfinished task when its cycle is
 * completed.
 */
model CycleTaskDecision {
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** decision is "move" (to to_cycle_id), "backlog" or "cancel". */
  @Onkt.required
  decision: string;

  /**
   * to_cycle_id is the upcoming or active cycle of the same project that
   * receives the task. Required for "move", ignored otherwise.
   */
  @encode(string)
  to_cycle_id?: int64;
}

model CompleteCycleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /**
   * decisions covers the cycle's unfinished tasks (status bucket backlog or
   * active). A task left out goes back to the backlog. A decision for a task
   * that is finished or not in the cycle is rejected and nothing is saved.
   */
  decisions: CycleTaskDecision[];
}

model CompleteCycleResponse {
  cycle: Cycle;
}

model MoveCycleTaskRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  task_id?: int64;

  /**
   * cycle_id is the upcoming or active cycle to place the task in. Absent or
   * zero places it in the unplanned backlog.
   */
  @encode(string)
  cycle_id?: int64;

  /**
   * before_task_id places the task directly before this task of the target
   * list. Give at most one of before_task_id and after_task_id; with neither,
   * the task goes to the end of a cycle or the top of the backlog.
   */
  @encode(string)
  before_task_id?: int64;

  /** after_task_id places the task directly after this task of the target list. */
  @encode(string)
  after_task_id?: int64;
}

model MoveCycleTaskResponse {
  /** cycle_id is the cycle the task is in now, absent for the backlog. */
  @encode(string)
  cycle_id?: int64;

  /** plan_rank is the task's new position value in that list. */
  @encode(string)
  plan_rank: int64;
}

model GetCycleCapacityRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model CycleMemberCapacity {
  @encode(string)
  user_id: int64;

  name: string;

  /** working_days is the number of Monday to Friday days in the cycle window. */
  working_days: int32;

  /**
   * leave_days is how many of those days are covered by approved leave. Zero
   * when leave_applied is false.
   */
  leave_days: int32;

  /** available_minutes is (working_days - leave_days) x minutes_per_day. */
  available_minutes: int32;

  /** committed_minutes is the summed estimate of the member's cycle tasks. */
  committed_minutes: int32;

  /** task_count is how many of the cycle's tasks are assigned to the member. */
  task_count: int32;
}

model GetCycleCapacityResponse {
  members: CycleMemberCapacity[];

  /**
   * leave_applied is false when approved leave could not be read, in which
   * case availability counts working days only.
   */
  leave_applied: boolean;

  /** minutes_per_day is the working day length the calculation used. */
  minutes_per_day: int32;

  /** unassigned_minutes is the summed estimate of the cycle's unassigned tasks. */
  unassigned_minutes: int32;

  /** unassigned_count is how many of the cycle's tasks have no assignee. */
  unassigned_count: int32;
}
