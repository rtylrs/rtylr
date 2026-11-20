import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.equipment.v1;

/** EquipmentService defines HR equipment endpoints. */
@route("/hr/equipment/v1")
interface EquipmentService {
  /** ListAssets returns equipment assets with optional filters. */
  @get
  @route("/assets")
  listAssets(...ListAssetsRequest): ListAssetsResponse;

  /** ListEquipmentRequests returns equipment requests with optional filters. */
  @get
  @route("/requests")
  listEquipmentRequests(
    ...ListEquipmentRequestsRequest,
  ): ListEquipmentRequestsResponse;

  /** CreateEquipmentRequest submits an equipment request. */
  @post
  @route("/requests")
  createEquipmentRequest(
    ...CreateEquipmentRequestRequest,
  ): CreateEquipmentRequestResponse;

  /** ApproveEquipment approves an equipment request. */
  @patch
  @route("/requests/{equipment_request_id}/approve")
  approveEquipment(...ApproveEquipmentRequest): ApproveEquipmentResponse;
}
