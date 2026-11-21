import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace recruit.career.v1;

/**
 * RecruitCareerService defines the read surface for recruit submissions and
 * the status transition endpoint. Create/update of careers and sections stay
 * on legacy routes until that migration happens.
 */
@route("/recruit/career/v1")
interface RecruitCareerService {
  /** Returns one page of submissions for this business. */
  @get
  @route("/submissions")
  listSubmissions(...ListSubmissionsRequest): ListSubmissionsResponse;

  /** Advances a candidate through the hiring funnel. */
  @patch
  @route("/submissions/status")
  updateSubmissionStatus(
    ...UpdateSubmissionStatusRequest,
  ): UpdateSubmissionStatusResponse;
}
