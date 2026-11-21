import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.benefit.v1;

@route("/hr/benefit/v1")
interface BenefitService {
  @get
  @route("/benefits")
  listBenefits(...ListBenefitsRequest): ListBenefitsResponse;

  @post
  @route("/benefits")
  createBenefit(...CreateBenefitRequest): CreateBenefitResponse;

  /** Enrolls the calling employee in a benefit plan. */
  @post
  @route("/benefits/enrollments")
  enrollInBenefit(...EnrollInBenefitRequest): EnrollInBenefitResponse;

  /** Returns the calling employee's own benefit enrollments. */
  @get
  @route("/benefits/enrollments/mine")
  listMyBenefitEnrollments(
    ...ListMyBenefitEnrollmentsRequest,
  ): ListMyBenefitEnrollmentsResponse;
}
