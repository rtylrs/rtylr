import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.inbox.v1;

model InboxItem {
  ticket_id: int64;
  customer_name: string;
  customer_email: string;
  subject: string;
  status: string;
  priority: string;
  channel: string;
  last_message: string;
  last_message_at: string;
  message_count: int32;
}

model ListInboxRequest {
  @query("page")
  page: int32;

  @query("page_size")
  page_size: int32;
}

model ListInboxResponse {
  inbox: InboxItem[];
  total: int32;
  page: int32;
  page_size: int32;
}

model GetUnreadCountRequest {}

model GetUnreadCountResponse {
  count: int32;
}
