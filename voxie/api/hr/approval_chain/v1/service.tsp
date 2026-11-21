import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.approval_chain.v1;

/** ApprovalChainService defines HR approval chain endpoints. */
@route("/hr/approval-chain/v1")
interface ApprovalChainService {
  /** UpsertApprovalChain creates or replaces the approval chain configuration. */
  @put
  @route("/approval-chains")
  upsertApprovalChain(
    ...UpsertApprovalChainRequest,
  ): UpsertApprovalChainResponse;
}
