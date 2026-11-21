import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.appointment.v1;

@route("/booking/appointment/v1")
interface AppointmentService {
  @get
  @route("/appointments")
  listAppointments(...ListAppointmentsRequest): ListAppointmentsResponse;

  @get
  @route("/appointments/today")
  listTodayAppointments(
    ...ListTodayAppointmentsRequest,
  ): ListTodayAppointmentsResponse;

  @post
  @route("/appointments")
  createAppointment(...CreateAppointmentRequest): CreateAppointmentResponse;

  @put
  @route("/appointments/{id}")
  updateAppointment(...UpdateAppointmentRequest): UpdateAppointmentResponse;

  @delete
  @route("/appointments/{id}")
  deleteAppointment(...DeleteAppointmentRequest): DeleteAppointmentResponse;
}
