import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.compat.v1;

model MyEmployee {
  id: string;
}

model EmptyRequest {}

/** Empty response for \@http routes whose handler writes its own response. */
model EmptyResponse {}

model JsonResponse {
  @Onkt.unwrap
  data: unknown;
}

model QueryRequest {
  @query("id")
  id?: string;

  @query("page")
  page?: int32;

  @query("page_size")
  page_size?: int32;

  @query("limit")
  limit?: int32;

  @query("offset")
  offset?: int32;

  @query("search")
  search?: string;

  @query("status")
  status?: string;

  @query("type")
  type?: string;

  @query("action")
  action?: string;

  @query("view")
  view?: string;

  @query("date")
  date?: string;

  @query("days")
  days?: int32;

  @query("start_date")
  start_date?: string;

  @query("end_date")
  end_date?: string;

  @query("from")
  from_date?: string;

  @query("to")
  to_date?: string;

  @query("employee_id")
  employee_id?: string;

  @query("department_id")
  department_id?: string;

  @query("location_id")
  location_id?: string;

  @query("cycle_id")
  cycle_id?: string;

  @query("goal_id")
  goal_id?: string;

  @query("account_id")
  account_id?: string;

  @query("week_start")
  week_start?: string;
}

model BodyRequest {
  @body
  body: unknown;
}

model PathRequest {
  id: string;
}

model PathBodyRequest {
  id: string;

  @body
  body: unknown;
}

model QueryPathRequest {
  id: string;

  @query("status")
  status?: string;
}

model PathIdRequest {
  id: string;
}

model PathIdBodyRequest {
  id: string;

  @body
  body: unknown;
}
