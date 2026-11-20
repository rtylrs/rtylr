import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.dashboard.v1;

model DashboardStats {
  /** total_staff is the total staff value for this contract. */
  @encode(string)
  total_staff: int64;

  /** on_shift_today is the on shift today value for this contract. */
  @encode(string)
  on_shift_today: int64;

  /** pending_time_off is the pending time off value for this contract. */
  @encode(string)
  pending_time_off: int64;

  /** payroll_draft is the payroll draft value for this contract. */
  @encode(string)
  payroll_draft: int64;
}

/** DashboardWidgetPayload is a widget layout definition on the HR dashboard. */
model DashboardWidgetPayload {
  /** Unique identifier for the widget. */
  @encode(string)
  id: int64;

  /** Tenant owner this widget belongs to. */
  @encode(string)
  owner_id: int64;

  /** Role that owns this widget layout. */
  @encode(string)
  role_id: int64;

  /** Widget type identifier (e.g. headcount_chart, leave_summary). */
  widget_type: string;

  /** Horizontal grid position. */
  position_x: int32;

  /** Vertical grid position. */
  position_y: int32;

  /** Widget width in grid units. */
  width: int32;

  /** Widget height in grid units. */
  height: int32;

  /** Widget-specific configuration serialised as JSON. */
  config: bytes;

  /** Unix timestamp when this widget was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this widget was last updated. */
  @encode(string)
  updated_at: int64;
}

/** DashboardDataPayload is pre-loaded widget data keyed by widget type. */
model DashboardDataPayload {
  /** Widget type this data corresponds to. */
  widget_type: string;

  /** Serialised widget data as JSON. */
  data: bytes;
}

model GetHRDashboardRequest {}

model GetHRDashboardResponse {
  stats: DashboardStats;
}

model GetDashboardWidgetsRequest {}

model GetDashboardWidgetsResponse {
  /** Widget layout definitions for the requesting user's dashboard. */
  widgets: DashboardWidgetPayload[];

  /** Pre-loaded data keyed by widget type. */
  data: Record<DashboardDataPayload>;
}

model UpdateDashboardWidgetsRequest {
  /** Updated widget positions and sizes. */
  @minItems(1)
  widgets: DashboardWidgetPayload[];
}

model UpdateDashboardWidgetsResponse {
  /** success indicates whether the operation completed. */
  success: boolean;
}
