import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.reconciliation.v1;

@route("/finance/reconciliation/v1")
interface ReconciliationService {
  @get
  @route("/reconciliation")
  getOverview(...GetOverviewRequest): GetOverviewResponse;

  @post
  @route("/reconciliation/match")
  match(...MatchRequest): MatchResponse;
}
