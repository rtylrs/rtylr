//! Hub menu service. Mirrors `internal/application/service.go`.
//! `MenuItem` is a JSON value here (the generated contract type arrives
//! with the module's API wiring); the validation rules are the ported
//! part.

use crate::domain::{Banner, BannerError, BannerInput};
use chrono::{DateTime, Timelike, Utc};
use thiserror::Error;

pub const MAX_BANNER_TITLE_LENGTH: usize = 255;
pub const MAX_BANNER_BODY_LENGTH: usize = 2000;
pub const MAX_BANNER_URL_LENGTH: usize = 2048;

#[derive(Debug, Error, PartialEq)]
pub enum MenuError {
    #[error("invalid hub menu items")]
    InvalidMenuItems,
    #[error("{0}")]
    Banner(#[from] BannerError),
}

/// Port: banners + menu items.
pub trait Repository: Send + Sync {
    fn list_banners(&self, business_id: i64) -> Result<Vec<Banner>, MenuError>;
    fn create_banner(&self, business_id: i64, input: &BannerInput) -> Result<i64, MenuError>;
    fn update_banner(
        &self,
        business_id: i64,
        id: i64,
        input: &BannerInput,
    ) -> Result<(), MenuError>;
    fn delete_banner(&self, business_id: i64, id: i64) -> Result<(), MenuError>;
    fn get_banner(&self, business_id: i64, id: i64) -> Result<Banner, MenuError>;
    fn list_menu_items(&self, business_id: i64) -> Result<Vec<serde_json::Value>, MenuError>;
    fn update_menu_items(
        &self,
        business_id: i64,
        items: &[serde_json::Value],
    ) -> Result<(), MenuError>;
}

/// The hub menu service.
pub struct Service {
    repo: std::sync::Arc<dyn Repository>,
}

impl Service {
    pub fn new(repo: std::sync::Arc<dyn Repository>) -> Self {
        Service { repo }
    }

    pub fn list_banners(&self, business_id: i64) -> Result<Vec<Banner>, MenuError> {
        self.repo.list_banners(business_id)
    }

    pub fn list_menu_items(&self, business_id: i64) -> Result<Vec<serde_json::Value>, MenuError> {
        self.repo.list_menu_items(business_id)
    }

    /// `Update`: at least one item, none null.
    pub fn update_menu_items(
        &self,
        business_id: i64,
        items: &[serde_json::Value],
    ) -> Result<(), MenuError> {
        if items.is_empty() {
            return Err(MenuError::InvalidMenuItems);
        }
        for (index, item) in items.iter().enumerate() {
            if item.is_null() {
                return Err(MenuError::InvalidMenuItems);
            }
            let _ = index;
        }
        self.repo.update_menu_items(business_id, items)
    }

    pub fn create_banner(
        &self,
        business_id: i64,
        input: &BannerInput,
    ) -> Result<Banner, MenuError> {
        let normalized = normalize_banner_input(input.clone())?;
        let id = self.repo.create_banner(business_id, &normalized)?;
        self.repo.get_banner(business_id, id)
    }

    pub fn update_banner(
        &self,
        business_id: i64,
        id: i64,
        input: &BannerInput,
    ) -> Result<Banner, MenuError> {
        if id <= 0 {
            return Err(BannerError::NotFound.into());
        }
        let normalized = normalize_banner_input(input.clone())?;
        self.repo.update_banner(business_id, id, &normalized)?;
        self.repo.get_banner(business_id, id)
    }

    pub fn delete_banner(&self, business_id: i64, id: i64) -> Result<(), MenuError> {
        if id <= 0 {
            return Err(BannerError::NotFound.into());
        }
        self.repo.delete_banner(business_id, id)
    }
}

/// `normalizeBannerInput`: trim, length caps, and the window ordering.
pub fn normalize_banner_input(mut input: BannerInput) -> Result<BannerInput, BannerError> {
    input.title = input.title.trim().to_string();
    input.body = input.body.trim().to_string();
    input.image_url = input.image_url.trim().to_string();
    input.link_url = input.link_url.trim().to_string();
    if input.title.is_empty() {
        return Err(BannerError::InvalidInput);
    }
    if input.title.len() > MAX_BANNER_TITLE_LENGTH {
        return Err(BannerError::InvalidInput);
    }
    if input.body.len() > MAX_BANNER_BODY_LENGTH {
        return Err(BannerError::InvalidInput);
    }
    if input.image_url.len() > MAX_BANNER_URL_LENGTH {
        return Err(BannerError::InvalidInput);
    }
    if input.link_url.len() > MAX_BANNER_URL_LENGTH {
        return Err(BannerError::InvalidInput);
    }
    // Go's `IsZero()` is the zero time (year 1), not the Unix epoch.
    if !is_zero_timestamp(&input.starts_at)
        && !is_zero_timestamp(&input.ends_at)
        && input.ends_at < input.starts_at
    {
        return Err(BannerError::InvalidInput);
    }
    Ok(input)
}

/// `IsZero` for a timestamp: Go's `time.Time.IsZero()` is the zero value
/// (January 1, year 1, 00:00:00 UTC), not the Unix epoch. A zero
/// `NaiveDateTime` converts to that same instant.
pub fn is_zero_timestamp(value: &DateTime<Utc>) -> bool {
    value.timestamp() == -62_135_596_800 && value.nanosecond() == 0
}

#[cfg(test)]
mod tests {
    use super::{
        is_zero_timestamp, normalize_banner_input, BannerError, BannerInput, MenuError,
        MAX_BANNER_BODY_LENGTH, MAX_BANNER_TITLE_LENGTH, MAX_BANNER_URL_LENGTH,
    };
    use chrono::{DateTime, TimeZone, Utc};

    fn input() -> BannerInput {
        BannerInput {
            title: " Hello ".into(),
            body: " World ".into(),
            image_url: String::new(),
            link_url: String::new(),
            position: 0,
            is_active: true,
            starts_at: DateTime::<Utc>::from_timestamp(0, 0).unwrap(),
            ends_at: DateTime::<Utc>::from_timestamp(0, 0).unwrap(),
        }
    }

    #[test]
    fn trims_and_accepts() {
        let normalized = normalize_banner_input(input()).unwrap();
        assert_eq!(normalized.title, "Hello");
        assert_eq!(normalized.body, "World");
    }

    #[test]
    fn rejects_empty_title_and_oversized_fields() {
        let mut bad = input();
        bad.title = "   ".into();
        assert_eq!(normalize_banner_input(bad), Err(BannerError::InvalidInput));
        let mut bad = input();
        bad.title = "x".repeat(MAX_BANNER_TITLE_LENGTH + 1);
        assert_eq!(normalize_banner_input(bad), Err(BannerError::InvalidInput));
        let mut bad = input();
        bad.body = "x".repeat(MAX_BANNER_BODY_LENGTH + 1);
        assert_eq!(normalize_banner_input(bad), Err(BannerError::InvalidInput));
        let mut bad = input();
        bad.image_url = "x".repeat(MAX_BANNER_URL_LENGTH + 1);
        assert_eq!(normalize_banner_input(bad), Err(BannerError::InvalidInput));
        let mut bad = input();
        bad.link_url = "x".repeat(MAX_BANNER_URL_LENGTH + 1);
        assert_eq!(normalize_banner_input(bad), Err(BannerError::InvalidInput));
    }

    #[test]
    fn rejects_inverted_windows() {
        let mut bad = input();
        bad.starts_at = Utc.with_ymd_and_hms(2026, 1, 2, 0, 0, 0).unwrap();
        bad.ends_at = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(normalize_banner_input(bad), Err(BannerError::InvalidInput));
    }

    #[test]
    fn zero_window_is_allowed() {
        let value = input();
        assert!(normalize_banner_input(value).is_ok());
        // Go's zero time is year 1, not the Unix epoch.
        assert!(is_zero_timestamp(
            &DateTime::<Utc>::from_timestamp(-62_135_596_800, 0).unwrap()
        ));
        assert!(!is_zero_timestamp(
            &DateTime::<Utc>::from_timestamp(0, 0).unwrap()
        ));
    }

    #[test]
    fn menu_item_rules() {
        assert_eq!(
            MenuError::InvalidMenuItems.to_string(),
            "invalid hub menu items"
        );
    }
}
