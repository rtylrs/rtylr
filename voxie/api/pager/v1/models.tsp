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
// too_many_open_tickets (429), invalid_push_subscription (400),
// too_many_push_subscriptions (429), join_by_number_off (409),
// type_your_number (409, a scan at a venue whose customers type their
// receipt number), invalid_review_url (400), invalid_prep_minutes (400) and
// invalid_ringtone (400).
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

  /** When the venue expects the order ready; absent when it gave no estimate. */
  expected_ready_at: string;
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

  /**
   * True when customers join by typing the number on their receipt instead
   * of scanning for a code the cashier labels.
   */
  customer_numbers: boolean;
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

model JoinTicketRequest {
  slug: string;

  /**
   * The order number on the customer's receipt, normalised like a label
   * number. Only while the venue lets customers type their number
   * (join_by_number_off otherwise); label_taken when that number already has
   * an open ticket, so a second phone cannot follow someone else's order.
   */
  @Onkt.required
  @minLength(1)
  @maxLength(32)
  number: string;
}

/** A new waiting ticket, already labelled with the customer's number. */
model JoinTicketResponse {
  public_token: string;
  label: string;
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

  /** When the venue expects the order ready; absent when it gave no estimate. */
  expected_ready_at: string;

  /** The venue's review page, offered once the order is in hand; absent when unset. */
  review_url: string;

  /** The venue's ringtone: bell, chime, marimba, ring or beep. The customer may pick another. */
  ringtone: string;

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

  /**
   * Minutes until the order should be ready, 0 for no estimate. The venue's
   * prep_minutes when absent. Labelling again starts the estimate again.
   */
  @minValue(0)
  @maxValue(120)
  prep_minutes?: int32;
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

  /** The venue's review page shown after pickup; absent when unset. */
  review_url: string;

  /** The usual minutes from labelling to ready; 0 for no estimate. */
  prep_minutes: int32;

  /** Customers type their receipt number instead of scanning for a code. */
  customer_numbers: boolean;

  /** What the customers' phones ring with: bell, chime, marimba, ring or beep. */
  ringtone: string;
}

model UpdateSettingsRequest {
  /** Unchanged when absent. */
  enabled?: boolean;

  /** number, name or either. Unchanged when absent. */
  label_mode?: string;

  /** An https link of up to 500 characters; empty clears it. Unchanged when absent. */
  @maxLength(500)
  review_url?: string;

  /** 0 to 120. Unchanged when absent. */
  @minValue(0)
  @maxValue(120)
  prep_minutes?: int32;

  /** Unchanged when absent. */
  customer_numbers?: boolean;

  /** bell, chime, marimba, ring or beep. Unchanged when absent. */
  @maxLength(16)
  ringtone?: string;
}

model UpdateSettingsResponse {
  enabled: boolean;
  label_mode: string;
  review_url: string;
  prep_minutes: int32;
  customer_numbers: boolean;
  ringtone: string;
}

model GetStatsRequest {
  /** How many days back, today included: 1 to 90, 30 when absent. */
  @query("days")
  days?: int32;
}

/** Orders a day, counted in the venue's own time zone. */
model DayCount {
  /** The local date, YYYY-MM-DD. */
  date: string;

  orders: int32;
}

/**
 * How the venue's orders went over the period: labelled tickets only, and
 * waits measured from the scan or join to Ready.
 */
model GetStatsResponse {
  days: int32;

  /** The IANA zone the hours and dates are counted in. */
  timezone: string;

  orders: int32;

  /** Orders that were marked ready, which the waits are measured on. */
  ready_orders: int32;

  avg_wait_seconds: int32;
  median_wait_seconds: int32;
  longest_wait_seconds: int32;

  /** Orders by the local hour they started in, 24 entries from 00:00. */
  by_hour: int32[];

  /** One entry per day of the period, oldest first. */
  by_day: DayCount[];
}
