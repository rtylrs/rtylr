import type {
  EffectiveAccess,
  PermissionKey,
  PermissionScope,
  TenantAppId,
} from "@voxire/types";

const APP_IDS: readonly TenantAppId[] = [
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

function appForPermission(permission: string): TenantAppId | null {
  const [prefix, resource] = permission.split(":");
  const app = prefix === "app" ? resource : prefix;
  return APP_IDS.includes(app as TenantAppId) ? (app as TenantAppId) : null;
}

/** Permission checks consume only the server-resolved authorization document. */
export function canEffectiveAccess(
  access: EffectiveAccess | null | undefined,
  permission: PermissionKey,
  appId?: TenantAppId,
): boolean {
  if (!access) return false;
  if (access.isOwner || access.isAdmin) {
    if (permission.startsWith("hub:")) return true;
    const privilegedApp = appId ?? appForPermission(permission);
    return privilegedApp !== null && access.subscribedApps.includes(privilegedApp);
  }
  if (permission.startsWith("hub:")) return access.dashPermissions.includes(permission);

  const resolvedAppId = appId ?? appForPermission(permission);
  if (!resolvedAppId) return false;
  const appAccess = access.apps.find((app) => app.appId === resolvedAppId);
  return Boolean(
    appAccess?.licensed && appAccess.hasAccess && appAccess.permissions.includes(permission),
  );
}

/** Checks an app permission against its own resource scopes. */
export function canEffectiveAccessInScope(
  access: EffectiveAccess | null | undefined,
  permission: PermissionKey,
  scope?: PermissionScope,
  appId?: TenantAppId,
): boolean {
  if (!canEffectiveAccess(access, permission, appId)) return false;
  if (!scope) return true;

  const resolvedAppId = appId ?? appForPermission(permission);
  const appAccess = access?.apps.find((app) => app.appId === resolvedAppId);
  if (!appAccess) return false;
  if (access?.isOwner || access?.isAdmin) return true;

  const allowedScopes = appAccess.permissionScopes[permission];
  if (!allowedScopes) return false;
  if (allowedScopes.length === 0) return true;
  return allowedScopes.some(
    (candidate) => candidate.scopeType === scope.scopeType && candidate.scopeId === scope.scopeId,
  );
}

export function canAllEffectiveAccess(
  access: EffectiveAccess | null | undefined,
  permissions: readonly PermissionKey[],
): boolean {
  return permissions.every((permission) => canEffectiveAccess(access, permission));
}

export function canAnyEffectiveAccess(
  access: EffectiveAccess | null | undefined,
  permissions: readonly PermissionKey[],
): boolean {
  return permissions.some((permission) => canEffectiveAccess(access, permission));
}

export function hasEffectiveAppAccess(
  access: EffectiveAccess | null | undefined,
  appId: "dash" | TenantAppId,
): boolean {
  if (!access) return false;
  if (access.isOwner || access.isAdmin) {
    return appId === "dash" || access.subscribedApps.includes(appId);
  }
  if (appId === "dash") return access.dashPermissions.length > 0;
  return access.apps.some(
    (app) => app.appId === appId && app.licensed && app.hasAccess,
  );
}
