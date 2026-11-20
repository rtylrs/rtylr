import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace hr.benefit.v1;

model BenefitPlan {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  type: string;
  provider: string;
  cost_employer: float64;
  cost_employee: float64;
  is_active: boolean;
  created_at: string;
}

model CreateBenefitRequest {
  @Onkt.required
  name: string;

  type: string;
  provider: string;
  cost_employer: float64;
  cost_employee: float64;
  is_active: boolean;
}

model CreateBenefitResponse {
  @encode(string)
  id: int64;

  message: string;
}

model ListBenefitsRequest {}

model ListBenefitsResponse {
  plans: BenefitPlan[];
}

model BenefitEnrollment {
  /** id is the enrollment record's unique identifier. */
  @encode(string)
  id: int64;

  /** owner_id is the business that owns the underlying BenefitPlan. */
  @encode(string)
  owner_id: int64;

  /** employee_id is the enrolled employee; always resolved from auth context, never client-supplied. */
  @encode(string)
  employee_id: int64;

  /** plan_id references the BenefitPlan this enrollment is for. */
  @encode(string)
  plan_id: int64;

  /** status is the enrollment lifecycle state (active | cancelled). */
  status: string;

  /** enrolled_at is when the employee enrolled, in ISO-8601 format. */
  enrolled_at: string;

  /** cancelled_at is when the enrollment was cancelled, in ISO-8601 format; absent while the enrollment is active. */
  cancelled_at?: string;
}

model EnrollInBenefitRequest {
  /** plan_id is the BenefitPlan to enroll in. The employee is resolved from auth context, not client-supplied. */
  @Onkt.required
  @encode(string)
  plan_id?: int64;
}

model EnrollInBenefitResponse {
  /** enrollment is the created BenefitEnrollment record. */
  enrollment: BenefitEnrollment;
}

model ListMyBenefitEnrollmentsRequest {
  /** page is the current result page. */
  @query("page")
  page: int32;

  /** limit is the maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** sort_by is the field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** sort_dir is the sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListMyBenefitEnrollmentsResponse {
  /** enrollments are the caller's own BenefitEnrollment records. */
  enrollments: BenefitEnrollment[];

  pagination: PaginationResponse;
}
