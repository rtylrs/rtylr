import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.tickets.v1;

model Ticket {
  id: int64;
  business_id: int64;
  customer_name: string;
  customer_email: string;
  subject: string;
  status: string;
  priority: string;
  assigned_to?: int64;
  channel: string;
  created_at: string;
  updated_at: string;
}

model TicketMessage {
  id: int64;
  ticket_id: int64;
  sender_type: string;
  sender_name: string;
  body: string;
  is_internal: boolean;
  created_at: string;
}

model ListTicketsRequest {
  @query("status")
  status: string;

  @query("page")
  page: int32;

  @query("page_size")
  page_size: int32;
}

model ListTicketsResponse {
  tickets: Ticket[];
  total: int32;
  page: int32;
  page_size: int32;
}

model GetTicketRequest {
  @query("id")
  id: int64;
}

model GetTicketResponse {
  ticket: Ticket;
  messages: TicketMessage[];
}

model CreateTicketRequest {
  customer_name: string;
  customer_email: string;
  subject: string;
  priority: string;
  channel: string;
  assigned_to?: int64;
  body: string;
  sender_name: string;
}

model CreateTicketResponse {
  ticket_id: int64;
  message: string;
}

model UpdateTicketRequest {
  id: int64;
  subject: string;
  status: string;
  priority: string;
  assigned_to?: int64;
  channel: string;
}

model UpdateTicketResponse {
  message: string;
}
