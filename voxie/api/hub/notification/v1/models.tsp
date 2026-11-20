import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.notification.v1;

/** Notification is a single hub notification entry. */
model Notification {
  @encode(string)
  id: int64;

  title: string;
  body: string;

  /** info, warning, success, error */
  type: string;

  read: boolean;

  /** When the record was created. */
  created_at: string;

  link: string;
}

model ListNotificationsRequest {
  /** Maximum records to return. Defaults to 50 and is capped by the server. */
  @query("limit")
  limit: int32;

  /** Opaque continuation token from a previous response. */
  @query("cursor")
  cursor: string;
}

model ListNotificationsResponse {
  notifications: Notification[];
  next_cursor: string;
}

model MarkAllReadRequest {}

model MarkAllReadResponse {
  message: string;
}

model MarkNotificationReadRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model MarkNotificationReadResponse {
  message: string;
}

model MarkNotificationsReadRequest {
  @minItems(1)
  @Onkt.encode("string")
  ids: int64[];
}

model MarkNotificationsReadResponse {
  message: string;
}
