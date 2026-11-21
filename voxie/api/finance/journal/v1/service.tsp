import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.journal.v1;

/** JournalService defines journal entry endpoints. */
@route("/finance/journal/v1")
interface JournalService {
  @get
  @route("/entries")
  listJournalEntries(...ListJournalEntriesRequest): ListJournalEntriesResponse;

  @post
  @route("/entries")
  createJournalEntry(...CreateJournalEntryRequest): CreateJournalEntryResponse;

  @get
  @route("/entries/{id}")
  getJournalEntry(...GetJournalEntryRequest): GetJournalEntryResponse;
}
