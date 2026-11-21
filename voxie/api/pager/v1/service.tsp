import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pager.v1;

/**
 * The order pager: a customer scans the venue's poster, enters the number on
 * the receipt and is rung when the staff mark that number ready. A venue is
 * an ordinary business.
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

  /** Follows the ticket with this number, creating it unless the venue is strict. */
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

  // Staff routes act on the signed-in caller's business.
  /** Lists the open tickets, oldest first. */
  @get
  @route("/tickets")
  listTickets(...ListTicketsRequest): ListTicketsResponse;

  /** Enters a number. Returns the existing ticket when the number is already open. */
  @post
  @route("/tickets")
  createTicket(...CreateTicketRequest): CreateTicketResponse;

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

  /** Returns the business's pager settings. */
  @get
  @route("/settings")
  getSettings(...GetSettingsRequest): GetSettingsResponse;

  /** Updates the business's pager settings. Owners and admins only. */
  @put
  @route("/settings")
  updateSettings(...UpdateSettingsRequest): UpdateSettingsResponse;
}
