import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.financial.compat.v1;

/** Compatibility contract for the existing cash-flow time-series route. */
@route("/insights/financial")
interface CompatibilityService {
  @get
  @route("/cash-flow")
  getCashFlow(...GetCashFlowRequest): GetCashFlowResponse;
}
