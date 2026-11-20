import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.feedback.v1;

/** Feedback is a single NPS score or written review from a customer. */
model Feedback {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  customer_id: int64;

  /** "nps" or "review" — score's scale and meaning depend on this. */
  type: string;

  /** NPS score (0-10) when type is "nps"; star rating (1-5) when type is "review". */
  score: int32;

  comment: string;
  source: string;
  created_at: string;

  /** Display name of the customer this feedback belongs to. */
  customer_name: string;
}

model CreateFeedbackRequest {
  @encode(string)
  customer_id: int64;

  @Onkt.required
  type: string;

  score: int32;
  comment: string;
  source: string;
}

model CreateFeedbackResponse {
  @encode(string)
  feedback_id: int64;

  message: string;
}

model ListFeedbackRequest {
  @query("type")
  type: string;

  @query("source")
  source: string;

  @query("min_score")
  min_score?: int32;

  @query("max_score")
  max_score?: int32;

  @query("page")
  page: int32;

  @query("page_size")
  page_size: int32;
}

model ListFeedbackResponse {
  feedback: Feedback[];
  total: int32;
  page: int32;
  page_size: int32;
}
