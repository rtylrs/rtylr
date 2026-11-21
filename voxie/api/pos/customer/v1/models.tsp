import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.customer.v1;

/**
 * Customer is a lightweight customer record for POS lookups. Named
 * distinctly from crm.customer.v1's Customer (a genuine cross-directory
 * name collision - see crm/segment/v1/models.onk which references the
 * unqualified "Customer" name expecting the full crm.customer.v1 record).
 */
model Customer {
  @encode(string)
  id: int64;

  first_name: string;
  last_name: string;

  /** email is the email address used for lookup or contact. */
  email: string;

  phone: string;
}

model SearchCustomersRequest {
  @Onkt.required
  @query("query")
  query: string;
}

model SearchCustomersResponse {
  customers: Customer[];
}
