import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.performance.v1;

/** ReviewGoalPayload is a performance goal assigned to an employee within a review cycle. */
model ReviewGoalPayload {
  /** Unique identifier for the review goal. */
  @encode(string)
  id: int64;

  /** Tenant owner this goal belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee this goal is assigned to. */
  @encode(string)
  employee_id: int64;

  /** Review cycle this goal is part of. */
  @encode(string)
  review_cycle_id: int64;

  /** Short title summarising the goal. */
  title: string;

  /** Detailed description of the goal and expectations. */
  description: string;

  /** Relative importance of this goal within the cycle; absent if all goals are weighted equally. */
  weight?: float64;

  /** Measurable criteria that define goal achievement; absent if not yet specified. */
  success_criteria?: string;

  /** Goal status (active | achieved | missed). */
  status: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** PerformanceReviewPayload is a single reviewer's evaluation of an employee within a review cycle. */
model PerformanceReviewPayload {
  /** Unique identifier for the performance review. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;

  /** Tenant owner this review belongs to. */
  @encode(string)
  owner_id: int64;

  /** Employee being reviewed. */
  @encode(string)
  employee_id: int64;

  /** Review cycle this review is part of. */
  @encode(string)
  review_cycle_id: int64;

  /** Employee conducting the review. */
  @encode(string)
  reviewer_employee_id: int64;

  /** Type of review (self | manager | peer). */
  review_type: string;

  /** Overall rating assigned by the reviewer; absent while still in draft. */
  overall_rating?: string;

  /** Free-form written feedback from the reviewer; absent while still in draft. */
  narrative?: string;

  /** Workflow status (draft | submitted | completed | shared). */
  status: string;

  /** Unix timestamp when the review was shared with the employee; absent if not yet shared. */
  @encode(string)
  shared_with_employee_at?: int64;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

/** ReviewCyclePayload is a recurring performance review period (e.g. an annual cycle). */
model ReviewCyclePayload {
  /** Unique identifier for the review cycle. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;

  /** Tenant owner this cycle belongs to. */
  @encode(string)
  owner_id: int64;

  /** Human-readable name for the cycle (e.g. "2026 Annual Review"). */
  name: string;

  /** Cycle type (annual | mid_year | probation | adhoc). */
  type: string;

  /** Start of the review period (ISO-8601 date). */
  period_start: string;

  /** End of the review period (ISO-8601 date). */
  period_end: string;

  /** Workflow status (draft | open | in_review | completed). */
  status: string;

  /** Unix timestamp when this record was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this record was last updated. */
  @encode(string)
  updated_at: int64;
}

model ListReviewCyclesRequest {}

model ListReviewCyclesResponse {
  /** Page of review cycles. */
  cycles: ReviewCyclePayload[];

  /** Total number of review cycles across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model CreateReviewCycleRequest {
  /** Human-readable name for the cycle. */
  @Onkt.required
  name: string;

  /** Cycle type (annual | mid_year | probation | adhoc). */
  @Onkt.required
  type: string;

  /** Start of the review period (ISO-8601 date). */
  @Onkt.required
  period_start: string;

  /** End of the review period (ISO-8601 date). */
  @Onkt.required
  period_end: string;
}

model CreateReviewCycleResponse {
  /** The newly created review cycle. */
  cycle: ReviewCyclePayload;
}

model OpenReviewCycleRequest {
  /** ID of the review cycle to open for submissions. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  cycle_id?: int64;
}

model OpenReviewCycleResponse {
  /** The review cycle after transitioning to open status. */
  cycle: ReviewCyclePayload;
}

model ListGoalsRequest {
  /** Filter by employee; absent to include all employees. */
  @query("employee_id")
  @encode(string)
  employee_id?: int64;

  /** Filter by review cycle; absent to include all cycles. */
  @query("review_cycle_id")
  @encode(string)
  review_cycle_id?: int64;
}

model ListGoalsResponse {
  /** Page of review goals matching the filters. */
  goals: ReviewGoalPayload[];

  /** Total number of matching goals across all pages. */
  total: int64;

  /** Page size used for this request. */
  limit: int32;

  /** Offset used for this request. */
  offset: int32;
}

model SubmitReviewRequest {
  /** ID of the performance review being submitted. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  review_id?: int64;

  /** Overall rating assigned by the reviewer. */
  @Onkt.required
  overall_rating: string;

  /** Free-form written feedback from the reviewer. */
  @Onkt.required
  narrative: string;
}

model SubmitReviewResponse {
  /** The submitted performance review. */
  review: PerformanceReviewPayload;
}
