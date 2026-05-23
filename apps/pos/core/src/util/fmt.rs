//! Text-row formatting used by the ESC/POS receipt builder.

pub const RECEIPT_WIDTH: usize = 42;

/// Render `label ... value` padded to `width`, followed by a newline byte.
pub fn write_row(buf: &mut Vec<u8>, label: &str, value: &str, width: usize) {
    let gap = width.saturating_sub(label.len() + value.len());
    buf.extend_from_slice(format!("{}{}{}", label, " ".repeat(gap), value).as_bytes());
    buf.push(0x0A);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn row_pads_to_width_and_ends_with_lf() {
        let mut buf = Vec::new();
        write_row(&mut buf, "A", "B", 10);
        assert_eq!(buf.len(), 11);
        assert_eq!(buf[0], b'A');
        assert_eq!(buf[9], b'B');
        assert_eq!(buf[10], 0x0A);
    }

    #[test]
    fn row_clamps_overflow_to_zero_gap() {
        let mut buf = Vec::new();
        write_row(&mut buf, "longlabel", "longvalue", 5);
        // No panic, no negative - gap is saturated to zero
        assert!(buf.ends_with(&[0x0A]));
    }
}
