import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.review.v1;

@route("/hr/review/v1")
interface ReviewService {
  @get
  @route("/reviews")
  listReviews(...ListReviewsRequest): ListReviewsResponse;

  @post
  @route("/reviews")
  createReview(...CreateReviewRequest): CreateReviewResponse;
}
