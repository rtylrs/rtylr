import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.terminal.v1;

/** Terminal is a registered POS terminal device. */
model Terminal {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  branch_id: int64;

  name: string;
  device_id?: string;
  is_active: boolean;
  last_seen_at?: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model ListTerminalsRequest {}

model ListTerminalsResponse {
  terminals: Terminal[];
}

model CreateTerminalRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  device_id: string;
}

model CreateTerminalResponse {
  @encode(string)
  terminal_id: int64;

  message: string;
}

model UpdateTerminalBody {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  device_id: string;

  is_active: boolean;
}

model UpdateTerminalRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable terminal fields bound from the HTTP payload. */
  @body
  body: UpdateTerminalBody;
}

model UpdateTerminalResponse {
  message: string;
}

model HeartbeatTerminalRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  terminal_id?: int64;
}

model HeartbeatTerminalResponse {
  message: string;
}
