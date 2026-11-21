import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.service.v1;

@route("/booking/service/v1")
interface BookingServiceService {
  @get
  @route("/services")
  listServices(...ListServicesRequest): ListServicesResponse;

  @post
  @route("/services")
  createService(...CreateServiceRequest): CreateServiceResponse;

  @put
  @route("/services/{id}")
  updateService(...UpdateServiceRequest): UpdateServiceResponse;

  @delete
  @route("/services/{id}")
  deleteService(...DeleteServiceRequest): DeleteServiceResponse;
}
