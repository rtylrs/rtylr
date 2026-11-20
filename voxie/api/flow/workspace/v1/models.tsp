import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.workspace.v1;

model Workspace {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** description is the description value for this contract. */
  description: string;

  /** icon is the icon value for this contract. */
  icon: string;

  /** color is the color value for this contract. */
  color: string;

  /** is_default is the is default value for this contract. */
  is_default: boolean;

  /** created_by is the created by value for this contract. */
  @encode(string)
  created_by: int64;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  /**
   * Count of members in this workspace, for sidebar / picker badges.
   * Display projection populated on list reads.
   */
  member_count: int32;

  /** True when this is the caller's auto-created Personal workspace. */
  is_personal: boolean;
}

model WorkspaceMember {
  /** id is the unique identifier for the membership record. */
  @encode(string)
  id: int64;

  /** workspace_id is the workspace id value for this contract. */
  @encode(string)
  workspace_id: int64;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** name is the member display name. */
  name: string;

  /** email is the member email address. */
  email: string;

  /** role is the role value for this contract. */
  role: string;

  /** invited_by is the user id of the inviter, if any. */
  @encode(string)
  invited_by: int64;

  /** joined_at is when the member joined the workspace. */
  joined_at: string;
}

model ListWorkspacesRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListWorkspacesResponse {
  /** workspaces is the workspaces value for this contract. */
  workspaces: Workspace[];

  pagination: PaginationResponse;
}

model CreateWorkspaceRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** icon is the icon value for this contract. */
  @Onkt.required
  icon: string;

  /** color is the color value for this contract. */
  @Onkt.required
  color: string;
}

model CreateWorkspaceResponse {
  /** workspace is the workspace value for this contract. */
  workspace: Workspace;
}

model UpdateWorkspaceBody {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** icon is the icon value for this contract. */
  @Onkt.required
  icon: string;

  /** color is the color value for this contract. */
  @Onkt.required
  color: string;
}

model UpdateWorkspaceRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable workspace fields bound from the HTTP payload. */
  @body
  body: UpdateWorkspaceBody;
}

model UpdateWorkspaceResponse {
  /** message is a human-readable result message. */
  message: string;
}

model DeleteWorkspaceRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteWorkspaceResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ListWorkspaceMembersRequest {
  /** workspace_id is the workspace id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @query("workspace_id")
  @encode(string)
  workspace_id?: int64;

  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListWorkspaceMembersResponse {
  /** members is the workspace members for this contract. */
  members: WorkspaceMember[];

  pagination: PaginationResponse;
}

model AddWorkspaceMemberRequest {
  /** workspace_id is the workspace id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  workspace_id?: int64;

  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  /** role is the role value for this contract. */
  @Onkt.required
  role: string;
}

model AddWorkspaceMemberResponse {
  /** message is a human-readable result message. */
  message: string;
}

model RemoveWorkspaceMemberRequest {
  /** workspace_id is the workspace id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @query("workspace_id")
  @encode(string)
  workspace_id?: int64;

  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @query("user_id")
  @encode(string)
  user_id?: int64;
}

model RemoveWorkspaceMemberResponse {
  /** message is a human-readable result message. */
  message: string;
}

model UpdateWorkspaceMemberRoleRequest {
  /** workspace_id is the workspace id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  workspace_id?: int64;

  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  /** role is the role value for this contract. */
  @Onkt.required
  role: string;
}

model UpdateWorkspaceMemberRoleResponse {
  /** message is a human-readable result message. */
  message: string;
}
