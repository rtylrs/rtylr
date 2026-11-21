import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.customer.v1;

/** Customer is a CRM customer profile with loyalty and engagement data. */
model Customer {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  date_of_birth?: string;
  gender: string;
  address: string;
  city: string;
  country: string;

  @encode(string)
  loyalty_points: int64;

  total_spent: float64;

  @encode(string)
  visit_count: int64;

  last_visit_at?: string;
  loyalty_tier: string;
  tags: string[];
  notes: string;

  /** Explicit communication consent captured at customer creation/update. */
  consent: Record<boolean>;

  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model CreateCustomerRequest {
  @Onkt.required
  first_name: string;

  last_name: string;

  /** Not yet enforced server-side; clients may send "". */
  email: string;

  phone: string;
  date_of_birth: string;
  gender: string;
  address: string;
  city: string;
  country: string;
  tags: string[];
  notes: string;
  consent: Record<boolean>;
}

model CreateCustomerResponse {
  @encode(string)
  customer_id: int64;

  message: string;
}

model DeleteCustomerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteCustomerResponse {
  message: string;
}

model GetCustomerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetCustomerResponse {
  customer: Customer;
}

model ListCustomersRequest {
  /** Filters customers by name, email, or phone. */
  @query("search")
  search: string;

  /** Filters by loyalty tier. */
  @query("tier")
  tier: string;

  /** Filters by customer tag. */
  @query("tag")
  tag: string;

  @query("page")
  @encode(string)
  page: int64;

  @query("page_size")
  @encode(string)
  page_size: int64;
}

model ListCustomersResponse {
  customers: Customer[];

  @encode(string)
  total: int64;

  @encode(string)
  page: int64;

  @encode(string)
  page_size: int64;
}

model UpdateCustomerBody {
  @Onkt.required
  first_name: string;

  last_name: string;

  /** Not yet enforced server-side; clients may send "". */
  email: string;

  phone: string;
  date_of_birth: string;
  gender: string;
  address: string;
  city: string;
  country: string;
  tags: string[];
  notes: string;
  consent: Record<boolean>;
  is_active?: boolean;
}

model UpdateCustomerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateCustomerBody;
}

model UpdateCustomerResponse {
  message: string;
}
