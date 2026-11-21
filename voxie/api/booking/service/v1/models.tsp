import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.service.v1;

model BookingService {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  name: string;
  description: string;
  duration_minutes: int32;
  price: float64;
  color: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model CreateServiceRequest {
  @Onkt.required
  name: string;

  description: string;

  @Onkt.required
  @minValueExclusive(0)
  duration_minutes?: int32;

  @minValue(0)
  price: float64;

  color: string;
  is_active: boolean;
}

model CreateServiceResponse {
  @encode(string)
  id: int64;

  message: string;
}

model DeleteServiceRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteServiceResponse {
  message: string;
}

model ListServicesRequest {}

model ListServicesResponse {
  services: BookingService[];
}

model UpdateServiceRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  name: string;

  description: string;

  @Onkt.required
  @minValueExclusive(0)
  duration_minutes?: int32;

  @minValue(0)
  price: float64;

  color: string;
  is_active: boolean;
}

model UpdateServiceResponse {
  message: string;
}
