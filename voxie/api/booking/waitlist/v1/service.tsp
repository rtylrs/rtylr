import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.waitlist.v1;

@route("/booking/waitlist/v1")
interface WaitlistService {
  @get
  @route("/entries")
  listEntries(...ListEntriesRequest): ListEntriesResponse;

  @post
  @route("/entries")
  createEntry(...CreateEntryRequest): CreateEntryResponse;

  @put
  @route("/entries/{id}")
  updateEntry(...UpdateEntryRequest): UpdateEntryResponse;
}
