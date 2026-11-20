import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.document.v1;

model Document {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  staff_id: int64;

  name: string;
  type: string;
  file_url: string;
  expiry_date: string;
  created_at: string;
}

/** EmployeeDocumentPayload is an employee document record. */
model EmployeeDocumentPayload {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  employee_id: int64;

  document_type: string;
  name: string;
  storage_mode: string;
  file_url: string;
  expiry_date?: string;

  @encode(string)
  uploaded_by: int64;

  @encode(string)
  uploaded_at: int64;

  @encode(string)
  created_at: int64;

  @encode(string)
  updated_at: int64;
}

model CreateDocumentRequest {
  @encode(string)
  staff_id: int64;

  @Onkt.required
  name: string;

  type: string;
  file_url: string;
  expiry_date: string;
}

model CreateDocumentResponse {
  @encode(string)
  id: int64;

  message: string;
}

model DeleteDocumentRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteDocumentResponse {
  message: string;
}

model ListDocumentsRequest {}

model ListDocumentsResponse {
  documents: Document[];
}
