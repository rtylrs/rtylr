import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.notification.v1;

/** NotificationService defines hub notification endpoints. */
@route("/hub/notification/v1")
interface NotificationService {
  /** Returns notifications for the current user. */
  @get
  @route("/notifications")
  listNotifications(...ListNotificationsRequest): ListNotificationsResponse;

  /** Marks all notifications as read. */
  @patch
  @route("/notifications/read-all")
  markAllRead(...MarkAllReadRequest): MarkAllReadResponse;

  /** Marks a single notification as read. */
  @patch
  @route("/notifications/{id}/read")
  markNotificationRead(
    ...MarkNotificationReadRequest,
  ): MarkNotificationReadResponse;
}
