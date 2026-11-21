import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.resource.v1;

@route("/booking/resource/v1")
interface ResourceService {
  @get
  @route("/resources")
  listResources(...ListResourcesRequest): ListResourcesResponse;

  @post
  @route("/resources")
  createResource(...CreateResourceRequest): CreateResourceResponse;

  @put
  @route("/resources/{id}")
  updateResource(...UpdateResourceRequest): UpdateResourceResponse;

  @delete
  @route("/resources/{id}")
  deleteResource(...DeleteResourceRequest): DeleteResourceResponse;
}
