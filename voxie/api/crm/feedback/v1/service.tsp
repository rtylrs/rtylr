import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.feedback.v1;

@route("/crm/feedback/v1")
interface FeedbackService {
  @get
  @route("/feedback")
  listFeedback(...ListFeedbackRequest): ListFeedbackResponse;

  @post
  @route("/feedback")
  createFeedback(...CreateFeedbackRequest): CreateFeedbackResponse;
}
