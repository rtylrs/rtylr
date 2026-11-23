import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pager.v1;

// Times are RFC 3339 UTC strings with millisecond precision, and absent until
// the moment they describe has happened. Statuses are waiting, ready,
// collected and expired.
//
// A ticket starts when a phone scans the venue's QR: it is waiting, has no
// label yet and shows a match code: a number such as "427". The customer shows
// the number, the cashier finds it on the board and gives the ticket its label,
// the order number or the customer's name. An unlabelled ticket expires after
// 15 minutes, a labelled one after 1 hour.
//
// Errors carry a stable `code` next to `message`: venue_not_found (404),
// ticket_not_found (404), invalid_label (400), label_taken (409),
// invalid_label_mode (400), invalid_transition (409),
// too_many_open_tickets (429), invalid_push_subscription (400) and
// too_many_push_subscriptions (429).
/** A ticket as the venue's staff see it. */
model Ticket {
  @encode(string)
  id: int64;

  /** The order number or the customer's name. Absent until the cashier labels the ticket. */
  label: string;

  /** number or name; absent while the ticket has no label. */
  label_kind: string;

  /**
   * The match code the phone shows, to be shown exactly as sent: three ASCII
   * digits such as "427", four if every three digit number is taken. Unique
   * among the venue's open unlabelled tickets. A ticket opened before the
   * numbers sends its old code in capitals ("PANDA2", "A7"). Absent once the
   * ticket is labelled.
   */
  code: string;

  /** One of waiting, ready, collected, expired. */
  status: string;

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
}

model ScanTicketRequest {
  slug: string;
}

/** A new waiting ticket for the phone that scanned the venue's QR. */
model ScanTicketResponse {
  /** Opaque credential for this ticket, the only one the phone holds. */
  public_token: string;

  /** The match code the customer shows the cashier, as on Ticket. */
  code: string;

  status: string;
}

model GetPublicTicketRequest {
  public_token: string;
}

model GetPublicTicketResponse {
  /** Absent until the cashier labels the ticket; until then show the code. */
  label: string;

  /** number or name; absent while the ticket has no label. */
  label_kind: string;

  /** The match code, as on Ticket. Absent once the ticket is labelled. */
  code: string;
  status: string;
  venue_name: string;

  /** Absent when the business has no logo. */
  venue_logo_url: string;

  /**
   * Labelled waiting tickets at the venue that scanned before this one. Zero
   * while this ticket has no label and once it is no longer waiting.
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

model GetPushKeyRequest {}

model GetPushKeyResponse {
  /**
   * The VAPID public key, URL-safe base64. Absent while lock-screen
   * notifications are off on this server; then do not ask for permission.
   */
  public_key: string;
}

/**
 * A browser's Web Push subscription, as PushSubscription.toJSON() reports it.
 * Taken only for an open ticket at a venue whose pager is on, and kept only
 * while the ticket is open. It carries no personal data.
 */
model SubscribePushRequest {
  public_token: string;

  /** Must be an https URL of a known browser push service. */
  @Onkt.required
  @minLength(1)
  @maxLength(1024)
  endpoint: string;

  /** keys.p256dh: the browser's P-256 public key. */
  @Onkt.required
  @minLength(1)
  @maxLength(128)
  p256dh: string;

  /** keys.auth: the browser's 16 byte auth secret. */
  @Onkt.required
  @minLength(1)
  @maxLength(64)
  auth: string;
}

model SubscribePushResponse {}

// ── Staff ─────────────────────────────────────────────────────────
model ListTicketsRequest {}

model ListTicketsResponse {
  /** Open tickets (waiting and ready), oldest first. */
  tickets: Ticket[];

  server_time: string;
}

model LabelTicketRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /**
   * The order number or the customer's name, as the venue's label mode
   * allows. Trimmed; a number is normalised like a receipt number (1 to 12
   * ASCII letters and digits, Arabic-Indic digits as ASCII) and a name is 1
   * to 40 characters. In "either" mode an entry with a digit that is a valid
   * number is a number, anything else a name.
   */
  @Onkt.required
  @minLength(1)
  @maxLength(64)
  label: string;
}

/**
 * label_taken when another open ticket already has this number. Names may
 * repeat.
 */
model LabelTicketResponse {
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

  /** What the cashier labels a ticket with: number, name or either. */
  label_mode: string;
}

model UpdateSettingsRequest {
  /** Unchanged when absent. */
  enabled?: boolean;

  /** number, name or either. Unchanged when absent. */
  label_mode?: string;
}

model UpdateSettingsResponse {
  enabled: boolean;
  label_mode: string;
}
