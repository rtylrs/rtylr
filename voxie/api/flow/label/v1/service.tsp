import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.label.v1;

@route("/flow/label/v1")
interface LabelService {
  @get
  @route("/labels")
  listLabels(...ListLabelsRequest): ListLabelsResponse;

  @post
  @route("/labels")
  createLabel(...CreateLabelRequest): CreateLabelResponse;

  @put
  @route("/labels/{id}")
  updateLabel(...UpdateLabelRequest): UpdateLabelResponse;

  @delete
  @route("/labels/{id}")
  deleteLabel(...DeleteLabelRequest): DeleteLabelResponse;

  @put
  @route("/tasks/labels")
  setTaskLabels(...SetTaskLabelsRequest): SetTaskLabelsResponse;
}
