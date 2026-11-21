import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.one_on_one.v1;

model OneOnOne {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  manager_id: int64;

  @encode(string)
  staff_id: int64;

  scheduled_at: string;
  notes: string;
  action_items: string;
  status: string;
  created_at: string;
}

model ListOneOnOnesRequest {}

model ListOneOnOnesResponse {
  one_on_ones: OneOnOne[];
}

model CreateOneOnOneRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  manager_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  staff_id?: int64;

  @Onkt.required
  scheduled_at: string;

  notes: string;
  action_items: string;
  status: string;
}

model CreateOneOnOneResponse {
  @encode(string)
  id: int64;

  message: string;
}
