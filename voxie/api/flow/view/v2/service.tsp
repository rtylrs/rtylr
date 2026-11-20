import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.view.v2;

@route("/flow/view/v2")
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
