//! ESC/POS byte constants and small helpers.
//!
//! Centralizes every ESC/POS escape sequence used by receipt, drawer,
//! and health-probe code paths. Anything that writes raw thermal-printer
//! bytes should pull constants from here, never inline them.

pub const INIT: &[u8] = &[0x1B, 0x40];
pub const CUT: &[u8] = &[0x1D, 0x56, 0x41, 0x03]; // partial cut
pub const FEED_4: &[u8] = &[0x1B, 0x64, 0x04]; // feed 4 lines
pub const LF: &[u8] = &[0x0A];

pub const BOLD_ON: &[u8] = &[0x1B, 0x45, 0x01];
pub const BOLD_OFF: &[u8] = &[0x1B, 0x45, 0x00];

pub const ALIGN_LEFT: &[u8] = &[0x1B, 0x61, 0x00];
pub const ALIGN_CENTER: &[u8] = &[0x1B, 0x61, 0x01];
#[allow(dead_code)]
pub const ALIGN_RIGHT: &[u8] = &[0x1B, 0x61, 0x02];

pub const SIZE_DOUBLE: &[u8] = &[0x1B, 0x21, 0x30];
pub const SIZE_NORMAL: &[u8] = &[0x1B, 0x21, 0x00];

pub const DRAWER_PIN2: &[u8] = &[0x1B, 0x70, 0x00, 0x19, 0xFA];
pub const DRAWER_PIN5: &[u8] = &[0x1B, 0x70, 0x01, 0x19, 0xFA];

/// ESC/POS DLE EOT 1 - real-time status transmission request.
/// A conforming printer responds with a 1-byte status packet within ~100 ms.
pub const STATUS_REQUEST: &[u8] = &[0x10, 0x04, 0x01];

/// Form feed: signals "end of ticket" to most network KDS displays.
pub const FORM_FEED: u8 = 0x0C;

/// Select the drawer-kick byte sequence for a given pin (2 or 5).
pub fn drawer_bytes(pin: u8) -> &'static [u8] {
    if pin == 5 { DRAWER_PIN5 } else { DRAWER_PIN2 }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn pin2_drawer_sequence() {
        assert_eq!(drawer_bytes(2), &[0x1B, 0x70, 0x00, 0x19, 0xFA]);
    }

    #[test]
    fn pin5_drawer_sequence() {
        assert_eq!(drawer_bytes(5), &[0x1B, 0x70, 0x01, 0x19, 0xFA]);
    }

    #[test]
    fn invalid_pin_falls_back_to_pin2() {
        // Anything not exactly 5 maps to pin 2 - this is the documented behavior
        // for the higher-level validation layer (see validation::drawer).
        assert_eq!(drawer_bytes(0), DRAWER_PIN2);
        assert_eq!(drawer_bytes(99), DRAWER_PIN2);
    }
}
