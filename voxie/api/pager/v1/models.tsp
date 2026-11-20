import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pager.v1;

// Times are RFC 3339 UTC strings with millisecond precision, and absent until
// the moment they describe has happened. Statuses are waiting, ready,
// collected and expired.
//
// Errors carry a stable `code` next to `message`: venue_not_found (404),
// ticket_not_found (404), invalid_number (400), invalid_transition (409) and
// too_many_open_tickets (429).
/** A ticket as the venue's staff see it. */
model Ticket {
  @encode(string)
  id: int64;

  /** The order number as normalised by the server: ASCII digits and capitals. */
  number: string;

  /** One of waiting, ready, collected, expired. */
  status: string;

  /** Who entered the number first: staff or customer. */
  source: string;

  /** True once a customer's phone has joined the ticket. */
  phone_linked: boolean;

  created_at: string;
  ready_at: string;

  /** Set by every "call again"; a change tells the phones to ring again. */
  reminded_at: string;

  /** Set when a customer silences the bell; cleared by the next reminder. */
  acknowledged_at: string;
}

// ── Public ────────────────────────────────────────────────────────
model GetPublicVenueRequest {
  /** The business slug, the same one the public menu uses. */
  slug: string;
}

/** Answered only while the venue's pager is on; otherwise venue_not_found. */
model GetPublicVenueResponse {
  name: string;

  /** Absent when the business has no logo. */
  logo_url: string;

  enabled: boolean;

  /** When true only numbers the staff already entered can be joined. */
  strict: boolean;
}

model JoinTicketRequest {
  slug: string;

  /**
   * The order number on the receipt. Trimmed, Arabic-Indic digits become
   * ASCII, letters are capitalised; then 1 to 12 ASCII letters and digits.
   */
  @Onkt.required
  @minLength(1)
  @maxLength(64)
  number: string;
}

/**
 * A strict venue answers 404 ticket_not_found until the staff enter the
 * number, so the client can keep retrying for a short while.
 */
model JoinTicketResponse {
  /**
   * Opaque credential for this ticket. Every phone that joins the same
   * ticket receives the same token.
   */
  public_token: string;

  number: string;
  status: string;
}

model GetPublicTicketRequest {
  public_token: string;
}

model GetPublicTicketResponse {
  number: string;
  status: string;
  venue_name: string;

  /**
   * Waiting tickets at the venue entered before this one. Zero once this
   * ticket is no longer waiting.
   */
  ahead: int32;

  ready_at: string;
  reminded_at: string;
  acknowledged_at: string;

  /**
   * When the server built this answer; drop an answer older than the one
   * already shown.
   */
  server_time: string;
}

model AcknowledgeTicketRequest {
  public_token: string;
}

model AcknowledgeTicketResponse {}

// ── Staff ─────────────────────────────────────────────────────────
model ListTicketsRequest {}

model ListTicketsResponse {
  /** Open tickets (waiting and ready), oldest first. */
  tickets: Ticket[];

  server_time: string;
}

model CreateTicketRequest {
  /** Normalised the same way as JoinTicketRequest.number. */
  @Onkt.required
  @minLength(1)
  @maxLength(64)
  number: string;
}

/** Returns the existing ticket when the number is already open. */
model CreateTicketResponse {
  ticket: Ticket;
}

model MarkReadyRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model MarkReadyResponse {
  ticket: Ticket;
}

model RemindTicketRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model RemindTicketResponse {
  ticket: Ticket;
}

model UndoReadyRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model UndoReadyResponse {
  ticket: Ticket;
}

model CollectTicketRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model CollectTicketResponse {
  ticket: Ticket;
}

model GetSettingsRequest {}

model GetSettingsResponse {
  /** Whether the public pager pages answer for this business. */
  enabled: boolean;

  strict: boolean;
}

model UpdateSettingsRequest {
  /** Unchanged when absent. */
  enabled?: boolean;

  /** Unchanged when absent. */
  strict?: boolean;
}

model UpdateSettingsResponse {
  enabled: boolean;
  strict: boolean;
}
