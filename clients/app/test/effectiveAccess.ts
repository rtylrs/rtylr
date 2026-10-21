import type {
  EffectiveAccess,
  PermissionKey,
  TenantAppId,
} from '../../packages/types/src/role-access';

interface EffectiveAccessOptions {
  appIds?: TenantAppId[];
  permissions?: PermissionKey[];
  dashPermissions?: PermissionKey[];
  isOwner?: boolean;
  isAdmin?: boolean;
  licensed?: boolean;
  hasAccess?: boolean;
}

export function createEffectiveAccess(
  appId?: TenantAppId,
  options: EffectiveAccessOptions = {}
): EffectiveAccess {
  const appIds = options.appIds ?? (appId ? [appId] : []);
  const permissions = options.permissions ?? [];
  return {
    businessId: '1',
    dashPermissions: options.dashPermissions ?? [],
    subscribedApps: appIds,
    licenses: [],
    apps: appIds.map((id) => {
      const appPermissions = id === appId ? permissions : [];
      return {
        appId: id,
        licensed: options.licensed ?? true,
        hasAccess: options.hasAccess ?? true,
        permissions: appPermissions,
        permissionScopes: Object.fromEntries(appPermissions.map((key) => [key, []])),
        sources: ['explicit'] as ['explicit'],
      } satisfies EffectiveAccess['apps'][number];
    }),
    isOwner: options.isOwner ?? false,
    isAdmin: options.isAdmin ?? false,
    authorizationVersion: 1,
  };
}
