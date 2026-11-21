import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.template.v1;

/** Template is a reusable message template for campaigns and notifications. */
model Template {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;

  /** Template kind, e.g. "email" or "sms". */
  type: string;

  subject: string;
  body: string;

  /** Arbitrary JSON template variables (source: google.protobuf.Value). */
  variables: string;

  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model CreateTemplateRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  type: string;

  subject: string;

  @Onkt.required
  body: string;

  /** Arbitrary JSON template variables (source: google.protobuf.Value). */
  variables: string;

  is_active: boolean;
}

model CreateTemplateResponse {
  @encode(string)
  template_id: int64;

  message: string;
}

model DeleteTemplateRequest {
  @encode(string)
  id: int64;
}

model DeleteTemplateResponse {
  message: string;
}

model ListTemplatesRequest {}

model ListTemplatesResponse {
  templates: Template[];
}

model UpdateTemplateBody {
  @Onkt.required
  name: string;

  type: string;
  subject: string;
  body: string;

  /** Arbitrary JSON template variables (source: google.protobuf.Value). */
  variables: string;

  is_active: boolean;
}

model UpdateTemplateRequest {
  @encode(string)
  id: int64;

  @body
  body: UpdateTemplateBody;
}

model UpdateTemplateResponse {
  message: string;
}
