//! Trusted browser origins, mirroring `internal/platform/origins`.
//!
//! Production patterns trust exactly one subdomain label; development
//! adds `localhost` / `127.0.0.1` on any port. Hand-rolled (no regex crate)
//! to match the Go regexes exactly: apex `voxire.tech` is NOT trusted.

pub const PRODUCTION_PATTERN_STRINGS: [&str; 9] = [
    "https://voxire.com",
    "https://*.voxire.com",
    "https://voxelect.com",
    "https://*.voxelect.com",
    "https://*.voxire.tech",
    "https://rtylr.com",
    "https://*.rtylr.com",
    "https://voxhires.com",
    "https://*.voxhires.com",
];

pub const DEVELOPMENT_PATTERN_STRINGS: [&str; 4] = [
    "http://localhost",
    "http://localhost:*",
    "http://127.0.0.1",
    "http://127.0.0.1:*",
];

fn split_origin(origin: &str) -> Option<(&str, &str)> {
    let (scheme, host) = origin.split_once("://")?;
    if host.contains('/') || host.contains('@') {
        return None;
    }
    let host = host.split(':').next().unwrap_or(host);
    Some((scheme, host))
}

fn is_single_label(host: &str) -> bool {
    !host.is_empty()
        && !host.contains('.')
        && host
            .chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '-')
}

fn matches_production(origin: &str) -> bool {
    let Some((scheme, host)) = split_origin(origin) else {
        return false;
    };
    if scheme != "https" {
        return false;
    }
    // Apex allows (https://voxire.com et al.), checked before suffixes so a
    // bare two-label host never counts as a subdomain of itself.
    if ["voxire.com", "voxelect.com", "rtylr.com", "voxhires.com"].contains(&host) {
        return true;
    }
    // One-label subdomains of the same families, plus voxire.tech (no apex).
    for suffix in [
        ".voxire.com",
        ".voxelect.com",
        ".voxire.tech",
        ".rtylr.com",
        ".voxhires.com",
    ] {
        if let Some(label) = host.strip_suffix(suffix) {
            if is_single_label(label) {
                return true;
            }
        }
    }
    false
}

fn matches_development(origin: &str) -> bool {
    let Some((scheme, host)) = split_origin(origin) else {
        return false;
    };
    if scheme != "http" {
        return false;
    }
    host == "localhost" || host == "127.0.0.1"
}

pub fn is_allowed(origin: &str, is_production: bool) -> bool {
    if matches_production(origin) {
        return true;
    }
    if is_production {
        return false;
    }
    matches_development(origin)
}

pub fn allowed_pattern_strings(is_production: bool) -> Vec<String> {
    let mut out: Vec<String> = PRODUCTION_PATTERN_STRINGS
        .iter()
        .map(|s| (*s).to_string())
        .collect();
    if !is_production {
        out.extend(DEVELOPMENT_PATTERN_STRINGS.iter().map(|s| (*s).to_string()));
    }
    out
}

#[cfg(test)]
mod tests {
    use super::is_allowed;

    #[test]
    fn origins_ported_development() {
        for (origin, want) in [
            ("http://localhost:3000", true),
            ("https://rtylr.com", true),
            ("https://dash.rtylr.com", true),
            ("https://voxire.com", true),
            ("https://api.voxire.tech", true),
            ("https://voxhires.com", true),
            ("https://careers.voxhires.com", true),
            ("http://rtylr.com", false),
            ("http://dash.rtylr.com", false),
            ("http://voxire.com", false),
            ("http://api.voxire.tech", false),
            ("http://voxhires.com", false),
        ] {
            assert_eq!(is_allowed(origin, false), want, "origin={origin}");
        }
    }

    #[test]
    fn origins_production_excludes_localhost() {
        assert!(!is_allowed("http://localhost:3000", true));
        assert!(is_allowed("https://dash.rtylr.com", true));
        assert!(is_allowed("https://api.voxire.tech", true));
        // No apex pattern for voxire.tech; deeper chains rejected.
        assert!(!is_allowed("https://voxire.tech", true));
        assert!(!is_allowed("https://a.b.rtylr.com", true));
    }
}
