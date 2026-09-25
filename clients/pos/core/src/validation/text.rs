//! Receipt / KDS string sanitization.
//!
//! Thermal printers and KDS displays handle a narrow slice of bytes well
//! and choke on the rest (form-feed mid-line resets the receipt, NULs
//! truncate spooler jobs, etc.). These helpers keep guest-facing text
//! from corrupting hardware output.

/// Reject the ESC/POS / KDS control characters that would corrupt the
/// printed output or trigger commands mid-text.
pub const FORBIDDEN_CONTROL_BYTES: &[u8] = &[
    0x00,       // NUL - truncates Windows spool jobs
    0x07,       // BEL - beeps in some printers, unprintable elsewhere
    0x0B,       // VT
    0x0C,       // FF - end-of-ticket marker for KDS, cannot appear in body
    0x1B,       // ESC - starts an ESC/POS command sequence
    0x1C, 0x1D, 0x1E, 0x1F, // FS / GS / RS / US - used as ESC/POS prefixes
];

/// Validate that a string contains no control bytes that could corrupt
/// the receipt output. Returns the trimmed string on success.
pub fn validate_printable(label: &str, value: &str, max_len: usize) -> Result<String, String> {
    let trimmed = value.trim_end_matches('\n').trim_end_matches('\r');
    if trimmed.len() > max_len {
        return Err(format!(
            "{} is too long ({} > {} bytes)",
            label,
            trimmed.len(),
            max_len
        ));
    }
    for &b in trimmed.as_bytes() {
        if FORBIDDEN_CONTROL_BYTES.contains(&b) {
            return Err(format!(
                "{} contains forbidden control byte 0x{:02X}",
                label, b
            ));
        }
    }
    Ok(trimmed.to_string())
}

/// Convenience for optional fields - None passes through unchanged.
pub fn validate_optional(label: &str, value: &Option<String>, max_len: usize) -> Result<(), String> {
    if let Some(v) = value {
        validate_printable(label, v, max_len)?;
    }
    Ok(())
}

/// Strip forbidden control bytes instead of erroring. Use this on text
/// arriving from frontends we don't fully trust (e.g. customer-facing
/// inputs at a kiosk) when erroring would block checkout.
#[allow(dead_code)] // exposed for upcoming kiosk / CFD input paths
pub fn sanitize_printable(value: &str) -> String {
    value
        .chars()
        .filter(|c| {
            let b = *c as u32;
            !(b < 0x20 && b != 0x09 && b != 0x0A) // keep TAB and LF, drop other C0 controls
                && b != 0x7F // DEL
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_escape_byte() {
        assert!(validate_printable("name", "abc\x1Bdef", 100).is_err());
    }

    #[test]
    fn rejects_form_feed_in_body() {
        assert!(validate_printable("notes", "before\x0Cafter", 100).is_err());
    }

    #[test]
    fn rejects_nul() {
        assert!(validate_printable("name", "abc\0def", 100).is_err());
    }

    #[test]
    fn rejects_overlong() {
        let s = "a".repeat(101);
        assert!(validate_printable("name", &s, 100).is_err());
    }

    #[test]
    fn accepts_unicode_and_emoji() {
        assert!(validate_printable("name", "Café - €5", 100).is_ok());
    }

    #[test]
    fn trims_trailing_newlines() {
        assert_eq!(validate_printable("x", "Hello\n", 100).unwrap(), "Hello");
    }

    #[test]
    fn sanitize_strips_control_bytes_but_keeps_tab_and_newline() {
        let s = "a\x07b\tc\nd\x1Be";
        let cleaned = sanitize_printable(s);
        assert_eq!(cleaned, "ab\tc\nde");
    }
}
