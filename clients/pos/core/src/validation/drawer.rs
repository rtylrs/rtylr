//! Cash-drawer specific validation.

/// Cash-drawer kick pin must be 2 or 5 (the two pins on RJ11/RJ12
/// drawer connectors). Any other value is rejected to surface bad
/// frontend config instead of silently falling back.
pub fn validate_drawer_pin(pin: u8) -> Result<u8, String> {
    match pin {
        2 | 5 => Ok(pin),
        _ => Err(format!("Drawer pin must be 2 or 5, got {}", pin)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn accepts_pin_2_and_5() {
        assert_eq!(validate_drawer_pin(2).unwrap(), 2);
        assert_eq!(validate_drawer_pin(5).unwrap(), 5);
    }

    #[test]
    fn rejects_other_pins() {
        assert!(validate_drawer_pin(0).is_err());
        assert!(validate_drawer_pin(1).is_err());
        assert!(validate_drawer_pin(3).is_err());
        assert!(validate_drawer_pin(255).is_err());
    }
}
