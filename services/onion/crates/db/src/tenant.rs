//! Tenant scoping, mirroring `internal/platform/tenant/owner.go`.
//!
//! Fail-closed in both directions: an unverifiable scope is
//! `NoBusinessScope` (403), a lookup failure is `OwnerLookupFailed` (5xx).
//! The resolver is passed explicitly (no process-global singleton).

use std::future::Future;
use std::pin::Pin;
use std::sync::Arc;

use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum ScopeError {
    #[error("no business scope for request")]
    NoBusinessScope,
    #[error("business scope lookup failed")]
    OwnerLookupFailed,
}

impl ScopeError {
    pub fn lookup_failed() -> Self {
        Self::OwnerLookupFailed
    }

    pub fn is_no_business_scope(&self) -> bool {
        matches!(self, Self::NoBusinessScope)
    }
}

pub trait ScopeResolver: Send + Sync {
    fn active_business_owner(
        &self,
        business_id: i64,
    ) -> Pin<Box<dyn Future<Output = Result<i64, ScopeError>> + Send + '_>>;
    fn user_business(
        &self,
        business_id: i64,
        user_id: i64,
    ) -> Pin<Box<dyn Future<Output = Result<i64, ScopeError>> + Send + '_>>;
}

/// A tenant scope backed by an explicit resolver.
#[derive(Clone, Default)]
pub struct TenantScope {
    resolver: Option<Arc<dyn ScopeResolver>>,
}

impl TenantScope {
    pub fn new(resolver: Arc<dyn ScopeResolver>) -> Self {
        Self {
            resolver: Some(resolver),
        }
    }

    pub fn unconfigured() -> Self {
        Self { resolver: None }
    }

    /// Verify `business_id` names an active business.
    pub async fn active_business_id(&self, business_id: i64) -> Result<i64, ScopeError> {
        if business_id <= 0 {
            return Err(ScopeError::NoBusinessScope);
        }
        let Some(resolver) = self.resolver.as_ref() else {
            return Err(ScopeError::lookup_failed());
        };
        let owner_id = resolver.active_business_owner(business_id).await?;
        if owner_id <= 0 {
            return Err(ScopeError::NoBusinessScope);
        }
        Ok(business_id)
    }

    /// Preserve a verified scope; reject unverified/default tokens.
    pub async fn resolve_business_id(
        &self,
        business_id: i64,
        user_id: i64,
    ) -> Result<i64, ScopeError> {
        let Some(resolver) = self.resolver.as_ref() else {
            return Err(ScopeError::NoBusinessScope);
        };
        if user_id <= 0 {
            return Err(ScopeError::NoBusinessScope);
        }
        let resolved = resolver.user_business(business_id, user_id).await?;
        if resolved <= 0 {
            return Err(ScopeError::NoBusinessScope);
        }
        Ok(resolved)
    }
}

#[cfg(test)]
mod tests {
    use super::{ScopeError, ScopeResolver, TenantScope};

    struct Stub {
        owner_id: i64,
        business_id: i64,
        fail: bool,
    }

    impl ScopeResolver for Stub {
        fn active_business_owner(
            &self,
            _business_id: i64,
        ) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<i64, ScopeError>> + Send + '_>>
        {
            let owner_id = self.owner_id;
            let fail = self.fail;
            Box::pin(async move {
                if fail {
                    Err(ScopeError::lookup_failed())
                } else {
                    Ok(owner_id)
                }
            })
        }

        fn user_business(
            &self,
            _business_id: i64,
            _user_id: i64,
        ) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<i64, ScopeError>> + Send + '_>>
        {
            let business_id = self.business_id;
            let fail = self.fail;
            Box::pin(async move {
                if fail {
                    Err(ScopeError::lookup_failed())
                } else {
                    Ok(business_id)
                }
            })
        }
    }

    #[tokio::test]
    async fn resolve_is_fail_closed() {
        let scope = TenantScope::new(std::sync::Arc::new(Stub {
            owner_id: 7,
            business_id: 42,
            fail: false,
        }));
        assert_eq!(scope.resolve_business_id(42, 7).await, Ok(42));
        assert_eq!(
            scope.resolve_business_id(42, 0).await,
            Err(ScopeError::NoBusinessScope)
        );
        assert_eq!(
            TenantScope::unconfigured().resolve_business_id(42, 7).await,
            Err(ScopeError::NoBusinessScope)
        );
        let scope = TenantScope::new(std::sync::Arc::new(Stub {
            owner_id: 0,
            business_id: 0,
            fail: false,
        }));
        assert_eq!(
            scope.resolve_business_id(42, 7).await,
            Err(ScopeError::NoBusinessScope)
        );
    }

    #[tokio::test]
    async fn active_is_fail_closed() {
        let scope = TenantScope::new(std::sync::Arc::new(Stub {
            owner_id: 7,
            business_id: 42,
            fail: false,
        }));
        assert_eq!(
            scope.active_business_id(0).await,
            Err(ScopeError::NoBusinessScope)
        );
        assert_eq!(scope.active_business_id(42).await, Ok(42));
        assert_eq!(
            TenantScope::unconfigured().active_business_id(42).await,
            Err(ScopeError::lookup_failed())
        );
        let scope = TenantScope::new(std::sync::Arc::new(Stub {
            owner_id: 0,
            business_id: 0,
            fail: false,
        }));
        assert_eq!(
            scope.active_business_id(42).await,
            Err(ScopeError::NoBusinessScope)
        );
        let scope = TenantScope::new(std::sync::Arc::new(Stub {
            owner_id: 7,
            business_id: 42,
            fail: true,
        }));
        assert_eq!(
            scope.active_business_id(42).await,
            Err(ScopeError::lookup_failed())
        );
    }
}
