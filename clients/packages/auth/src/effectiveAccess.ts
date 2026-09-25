import type {
  AssignmentSource,
  AssignmentStatus,
  AuthorizationAppId,
  DashRole,
  EffectiveAccess,
  EffectiveAppAccess,
  LicenseAssignment,
  PermissionKey,
  RoleScope,
  RoleScopeType,
  TenantAppId,
} from "@voxire/types";

type UnknownRecord = Record<string, unknown>;

function isRecord(value: unknown): value is UnknownRecord {
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
  // OneKit wire int64s arrive as strings to avoid JS float precision loss
  // (see the client's int64-handling convention); accept that shape too.
  if (
    typeof value === "string" &&
    value.length > 0 &&
    Number.isFinite(Number(value))
  ) {
    return Number(value);
  }
  return null;
}

function stringArray(value: unknown): string[] | null {
  if (!Array.isArray(value) || !value.every((item) => typeof item === "string"))
    return null;
  return value;
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
  return value === "explicit" ||
    value === "dash_default" ||
    value === "group" ||
    value === "system"
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

function scopeType(value: unknown): RoleScopeType | null {
  const types: readonly RoleScopeType[] = [
    "business",
    "location",
    "department",
    "team",
    "project",
    "terminal",
    "portfolio",
  ];
  return typeof value === "string" && types.includes(value as RoleScopeType)
    ? (value as RoleScopeType)
    : null;
}

function parseScopes(value: unknown): RoleScope[] | null {
  if (!Array.isArray(value)) return null;
  const scopes: RoleScope[] = [];
  for (const item of value) {
    if (!isRecord(item)) return null;
    const type = scopeType(item.scopeType ?? item.scope_type);
    const id = stringValue(item.scopeId ?? item.scope_id);
    if (!type || !id) return null;
    scopes.push({ scopeType: type, scopeId: id });
  }
  return scopes;
}

function parseLicense(value: unknown): LicenseAssignment | null {
  if (!isRecord(value)) return null;
  const app = appId(value.appId ?? value.app_id);
  const status = assignmentStatus(value.status);
  const source = assignmentSource(
    value.source ?? value.assignmentSource ?? value.assignment_source,
  );
  const grantedAt = stringValue(value.grantedAt ?? value.granted_at);
  if (!app || !status || !source || !grantedAt) return null;
  const revokedAtValue = value.revokedAt ?? value.revoked_at;
  const groupIdValue = value.groupId ?? value.group_id;
  return {
    appId: app,
    status,
    source,
    groupId:
      groupIdValue === null || groupIdValue === undefined
        ? null
        : stringValue(groupIdValue),
    grantedAt,
    revokedAt:
      revokedAtValue === null || revokedAtValue === undefined
        ? null
        : stringValue(revokedAtValue),
  };
}

function parseAppAccess(value: unknown): EffectiveAppAccess | null {
  if (!isRecord(value)) return null;
  const app = appId(value.appId ?? value.app_id);
  // licensed/hasAccess are generated as `bool ,omitempty` on the backend (see
  // EffectiveAccess.IsOwner below for the same shape), so a false value is
  // indistinguishable on the wire from an absent one. Absent must default to
  // false rather than fail the whole entry closed, or every app a caller does
  // NOT have now denies apps they DO have too, since one bad entry nulls the
  // entire parse (rtylr AppGuard false-denial: an admin/member/staff account
  // with a real, active license was shown "you don't have access" because an
  // unrelated app entry's omitted false boolean poisoned parseEffectiveAccess).
  const rawLicensed = value.licensed;
  const licensed = rawLicensed == null ? false : booleanValue(rawLicensed);
  const rawHasAccess = value.hasAccess ?? value.has_access;
  const hasAccess = rawHasAccess == null ? false : booleanValue(rawHasAccess);
  // Per-role and per-assignment details are not always populated on apps[]
  // entries. Missing is treated as "none", not as malformed data; only a
  // present-but-invalid value still fails closed.
  const rawRoleIds = value.roleIds ?? value.role_ids;
  const roleIds = rawRoleIds == null ? [] : stringArray(rawRoleIds);
  const permissions =
    value.permissions == null ? [] : stringArray(value.permissions);
  const rawScopes = value.scopes;
  const scopes = rawScopes == null ? [] : parseScopes(rawScopes);
  const sourcesValue =
    value.sources == null
      ? []
      : Array.isArray(value.sources)
        ? value.sources.map(assignmentSource)
        : null;
  const rawPermissionSources =
    value.permissionSources ?? value.permission_sources;
  const permissionSourcesValue =
    rawPermissionSources == null ? {} : rawPermissionSources;
  if (
    !app ||
    licensed === null ||
    hasAccess === null ||
    !roleIds ||
    !permissions ||
    !scopes ||
    !sourcesValue ||
    sourcesValue.some((source): source is null => source === null) ||
    !isRecord(permissionSourcesValue)
  ) {
    return null;
  }
  const permissionSources: Record<string, string[]> = {};
  for (const [permission, sourceValue] of Object.entries(
    permissionSourcesValue,
  )) {
    const sourceIds = Array.isArray(sourceValue)
      ? stringArray(sourceValue)
      : typeof sourceValue === "string"
        ? sourceValue.length === 0
          ? []
          : sourceValue.split(",")
        : null;
    if (!sourceIds) return null;
    permissionSources[permission] = sourceIds;
  }
  const primaryRoleIdValue = value.primaryRoleId ?? value.primary_role_id;
  const displayRoleValue = value.displayRole ?? value.display_role;
  return {
    appId: app,
    licensed,
    hasAccess,
    roleIds,
    primaryRoleId:
      primaryRoleIdValue === null ? null : stringValue(primaryRoleIdValue),
    displayRole:
      displayRoleValue === null ? null : stringValue(displayRoleValue),
    permissions: permissions as PermissionKey[],
    scopes,
    sources: sourcesValue as AssignmentSource[],
    permissionSources,
  };
}

/** Parses an untrusted profile field and returns null instead of weakening access. */
export function parseEffectiveAccess(value: unknown): EffectiveAccess | null {
  if (!isRecord(value)) return null;
  const businessId = stringValue(value.businessId ?? value.business_id);
  const dashRoleValue = value.dashRole ?? value.dash_role;
  const dashRole = stringValue(dashRoleValue) as DashRole | string | null;
  const dashRoleIdValue = value.dashRoleId ?? value.dash_role_id;
  // dashPermissions, subscribedApps, licenses and apps are all `,omitempty` on
  // the backend, so an empty collection for this caller (e.g. a plain Flow
  // contributor has zero hub-level permissions) arrives as an absent field,
  // not an empty array. Absent defaults to empty rather than failing the
  // whole parse - see the isOwner/isAdmin comment below for why "absent" and
  // "malformed" must stay distinguishable here.
  const rawDashPermissions = value.dashPermissions ?? value.dash_permissions;
  const dashPermissions =
    rawDashPermissions == null ? [] : stringArray(rawDashPermissions);
  const rawSubscribedApps = value.subscribedApps ?? value.subscribed_apps;
  const subscribedAppsValue: unknown[] | null =
    rawSubscribedApps == null
      ? []
      : Array.isArray(rawSubscribedApps)
        ? rawSubscribedApps
        : null;
  const subscribedApps = subscribedAppsValue
    ? subscribedAppsValue.map(appId)
    : null;
  const rawLicenses = value.licenses;
  const licensesValue: (LicenseAssignment | null)[] | null =
    rawLicenses == null
      ? []
      : Array.isArray(rawLicenses)
        ? rawLicenses.map((item: unknown) => parseLicense(item))
        : null;
  const rawApps = value.apps;
  const appsValue: (EffectiveAppAccess | null)[] | null =
    rawApps == null
      ? []
      : Array.isArray(rawApps)
        ? rawApps.map((item: unknown) => parseAppAccess(item))
        : null;
  // isOwner/isAdmin are generated as `bool ,omitempty` on the backend, so a
  // false value is indistinguishable on the wire from an absent one - any
  // non-owner, non-admin caller (the common case: an admin or plain member
  // account) omits both fields entirely rather than sending them as false.
  // Treating "absent" the same as "present but invalid" here used to null out
  // the whole parse for exactly those callers, which made hasEffectiveAppAccess
  // fail closed for every app regardless of their real license - an admin or
  // member account with a genuinely active app license saw "you don't have
  // access to this app" every time. Only a present-but-non-boolean value
  // should still fail closed.
  const rawIsOwner = value.isOwner ?? value.is_owner;
  const isOwner = rawIsOwner == null ? false : booleanValue(rawIsOwner);
  const rawIsAdmin = value.isAdmin ?? value.is_admin;
  const isAdmin = rawIsAdmin == null ? false : booleanValue(rawIsAdmin);
  const rawAuthorizationVersion =
    value.authorizationVersion ?? value.authorization_version;
  const authorizationVersion =
    rawAuthorizationVersion == null ? 0 : numberValue(rawAuthorizationVersion);
  if (
    !businessId ||
    !dashRole ||
    !dashPermissions ||
    !subscribedApps ||
    subscribedApps.some((app): app is null => app === null) ||
    !licensesValue ||
    licensesValue.some((license): license is null => license === null) ||
    !appsValue ||
    appsValue.some((app): app is null => app === null) ||
    isOwner === null ||
    isAdmin === null ||
    authorizationVersion === null
  ) {
    return null;
  }
  return {
    businessId,
    dashRole,
    dashRoleId: dashRoleIdValue === null ? null : stringValue(dashRoleIdValue),
    dashPermissions: dashPermissions as PermissionKey[],
    subscribedApps: subscribedApps as TenantAppId[],
    licenses: licensesValue as LicenseAssignment[],
    apps: appsValue as EffectiveAppAccess[],
    isOwner,
    isAdmin,
    authorizationVersion,
  };
}

export function hasEffectiveAppAccess(
  access: EffectiveAccess | null | undefined,
  appId: AuthorizationAppId | string,
): boolean {
  if (!access) return false;
  if (appId === "dash") return access.dashRole !== null;
  return access.apps.some(
    (app) => app.appId === appId && app.licensed && app.hasAccess,
  );
}
