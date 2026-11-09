//! Product platform resolution, mirroring `auth/product_platform.go`.
//!
//! Headers win (`X-Product-Platform`, then `Product-Platform`), then
//! Origin, Host, absolute URL host and Referer are substring-scanned.

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ProductPlatform {
    Unknown,
    Rtylr,
    Voxire,
}

pub const PRODUCT_PLATFORM_HEADER: &str = "X-Product-Platform";

pub fn normalize_product_platform(raw: &str) -> ProductPlatform {
    match raw.trim().to_ascii_lowercase().as_str() {
        "rtylr" => ProductPlatform::Rtylr,
        "voxire" => ProductPlatform::Voxire,
        _ => ProductPlatform::Unknown,
    }
}

fn infer_product_platform(candidate: &str) -> ProductPlatform {
    let lowered = candidate.to_ascii_lowercase();
    if lowered.contains("rtylr") {
        ProductPlatform::Rtylr
    } else if lowered.contains("voxire") || lowered.contains("voxelect") {
        ProductPlatform::Voxire
    } else {
        ProductPlatform::Unknown
    }
}

pub fn resolve_product_platform(
    x_product_platform: &str,
    product_platform: &str,
    origin: &str,
    host: &str,
    url_host: &str,
    referer: &str,
) -> ProductPlatform {
    let from_header = normalize_product_platform(x_product_platform);
    if from_header != ProductPlatform::Unknown {
        return from_header;
    }
    let from_header = normalize_product_platform(product_platform);
    if from_header != ProductPlatform::Unknown {
        return from_header;
    }
    for candidate in [origin, host, url_host, referer] {
        let platform = infer_product_platform(candidate);
        if platform != ProductPlatform::Unknown {
            return platform;
        }
    }
    ProductPlatform::Unknown
}

#[cfg(test)]
mod tests {
    use super::{resolve_product_platform, ProductPlatform};

    #[test]
    fn header_wins_then_origin_fallback() {
        assert_eq!(
            resolve_product_platform("rtylr", "", "", "api.voxire.com", "", ""),
            ProductPlatform::Rtylr
        );
        assert_eq!(
            resolve_product_platform("", "", "https://app.voxire.com", "api.rtylr.com", "", ""),
            ProductPlatform::Voxire
        );
        assert_eq!(
            resolve_product_platform("", "", "", "", "", ""),
            ProductPlatform::Unknown
        );
    }
}
