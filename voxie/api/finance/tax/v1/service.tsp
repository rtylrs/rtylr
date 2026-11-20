import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.tax.v1;

@route("/finance/tax/v1")
interface TaxService {
  @get
  @route("/tax")
  getTaxConfig(...GetTaxConfigRequest): GetTaxConfigResponse;

  @put
  @route("/tax")
  updateTaxConfig(...UpdateTaxConfigRequest): UpdateTaxConfigResponse;

  @get
  @route("/pl")
  getPLReport(...GetPLReportRequest): GetPLReportResponse;
}
