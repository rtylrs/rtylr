import type {
  AssignmentSource,
  AssignmentStatus,
  EffectiveAccess,
  EffectiveAppAccess,
  LicenseAssignment,
  PermissionKey,
  PermissionScopeType,
  PermissionScope,
  TenantAppId,
} from "@voxire/types";

type RecordValue = Record<string, unknown>;

function isRecord(value: unknown): value is RecordValue {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function stringValue(value: unknown): string | null {
  return typeof value === "string" && value.length > 0 ? value : null;
}

function booleanValue(value: unknown): boolean | null {
  return typeof value === "boolean" ? value : null;
}

function numberValue(value: unknown): number | null {
  if (typeof value === "number") return Number.isFinite(value) ? value : null;
  if (typeof value !== "string" || value.trim() === "") return null;
  const number = Number(value);
  return Number.isFinite(number) ? number : null;
}

function stringArray(value: unknown): string[] | null {
  return Array.isArray(value) && value.every((item) => typeof item === "string")
    ? value
    : null;
}

function optionalStringArray(value: unknown): string[] | null {
  return value == null ? [] : stringArray(value);
}

function optionalBoolean(value: unknown): boolean | null {
  return value == null ? false : booleanValue(value);
}

function appId(value: unknown): TenantAppId | null {
  const apps: readonly TenantAppId[] = [
    "pos",
    "erp",
    "hr",
    "crm",
    "finance",
    "flow",
    "insights",
    "menu",
    "recruit",
  ];
  return typeof value === "string" && apps.includes(value as TenantAppId)
    ? (value as TenantAppId)
    : null;
}

function assignmentSource(value: unknown): AssignmentSource | null {
  return value === "explicit" || value === "group" || value === "system"
    ? value
    : null;
}

function assignmentStatus(value: unknown): AssignmentStatus | null {
  return value === "pending" ||
    value === "active" ||
    value === "revoked" ||
    value === "archived"
    ? value
    : null;
}

function parseScope(value: unknown): PermissionScope | null {
  if (!isRecord(value)) return null;
  const scopeType = value.scopeType ?? value.scope_type;
  const scopeId = stringValue(value.scopeId ?? value.scope_id);
  const types: readonly PermissionScopeType[] = [
    "pos_location",
    "flow_project",
  ];
  if (
    typeof scopeType !== "string" ||
    !types.includes(scopeType as PermissionScopeType) ||
    !scopeId
  ) {
    return null;
  }
  return { scopeType: scopeType as PermissionScopeType, scopeId };
}

function parseScopes(value: unknown): PermissionScope[] | null {
  // Protobuf JSON omits an empty repeated field for unrestricted grants.
  if (value == null) return [];
  if (!Array.isArray(value)) return null;
  const scopes = value.map(parseScope);
  return scopes.some((scope) => scope === null)
    ? null
    : (scopes as PermissionScope[]);
}

function parseLicense(value: unknown): LicenseAssignment | null {
  if (!isRecord(value)) return null;
  const appIdValue = appId(value.appId ?? value.app_id);
  const status = assignmentStatus(value.status);
  const source = assignmentSource(
    value.source ?? value.assignmentSource ?? value.assignment_source,
  );
  const grantedAt = stringValue(value.grantedAt ?? value.granted_at);
  if (!appIdValue || !status || !source || !grantedAt) return null;
  const revokedAt = value.revokedAt ?? value.revoked_at;
  const groupId = value.groupId ?? value.group_id;
  return {
    appId: appIdValue,
    status,
    source,
    groupId: groupId == null ? null : stringValue(groupId),
    grantedAt,
    revokedAt: revokedAt == null ? null : stringValue(revokedAt),
  };
}

function parsePermissionScopes(
  value: unknown,
): Record<string, PermissionScope[]> | null {
  if (value == null) return {};
  if (!Array.isArray(value)) return null;
  const result: Record<string, PermissionScope[]> = {};
  for (const entry of value) {
    if (!isRecord(entry)) return null;
    const permission = stringValue(entry.permissionKey ?? entry.permission_key);
    const scopes = parseScopes(entry.scopes);
    if (
      !permission ||
      !scopes ||
      Object.prototype.hasOwnProperty.call(result, permission)
    )
      return null;
    result[permission] = scopes;
  }
  return result;
}

function parseAppAccess(value: unknown): EffectiveAppAccess | null {
  if (!isRecord(value)) return null;
  const app = appId(value.appId ?? value.app_id);
  const licensed = optionalBoolean(value.licensed);
  const hasAccess = optionalBoolean(value.hasAccess ?? value.has_access);
  const permissions = optionalStringArray(value.permissions);
  const permissionScopes = parsePermissionScopes(
    value.permissionScopes ?? value.permission_scopes,
  );
  const rawSources = value.sources;
  const sources =
    rawSources == null
      ? []
      : Array.isArray(rawSources)
        ? rawSources.map(assignmentSource)
        : null;
  if (
    !app ||
    licensed === null ||
    hasAccess === null ||
    !permissions ||
    !permissionScopes ||
    !sources ||
    sources.some((source) => source === null) ||
    permissions.some(
      (permission) =>
        !Object.prototype.hasOwnProperty.call(permissionScopes, permission),
    )
  ) {
    return null;
  }
  return {
    appId: app,
    licensed,
    hasAccess,
    permissions: permissions as PermissionKey[],
    permissionScopes,
    sources: sources as AssignmentSource[],
  };
}

/** Parses the server-resolved access document. Invalid or missing data denies. */
export function parseEffectiveAccess(value: unknown): EffectiveAccess | null {
  if (!isRecord(value)) return null;
  const businessId = stringValue(value.businessId ?? value.business_id);
  const rawDashPermissions = value.dashPermissions ?? value.dash_permissions;
  const dashPermissions = optionalStringArray(rawDashPermissions);
  const rawSubscribedApps = value.subscribedApps ?? value.subscribed_apps;
  const subscribedApps =
    rawSubscribedApps == null
      ? []
      : Array.isArray(rawSubscribedApps)
        ? rawSubscribedApps.map(appId)
        : null;
  const licenses =
    value.licenses == null
      ? []
      : Array.isArray(value.licenses)
        ? value.licenses.map(parseLicense)
        : null;
  const apps =
    value.apps == null
      ? []
      : Array.isArray(value.apps)
        ? value.apps.map(parseAppAccess)
        : null;
  const isOwner = optionalBoolean(value.isOwner ?? value.is_owner);
  const isAdmin = optionalBoolean(value.isAdmin ?? value.is_admin);
  const authorizationVersion = numberValue(
    value.authorizationVersion ?? value.authorization_version,
  );
  if (
    !businessId ||
    !dashPermissions ||
    !subscribedApps ||
    subscribedApps.some((app) => app === null) ||
    !licenses ||
    licenses.some((license) => license === null) ||
    !apps ||
    apps.some((entry) => entry === null) ||
    isOwner === null ||
    isAdmin === null ||
    authorizationVersion === null
  ) {
    return null;
  }
  return {
    businessId,
    dashPermissions: dashPermissions as PermissionKey[],
    subscribedApps: subscribedApps as TenantAppId[],
    licenses: licenses as LicenseAssignment[],
    apps: apps as EffectiveAppAccess[],
    isOwner,
    isAdmin,
    authorizationVersion,
  };
}

export function hasEffectiveAppAccess(
  access: EffectiveAccess | null | undefined,
  appId: "dash" | TenantAppId | string,
): boolean {
  if (!access) return false;
  if (access.isOwner || access.isAdmin) {
    return (
      appId === "dash" ||
      access.subscribedApps.some((subscribed) => subscribed === appId)
    );
  }
  if (appId === "dash") return access.dashPermissions.length > 0;
  return access.apps.some(
    (app) => app.appId === appId && app.licensed && app.hasAccess,
  );
}
