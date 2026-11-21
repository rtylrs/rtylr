import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace finance.journal.v1;

// ── Journal Entries ─────────────────────────────────────────────────
model JournalEntry {
  @encode(string)
  id: int64;

  number: string;
  date: string;
  description: string;
  lines: JournalLine[];

  /** draft, posted */
  status: string;

  /** When the record was created. */
  created_at: string;
}

model JournalLine {
  @encode(string)
  account_id: int64;

  account_name: string;
  debit: float64;
  credit: float64;
  description: string;
}

model CreateJournalEntryRequest {
  @Onkt.required
  date: string;

  @Onkt.required
  description: string;

  @minItems(1)
  lines: JournalLine[];
}

model CreateJournalEntryResponse {
  entry: JournalEntry;
}

model GetJournalEntryRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetJournalEntryResponse {
  entry: JournalEntry;
}

model ListJournalEntriesRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;

  /** The inclusive start date filter. */
  @query("from_date")
  from_date: string;

  /** The inclusive end date filter. */
  @query("to_date")
  to_date: string;
}

model ListJournalEntriesResponse {
  entries: JournalEntry[];
  pagination: PaginationResponse;
}
