import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.customer.v1;

/** CustomerService defines POS customer lookup endpoints. */
@route("/pos/customer/v1")
interface CustomerService {
  /** Searches customers by name, email, or phone for POS lookups. */
  @get
  @route("/customers")
  searchCustomers(...SearchCustomersRequest): SearchCustomersResponse;

  /** Creates a CRM customer directly from POS customer lookup. */
  @post
  @route("/customers")
  createCustomer(@body body: CreateCustomerRequest): CreateCustomerResponse;
}
