import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.department.v1;

/** DepartmentPayload is an HR department record. */
model DepartmentPayload {
  /** Unique identifier for the department. */
  @encode(string)
  id: int64;

  /** Tenant owner this department belongs to. */
  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  /** Display name of the department. */
  name: string;

  description: string;

  @encode(string)
  manager_staff_id: int64;

  /** Unix timestamp when this department was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this department was last updated. */
  @encode(string)
  updated_at: int64;
}

model CreateDepartmentRequest {
  /** Display name of the department. */
  @Onkt.required
  name: string;

  description: string;

  @encode(string)
  manager_staff_id: int64;
}

model CreateDepartmentResponse {
  @encode(string)
  id: int64;
}

model DeleteDepartmentRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteDepartmentResponse {
  success: boolean;
}

model ListDepartmentsRequest {}

model ListDepartmentsResponse {
  departments: DepartmentPayload[];
}

model UpdateDepartmentBody {
  /** name is the updated department name. */
  @Onkt.required
  name: string;

  description: string;

  @encode(string)
  manager_staff_id: int64;
}

model UpdateDepartmentRequest {
  /** ID of the department to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable department fields bound from the HTTP payload. */
  @body
  body: UpdateDepartmentBody;
}

model UpdateDepartmentResponse {
  success: boolean;
}
