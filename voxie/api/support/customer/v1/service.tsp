import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.customer.v1;

@route("/support/customer/v1")
interface CustomerService {
  @get
  @route("/customers")
  listCustomers(...ListCustomersRequest): ListCustomersResponse;

  @get
  @route("/customers/{id}")
  getCustomer(...GetCustomerRequest): GetCustomerResponse;
}
