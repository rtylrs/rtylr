import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.role.v1;

/** PermissionPayload is a single grantable permission on a role. */
model PermissionPayload {
  /** Machine-readable permission identifier (e.g. "hr.employees.create"). */
  key: string;

  /** Human-readable label shown in the UI. */
  label: string;

  /** Whether this permission is currently granted. */
  is_granted: boolean;
}

/** RolePayload is an HR role with its attached permission set. */
model RolePayload {
  /** Unique identifier for the role. */
  @encode(string)
  id: int64;

  /** Tenant owner this role belongs to. */
  @encode(string)
  owner_id: int64;

  /** Display name of the role. */
  name: string;

  /** Whether this role is the tenant default assigned to new employees. */
  is_default: boolean;

  /** Full set of permissions attached to this role. */
  permissions: PermissionPayload[];

  /** Unix timestamp when this role was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this role was last updated. */
  @encode(string)
  updated_at: int64;
}

/** ListRolesResponse lists all roles for the tenant. */
model ListRolesResponse {
  /** All roles for the tenant. */
  roles: RolePayload[];
}

model CreateRoleRequest {
  /** Display name of the new role. */
  @Onkt.required
  name: string;

  /** Permissions to attach to the new role. */
  @minItems(1)
  permissions: PermissionPayload[];
}

model CreateRoleResponse {
  /** The newly created role. */
  role: RolePayload;
}

model UpdateRoleBody {
  /** New name; absent to keep current. */
  name?: string;

  /** New permission set; if provided, replaces all existing permissions. */
  @minItems(1)
  permissions: PermissionPayload[];
}

model UpdateRoleRequest {
  /** ID of the role to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable role fields bound from the HTTP payload. */
  @body
  body: UpdateRoleBody;
}

model UpdateRoleResponse {
  /** The updated role. */
  role: RolePayload;
}

model AssignRoleRequest {
  /** ID of the employee to assign the role to. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;

  /** ID of the role to assign. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  role_id?: int64;
}

model AssignRoleResponse {
  /** Whether the role was assigned successfully. */
  success: boolean;
}
