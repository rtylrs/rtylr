import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.performance.v1;

/** PerformanceService defines HR performance review endpoints. */
@route("/hr/performance/v1")
interface PerformanceService {
  /** Returns review cycles. */
  @get
  @route("/review-cycles")
  listReviewCycles(...ListReviewCyclesRequest): ListReviewCyclesResponse;

  /** Creates a review cycle. */
  @post
  @route("/review-cycles")
  createReviewCycle(...CreateReviewCycleRequest): CreateReviewCycleResponse;

  /** Opens a review cycle for submissions. */
  @post
  @route("/review-cycles/open")
  openReviewCycle(...OpenReviewCycleRequest): OpenReviewCycleResponse;

  /** Returns performance goals with optional filters. */
  @get
  @route("/goals")
  listGoals(...ListGoalsRequest): ListGoalsResponse;

  /** Submits a performance review. */
  @post
  @route("/reviews")
  submitReview(...SubmitReviewRequest): SubmitReviewResponse;
}
