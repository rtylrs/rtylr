import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.review.v1;

model Review {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  staff_id: int64;

  @encode(string)
  reviewer_id: int64;

  period: string;
  rating: float64;
  goals_rating: float64;
  values_rating: float64;
  skills_rating: float64;
  notes: string;
  status: string;
  created_at: string;
  updated_at: string;
}

model ListReviewsRequest {}

model ListReviewsResponse {
  reviews: Review[];
}

model CreateReviewRequest {
  @Onkt.required
  @encode(string)
  staff_id?: int64;

  @encode(string)
  reviewer_id: int64;

  @Onkt.required
  period: string;

  rating: float64;
  goals_rating: float64;
  values_rating: float64;
  skills_rating: float64;
  notes: string;
  status: string;
}

model CreateReviewResponse {
  @encode(string)
  id: int64;

  message: string;
}
