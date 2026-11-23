import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pager.v1;

/**
 * The order pager: a customer scans the venue's QR and gets a match number, the
 * cashier labels that phone's ticket with the order number or the customer's
 * name, and the phone rings when the staff mark it ready. A venue is an
 * ordinary business.
 */
@route("/pager/v1")
interface PagerService {
  // Public routes are anonymous. Budgets are per client and per concrete
  // path, so each venue and each ticket has its own.
  /** Returns the venue behind a pager link. */
  @get
  @route("/public/venues/{slug}")
  @Onkt.meta("rate_limit", "60/m")
  getPublicVenue(...GetPublicVenueRequest): GetPublicVenueResponse;

  /** Opens a waiting ticket for the phone that scanned the venue's QR. */
  @post
  @route("/public/venues/{slug}/scan")
  @Onkt.meta("rate_limit", "60/m")
  scanTicket(...ScanTicketRequest): ScanTicketResponse;

  /** Opens a ticket from the number on the customer's receipt, when the venue allows it. */
  @post
  @route("/public/venues/{slug}/join")
  @Onkt.meta("rate_limit", "60/m")
  joinTicket(...JoinTicketRequest): JoinTicketResponse;

  /** Returns the ticket a phone is following. Polled every few seconds. */
  @get
  @route("/public/tickets/{public_token}")
  @Onkt.meta("rate_limit", "120/m")
  getPublicTicket(...GetPublicTicketRequest): GetPublicTicketResponse;

  /** Silences the bell on every phone following the ticket. The status is unchanged. */
  @post
  @route("/public/tickets/{public_token}/acknowledge")
  @Onkt.meta("rate_limit", "30/m")
  acknowledgeTicket(...AcknowledgeTicketRequest): AcknowledgeTicketResponse;

  /** Returns the key browsers subscribe to lock-screen notifications with. */
  @get
  @route("/public/push-key")
  @Onkt.meta("rate_limit", "60/m")
  getPushKey(...GetPushKeyRequest): GetPushKeyResponse;

  /** Notifies this browser when the ticket turns ready and on every call again. */
  @post
  @route("/public/tickets/{public_token}/push")
  @Onkt.meta("rate_limit", "30/m")
  subscribePush(...SubscribePushRequest): SubscribePushResponse;

  // Staff routes act on the signed-in caller's business.
  /** Lists the open tickets, labelled or not, oldest first. */
  @get
  @route("/tickets")
  listTickets(...ListTicketsRequest): ListTicketsResponse;

  /** Labels a waiting ticket with the order number or the customer's name. */
  @post
  @route("/tickets/{id}/label")
  labelTicket(...LabelTicketRequest): LabelTicketResponse;

  /** Marks a waiting ticket ready, which rings the phones following it. */
  @post
  @route("/tickets/{id}/ready")
  markReady(...MarkReadyRequest): MarkReadyResponse;

  /** Rings the phones of a ready ticket again. */
  @post
  @route("/tickets/{id}/remind")
  remindTicket(...RemindTicketRequest): RemindTicketResponse;

  /** Puts a ready ticket back to waiting. */
  @post
  @route("/tickets/{id}/undo")
  undoReady(...UndoReadyRequest): UndoReadyResponse;

  /** Closes a ticket as collected and frees its number. */
  @post
  @route("/tickets/{id}/collect")
  collectTicket(...CollectTicketRequest): CollectTicketResponse;

  /** How the business's orders went: counts, waits and busy hours. */
  @get
  @route("/stats")
  getStats(...GetStatsRequest): GetStatsResponse;

  /** Returns the business's pager settings. */
  @get
  @route("/settings")
  getSettings(...GetSettingsRequest): GetSettingsResponse;

  /** Updates the business's pager settings. Owners and admins only. */
  @put
  @route("/settings")
  updateSettings(...UpdateSettingsRequest): UpdateSettingsResponse;
}
