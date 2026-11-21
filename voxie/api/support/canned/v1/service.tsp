import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.canned.v1;

@route("/support/canned/v1")
interface CannedResponseService {
  @get
  @route("/canned-responses")
  listCannedResponses(
    ...ListCannedResponsesRequest,
  ): ListCannedResponsesResponse;

  @post
  @route("/canned-responses")
  createCannedResponse(
    ...CreateCannedResponseRequest,
  ): CreateCannedResponseResponse;

  @put
  @route("/canned-responses/{id}")
  updateCannedResponse(
    ...UpdateCannedResponseRequest,
  ): UpdateCannedResponseResponse;

  @delete
  @route("/canned-responses/{id}")
  deleteCannedResponse(
    ...DeleteCannedResponseRequest,
  ): DeleteCannedResponseResponse;
}
