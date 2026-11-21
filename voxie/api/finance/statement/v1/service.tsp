import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.statement.v1;

@route("/finance/statement/v1")
interface StatementService {
  @get
  @route("/balance-sheet")
  getBalanceSheet(...GetBalanceSheetRequest): GetBalanceSheetResponse;

  @get
  @route("/cash-flow")
  getCashFlow(...GetCashFlowRequest): GetCashFlowResponse;
}
