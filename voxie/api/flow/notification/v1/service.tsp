import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.notification.v1;

@route("/flow/notification/v1")
interface NotificationService {
  @get
  @route("/notifications")
  listNotifications(...ListNotificationsRequest): ListNotificationsResponse;

  @put
  @route("/notifications/{id}/read")
  markNotificationRead(
    ...MarkNotificationReadRequest,
  ): MarkNotificationReadResponse;

  @put
  @route("/notifications/read-all")
  markAllNotificationsRead(
    ...MarkAllNotificationsReadRequest,
  ): MarkAllNotificationsReadResponse;
}
