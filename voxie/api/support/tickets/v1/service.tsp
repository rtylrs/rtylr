import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.tickets.v1;

/**
 * Support tickets for the signed-in business. Paths match the routes the
 * legacy handlers served.
 */
@route("/")
interface TicketService {
  @get
  @route("/support/tickets")
  listTickets(...ListTicketsRequest): ListTicketsResponse;

  @get
  @route("/support/tickets/detail")
  getTicket(...GetTicketRequest): GetTicketResponse;

  @post
  @route("/support/tickets")
  createTicket(...CreateTicketRequest): CreateTicketResponse;

  @put
  @route("/support/tickets")
  updateTicket(...UpdateTicketRequest): UpdateTicketResponse;
}
