//! Domain types. Mirrors `internal/domain/{lead,subscriber}.go`.

/// One contact/quote form submission.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Lead {
    pub id: i64,
    pub name: String,
    pub email: String,
    pub phone: String,
    pub company: String,
    pub service: String,
    pub project: String,
    pub budget: String,
    pub source: String,
}

/// Lead create payload.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct LeadInput {
    pub name: String,
    pub email: String,
    pub phone: String,
    pub company: String,
    pub service: String,
    pub project: String,
    pub budget: String,
    pub source: String,
}

/// One newsletter signup from voxire.com.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct SubscriberInput {
    pub email: String,
    pub source: String,
    pub post_slug: String,
    pub post_title: String,
    pub post_category: String,
    pub language: String,
}

/// Form placements on voxire.com that can produce a signup. Anything else
/// is rejected at validation.
pub const ALLOWED_NEWSLETTER_SOURCES: [&str; 4] = [
    "intel_drop",
    "report_banner",
    "article_pdf_mid",
    "article_pdf_end",
];

/// The two languages the site is served in.
pub const ALLOWED_LANGUAGES: [&str; 2] = ["en", "ar"];

pub fn is_allowed_newsletter_source(source: &str) -> bool {
    ALLOWED_NEWSLETTER_SOURCES.contains(&source)
}

pub fn is_allowed_language(language: &str) -> bool {
    ALLOWED_LANGUAGES.contains(&language)
}

#[cfg(test)]
mod tests {
    use super::{is_allowed_language, is_allowed_newsletter_source};

    #[test]
    fn source_allowlist() {
        assert!(is_allowed_newsletter_source("intel_drop"));
        assert!(is_allowed_newsletter_source("article_pdf_end"));
        assert!(!is_allowed_newsletter_source("footer"));
        assert!(!is_allowed_newsletter_source(""));
    }

    #[test]
    fn language_allowlist() {
        assert!(is_allowed_language("en"));
        assert!(is_allowed_language("ar"));
        assert!(!is_allowed_language("fr"));
    }
}
