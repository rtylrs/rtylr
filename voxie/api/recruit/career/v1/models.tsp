import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace recruit.career.v1;

/** RecruitCareer is a job posting on the careers page. */
model RecruitCareer {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  slug: string;
  title: string;
  department: string;
  location: string;
  employment_type: string;
  work_mode: string;
  seniority: string;
  details: string;
  requirements: string;
  salary_min: string;
  salary_max: string;
  salary_currency: string;
  is_published: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/** Submission is a candidate's application to one career. */
model Submission {
  @encode(string)
  id: int64;

  @encode(string)
  career_id: int64;

  candidate_name: string;
  email: string;
  phone: string;
  resume_url: string;
  cover_letter: string;
  status: string;
  career_title: string;
  created_at: string;
  updated_at: string;
}

/** ListSubmissionsRequest selects submissions for this business. */
model ListSubmissionsRequest {
  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

/** ListSubmissionsResponse returns one page of submissions. */
model ListSubmissionsResponse {
  submissions: Submission[];
  total: int64;
  page: int64;
  page_size: int64;
}

/** UpdateSubmissionStatusRequest advances a candidate through the funnel. */
model UpdateSubmissionStatusRequest {
  /** body carries the submission id and new status bound from the HTTP payload. */
  @body
  body: UpdateSubmissionStatusBody;
}

/** UpdateSubmissionStatusBody carries the status transition. */
model UpdateSubmissionStatusBody {
  /** id identifies the submission to update. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  id?: int64;

  /** status is one of: pending | reviewing | interview | offer | hired | rejected. */
  @Onkt.required
  status: string;
}

/** UpdateSubmissionStatusResponse confirms the transition. */
model UpdateSubmissionStatusResponse {
  /** message is a human-readable result. */
  message: string;
}
