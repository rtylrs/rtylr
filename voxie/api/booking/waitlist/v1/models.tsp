import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.waitlist.v1;

model WaitlistEntry {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  service_id: int64;

  customer_name: string;
  customer_email: string;
  customer_phone: string;
  preferred_date: string;
  preferred_time: string;
  status: string;
  created_at: string;
}

model CreateEntryRequest {
  @encode(string)
  service_id: int64;

  @Onkt.required
  customer_name: string;

  customer_email: string;
  customer_phone: string;
  preferred_date: string;
  preferred_time: string;
}

model CreateEntryResponse {
  @encode(string)
  id: int64;

  message: string;
}

model ListEntriesRequest {}

model ListEntriesResponse {
  entries: WaitlistEntry[];
}

model UpdateEntryRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @encode(string)
  service_id: int64;

  customer_name: string;
  customer_email: string;
  customer_phone: string;
  preferred_date: string;
  preferred_time: string;
  status: string;
}

model UpdateEntryResponse {
  message: string;
}
