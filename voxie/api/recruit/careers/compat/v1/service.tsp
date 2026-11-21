import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace recruit.careers.compat.v1;

/** Legacy public careers routes, served with their existing wire format. */
@route("/careers")
interface CompatibilityService {
  @get
  listCareers(...EmptyRequest): JsonResponse;

  @post
  @route("/apply")
  @Onkt.http
  @Onkt.maxBody("16MiB")
  createApply(...EmptyRequest): EmptyResponse;
}
