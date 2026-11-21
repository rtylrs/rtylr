import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.compat.v1;

model EmptyRequest {}

model JsonResponse {
  @Onkt.unwrap
  data: unknown;
}

model PageRequest {
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

  @query("status")
  status?: string;

  @query("category")
  category?: string;

  @query("start")
  start?: string;

  @query("end")
  end?: string;
}

model IdQueryRequest {
  @query("id")
  id?: string;
}

model ReportRequest {
  @query("period")
  period?: string;

  @query("date")
  date?: string;
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
  body: unknown;
}

model BankFeedRequest {
  @query("account_id")
  account_id?: string;
}
