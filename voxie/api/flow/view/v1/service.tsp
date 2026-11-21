import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.view.v1;

@route("/flow/view/v1")
interface ViewService {
  @get
  @route("/views")
  listViews(...ListViewsRequest): ListViewsResponse;

  @post
  @route("/views")
  createView(...CreateViewRequest): CreateViewResponse;

  @put
  @route("/views/{id}")
  updateView(...UpdateViewRequest): UpdateViewResponse;

  @delete
  @route("/views/{id}")
  deleteView(...DeleteViewRequest): DeleteViewResponse;
}
