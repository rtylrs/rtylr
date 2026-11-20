import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.notification.v1;

model Notification {
  @encode(string)
  id: int64;

  type: string;
  title: string;
  description: string;
  task_id: string;
  task_title: string;

  /** The user whose action caused this notification, when the source recorded one. */
  @encode(string)
  actor_id?: int64;

  /** Display name resolved from the acting user's current profile. */
  actor_name: string;

  actor_initials: string;
  project: string;
  project_color: string;
  read: boolean;
  timestamp: string;
}

model ListNotificationsRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListNotificationsResponse {
  notifications: Notification[];
  pagination: PaginationResponse;
}

model MarkNotificationReadRequest {
  @encode(string)
  id: int64;
}

model MarkNotificationReadResponse {}

model MarkAllNotificationsReadRequest {}

model MarkAllNotificationsReadResponse {}
