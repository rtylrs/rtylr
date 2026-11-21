import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.appointment.v1;

model Appointment {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  service_id: int64;

  @encode(string)
  resource_id: int64;

  @encode(string)
  staff_id: int64;

  customer_name: string;
  customer_email: string;
  customer_phone: string;
  start_time: string;
  end_time: string;
  status: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

model CreateAppointmentRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  service_id?: int64;

  @encode(string)
  resource_id: int64;

  @encode(string)
  staff_id: int64;

  @Onkt.required
  customer_name: string;

  customer_email: string;
  customer_phone: string;

  @Onkt.required
  start_time: string;

  @Onkt.required
  end_time: string;

  status: string;
  notes: string;
}

model CreateAppointmentResponse {
  @encode(string)
  id: int64;

  message: string;
}

model DeleteAppointmentRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteAppointmentResponse {
  message: string;
}

model ListAppointmentsRequest {}

model ListAppointmentsResponse {
  appointments: Appointment[];
}

model ListTodayAppointmentsRequest {}

model ListTodayAppointmentsResponse {
  appointments: Appointment[];
}

model UpdateAppointmentRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @encode(string)
  service_id: int64;

  @encode(string)
  resource_id: int64;

  @encode(string)
  staff_id: int64;

  customer_name: string;
  customer_email: string;
  customer_phone: string;
  start_time: string;
  end_time: string;
  status: string;
  notes: string;
}

model UpdateAppointmentResponse {
  message: string;
}
