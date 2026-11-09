//! Authenticated-user predicates, mirroring `internal/platform/auth`
//! (`auth.go`, `roles.go`, `oauth_scope.go`).

use crate::authorization::{AppId, EffectiveAccess, ProvisioningState, ProvisioningStatus};

pub const OAUTH_ACCESS_READ: &str = "read";
pub const OAUTH_ACCESS_WRITE: &str = "write";

pub const OAUTH_APPS: [&str; 5] = ["flow", "crm", "finance", "erp", "pos"];

#[derive(Debug, Clone, Default, PartialEq)]
pub struct AuthUser {
    pub user_id: i64,
    pub business_id: i64,
    pub is_console_admin: bool,
    pub platform: String,
    pub platform_user_id: i64,
    pub plan_id: i64,
    pub effective_access: Option<EffectiveAccess>,
    pub provisioning_status: Option<ProvisioningStatus>,
    pub oauth_client_id: String,
    pub oauth_scopes: Vec<String>,
}

fn has_active_provisioning(user: Option<&AuthUser>) -> bool {
    let Some(user) = user else {
        return false;
    };
    match &user.provisioning_status {
        None => true,
        Some(status) => status.state == ProvisioningState::Active,
    }
}

pub fn is_business_admin(user: Option<&AuthUser>) -> bool {
    if !has_active_provisioning(user) {
        return false;
    }
    match user.and_then(|u| u.effective_access.as_ref()) {
        Some(access) => access.is_owner || access.is_admin,
        None => false,
    }
}

pub fn is_console_admin(user: Option<&AuthUser>) -> bool {
    user.is_some_and(|u| u.is_console_admin)
}

pub fn has_app_permission(user: Option<&AuthUser>, app: AppId, permission: &str) -> bool {
    if !has_active_provisioning(user) {
        return false;
    }
    let Some(access) = user.and_then(|u| u.effective_access.as_ref()) else {
        return false;
    };
    for app_access in &access.apps {
        if app_access.app_id == app
            && app_access.licensed
            && app_access.has_access
            && (access.is_owner || access.is_admin)
        {
            return true;
        }
    }
    for app_access in &access.apps {
        if app_access.app_id != app || !app_access.licensed || !app_access.has_access {
            continue;
        }
        if app_access
            .permissions
            .iter()
            .any(|granted| granted == permission)
        {
            return true;
        }
    }
    false
}

pub fn has_app_access(user: Option<&AuthUser>, app: AppId) -> bool {
    if !has_active_provisioning(user) {
        return false;
    }
    let Some(access) = user.and_then(|u| u.effective_access.as_ref()) else {
        return false;
    };
    access
        .apps
        .iter()
        .any(|a| a.app_id == app && a.licensed && a.has_access)
}

pub fn has_dashboard_permission(user: Option<&AuthUser>, permission: &str) -> bool {
    if !has_active_provisioning(user) {
        return false;
    }
    let Some(access) = user.and_then(|u| u.effective_access.as_ref()) else {
        return false;
    };
    if access.is_owner || access.is_admin {
        return true;
    }
    access
        .dash_permissions
        .iter()
        .any(|granted| granted == permission)
}

pub fn has_any_app_scope(user: Option<&AuthUser>, app: AppId) -> bool {
    if !has_active_provisioning(user) {
        return false;
    }
    let Some(access) = user.and_then(|u| u.effective_access.as_ref()) else {
        return false;
    };
    for app_access in &access.apps {
        if app_access.app_id != app || !app_access.licensed || !app_access.has_access {
            continue;
        }
        if access.is_owner || access.is_admin {
            return true;
        }
        if !app.app_requires_scope() {
            return true;
        }
        // An empty scope list is a business-wide grant; an empty
        // permission_scopes map fails closed. The Go code returns
        // `len(PermissionScopes) > 0` on presence, except that an empty
        // per-permission list counts as granted.
        return !app_access.permission_scopes.is_empty();
    }
    false
}

pub fn oauth_scope(app: &str, access: &str) -> String {
    format!("{app}:{access}")
}

pub fn parse_oauth_scopes(scope: &str) -> Vec<String> {
    let mut seen = std::collections::HashSet::new();
    let mut out = Vec::new();
    for field in scope.split_whitespace() {
        if seen.insert(field.to_string()) {
            out.push(field.to_string());
        }
    }
    out
}

pub fn is_oauth_token(user: Option<&AuthUser>) -> bool {
    user.is_some_and(|u| !u.oauth_client_id.is_empty())
}

/// Write implies read for the same app, as in Go.
pub fn has_oauth_scope(user: Option<&AuthUser>, required: &str) -> bool {
    let Some(user) = user else {
        return false;
    };
    let Some((app, access)) = required.split_once(':') else {
        return false;
    };
    user.oauth_scopes.iter().any(|granted| {
        granted == required
            || (access == OAUTH_ACCESS_READ && granted == &oauth_scope(app, OAUTH_ACCESS_WRITE))
    })
}

#[cfg(test)]
mod tests {
    use super::{
        has_any_app_scope, has_app_access, has_app_permission, has_oauth_scope, is_business_admin,
        is_console_admin, parse_oauth_scopes, AuthUser,
    };
    use crate::authorization::{AppId, EffectiveAccess, EffectiveAppAccess};

    fn access(
        is_owner: bool,
        is_admin: bool,
        apps: Vec<EffectiveAppAccess>,
    ) -> Option<EffectiveAccess> {
        Some(EffectiveAccess {
            business_id: 42,
            dash_permissions: vec![],
            apps,
            is_owner,
            is_admin,
        })
    }

    fn app(
        app_id: AppId,
        licensed: bool,
        has_access: bool,
        permissions: &[&str],
    ) -> EffectiveAppAccess {
        EffectiveAppAccess {
            app_id,
            licensed,
            has_access,
            permissions: permissions.iter().map(|s| (*s).to_string()).collect(),
            permission_scopes: std::collections::HashMap::new(),
        }
    }

    #[test]
    fn admin_predicates() {
        assert!(is_business_admin(Some(&AuthUser {
            effective_access: access(true, false, vec![]),
            ..AuthUser::default()
        })));
        assert!(is_business_admin(Some(&AuthUser {
            effective_access: access(false, true, vec![]),
            ..AuthUser::default()
        })));
        assert!(!is_business_admin(Some(&AuthUser {
            effective_access: access(false, false, vec![]),
            ..AuthUser::default()
        })));
        assert!(!is_business_admin(Some(&AuthUser::default())));
        assert!(!is_business_admin(None));
        let user = AuthUser {
            is_console_admin: true,
            effective_access: access(false, false, vec![]),
            ..AuthUser::default()
        };
        assert!(is_console_admin(Some(&user)));
        assert!(!is_business_admin(Some(&user)));
    }

    #[test]
    fn app_permission_requires_license_access_and_permission() {
        let allowed = AuthUser {
            effective_access: access(
                false,
                false,
                vec![app(AppId::Hr, true, true, &["hr:employees:view"])],
            ),
            ..AuthUser::default()
        };
        assert!(has_app_permission(
            Some(&allowed),
            AppId::Hr,
            "hr:employees:view"
        ));
        assert!(!has_app_permission(
            Some(&allowed),
            AppId::Hr,
            "hr:employees:edit"
        ));
        let no_license = AuthUser {
            effective_access: access(false, false, vec![app(AppId::Hr, false, true, &["x"])]),
            ..AuthUser::default()
        };
        assert!(!has_app_permission(Some(&no_license), AppId::Hr, "x"));
        let no_access = AuthUser {
            effective_access: access(false, false, vec![app(AppId::Hr, true, false, &["x"])]),
            ..AuthUser::default()
        };
        assert!(!has_app_access(Some(&no_access), AppId::Hr));
    }

    #[test]
    fn any_scope_semantics() {
        let scoped = AuthUser {
            effective_access: Some(EffectiveAccess {
                business_id: 42,
                dash_permissions: vec![],
                apps: vec![{
                    let mut a = app(AppId::Flow, true, true, &[]);
                    a.permission_scopes.insert(
                        "flow:tasks:view".to_string(),
                        vec![crate::authorization::Scope {
                            kind: "flow_project".to_string(),
                            id: "1".to_string(),
                        }],
                    );
                    a
                }],
                is_owner: false,
                is_admin: false,
            }),
            ..AuthUser::default()
        };
        assert!(has_any_app_scope(Some(&scoped), AppId::Flow));
        let owner = AuthUser {
            effective_access: access(true, false, vec![app(AppId::Flow, true, true, &[])]),
            ..AuthUser::default()
        };
        assert!(has_any_app_scope(Some(&owner), AppId::Flow));
        let inactive = AuthUser {
            effective_access: access(false, false, vec![app(AppId::Hr, true, false, &[])]),
            ..AuthUser::default()
        };
        assert!(!has_any_app_scope(Some(&inactive), AppId::Hr));
    }

    #[test]
    fn oauth_scopes_ported() {
        assert_eq!(
            parse_oauth_scopes("flow:read flow:write flow:read"),
            vec!["flow:read", "flow:write"]
        );
        let user = AuthUser {
            oauth_client_id: "client".to_string(),
            oauth_scopes: vec!["flow:write".to_string()],
            ..AuthUser::default()
        };
        assert!(has_oauth_scope(Some(&user), "flow:read"));
        assert!(has_oauth_scope(Some(&user), "flow:write"));
        assert!(!has_oauth_scope(Some(&user), "crm:read"));
        assert!(!has_oauth_scope(None, "flow:read"));
        assert!(!has_oauth_scope(Some(&user), "flowread"));
    }
}
