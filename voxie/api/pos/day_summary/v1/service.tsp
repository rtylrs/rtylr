import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.day_summary.v1;

/** DaySummaryService defines POS end-of-day summary endpoints. */
@route("/pos/day-summary/v1")
interface DaySummaryService {
  /** Returns aggregate sales, tax, tips, and payment breakdown for a given day. */
  @get
  @route("/day-summary")
  daySummary(...DaySummaryRequest): DaySummaryResponse;
}
