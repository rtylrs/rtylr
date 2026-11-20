import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.customer.v1;

/** CustomerService defines CRM customer endpoints. */
@route("/crm/customer/v1")
interface CustomerService {
  /** Returns customers for the active business with optional filters. */
  @get
  @route("/customers")
  listCustomers(...ListCustomersRequest): ListCustomersResponse;

  /** Creates a new customer record. */
  @post
  @route("/customers")
  createCustomer(...CreateCustomerRequest): CreateCustomerResponse;

  /** Returns a single customer by ID. */
  @get
  @route("/customers/{id}")
  getCustomer(...GetCustomerRequest): GetCustomerResponse;

  /** Updates an existing customer. */
  @put
  @route("/customers/{id}")
  updateCustomer(...UpdateCustomerRequest): UpdateCustomerResponse;

  /** Removes a customer by ID. */
  @delete
  @route("/customers/{id}")
  deleteCustomer(...DeleteCustomerRequest): DeleteCustomerResponse;
}
