import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.setting.v1;

model GetUserSettingsRequest {}

model GetUserSettingsResponse {
  task_assigned: boolean;
  task_comments: boolean;
  mentions: boolean;
  status_changes: boolean;
  weekly_digest: boolean;
  email_notifications: boolean;

  /** JSON-encoded per-user Flow sidebar layout and visibility preferences. */
  sidebar_preferences?: string;
}

model UpdateUserSettingsRequest {
  /**
   * Notify the user when a task is assigned to them. Omit to preserve the
   * stored preference; send false to turn the notification off.
   */
  task_assigned?: boolean;

  /**
   * Notify the user about comments on their tasks. Omit to preserve the
   * stored preference; send false to turn the notification off.
   */
  task_comments?: boolean;

  /**
   * Notify the user when they are mentioned. Omit to preserve the stored
   * preference; send false to turn the notification off.
   */
  mentions?: boolean;

  /**
   * Notify the user when a task status changes. Omit to preserve the stored
   * preference; send false to turn the notification off.
   */
  status_changes?: boolean;

  /**
   * Send the user the weekly digest email. Omit to preserve the stored
   * preference; send false to turn the digest off.
   */
  weekly_digest?: boolean;

  /**
   * Master switch for this user's Flow email notifications. Omit to preserve
   * the stored preference; send false to turn all of them off.
   */
  email_notifications?: boolean;

  /** JSON-encoded per-user Flow sidebar layout and visibility preferences. */
  sidebar_preferences?: string;
}

model UpdateUserSettingsResponse {}
