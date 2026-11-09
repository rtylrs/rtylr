//! Limit/offset and page helpers, mirroring the module pagination
//! normalizers (attendance `50`/`100`, insights `10`/`100`).

/// A normalized `(limit, offset)` page request.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct LimitOffset {
    pub limit: i32,
    pub offset: i32,
}

#[derive(Debug, thiserror::Error, PartialEq)]
pub enum PaginationError {
    #[error("limit must be non-negative")]
    NegativeLimit,
    #[error("offset must be non-negative")]
    NegativeOffset,
    #[error("limit must be at most {0}")]
    LimitTooLarge(i32),
    #[error("limit must be an integer")]
    LimitNotInteger,
    #[error("offset must be an integer")]
    OffsetNotInteger,
}

/// `limit == 0` falls back to `default_limit`; negatives and values above
/// `max_limit` are rejected, matching the attendance middleware.
pub fn normalize_limit_offset(
    limit: i32,
    offset: i32,
    default_limit: i32,
    max_limit: i32,
) -> Result<LimitOffset, PaginationError> {
    if limit < 0 {
        return Err(PaginationError::NegativeLimit);
    }
    if limit > max_limit {
        return Err(PaginationError::LimitTooLarge(max_limit));
    }
    if offset < 0 {
        return Err(PaginationError::NegativeOffset);
    }
    Ok(LimitOffset {
        limit: if limit == 0 { default_limit } else { limit },
        offset,
    })
}

/// Insights-style list limit: `<= 0` → 10, capped at 100.
pub fn normalize_list_limit(limit: i32) -> i32 {
    if limit <= 0 {
        10
    } else if limit > 100 {
        100
    } else {
        limit
    }
}

pub fn normalize_page(page: i32) -> i32 {
    if page <= 0 {
        1
    } else {
        page
    }
}

/// `(page - 1) * limit` with the same overflow guard as Go.
pub fn page_offset(page: i32, limit: i32) -> i32 {
    if page > (i32::MAX - (limit - 1)) / limit {
        return i32::MAX - limit + 1;
    }
    (page - 1) * limit
}

#[cfg(test)]
mod tests {
    use super::{
        normalize_limit_offset, normalize_list_limit, normalize_page, page_offset, PaginationError,
    };

    #[test]
    fn attendance_defaults() {
        assert_eq!(
            normalize_limit_offset(0, 0, 50, 100),
            Ok(super::LimitOffset {
                limit: 50,
                offset: 0
            })
        );
        assert_eq!(
            normalize_limit_offset(-1, 0, 50, 100),
            Err(PaginationError::NegativeLimit)
        );
        assert_eq!(
            normalize_limit_offset(101, 0, 50, 100),
            Err(PaginationError::LimitTooLarge(100))
        );
        assert_eq!(
            normalize_limit_offset(50, -1, 50, 100),
            Err(PaginationError::NegativeOffset)
        );
    }

    #[test]
    fn insights_page_helpers() {
        assert_eq!(normalize_list_limit(0), 10);
        assert_eq!(normalize_list_limit(250), 100);
        assert_eq!(normalize_page(0), 1);
        assert_eq!(page_offset(3, 25), 50);
    }
}
