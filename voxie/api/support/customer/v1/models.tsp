import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.customer.v1;

model Customer {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  name: string;
  email: string;
  phone: string;

  /**
   * customer_id is the CRM customer this support record refers to; absent
   * when the party has not been matched to one. It is nullable because
   * support conversations arrive by email from people who may not be
   * customers yet: matching on email alone is not reliable enough to assert
   * the link, so the server sets it only on a confident match or when a human
   * confirms one.
   */
  @encode(string)
  customer_id?: int64;

  total_tickets: int32;
  last_ticket_at: string;
  created_at: string;
}

model ListCustomersRequest {
  @query("page")
  page: int32;

  @query("page_size")
  page_size: int32;
}

model ListCustomersResponse {
  customers: Customer[];
  total: int32;
  page: int32;
  page_size: int32;
}

model GetCustomerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetCustomerResponse {
  customer: Customer;
}
