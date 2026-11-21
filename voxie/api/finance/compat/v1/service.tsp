import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.compat.v1;

/**
 * Typed compatibility surface for finance routes that predate the canonical
 * finance.* service packages. It preserves their deployed wire paths while
 * moving app transport to generated OneKit clients.
 */
@route("/finance")
interface CompatibilityService {
  @get
  @route("/bank-feeds")
  listBankFeeds(...BankFeedRequest): JsonResponse;

  @post
  @route("/bank-feeds/sync")
  syncBankFeeds(...BodyRequest): JsonResponse;
}
