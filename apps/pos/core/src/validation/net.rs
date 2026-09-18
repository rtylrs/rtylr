//! Validation for hosts, IPv4 subnets, TCP ports, and serial baud rates.

use std::net::Ipv4Addr;

/// Maximum hostname length per RFC 1035.
pub const MAX_HOST_LEN: usize = 253;

/// Whitelist of baud rates supported by every printer/KDS vendor we ship today.
pub const SUPPORTED_BAUDS: &[u32] = &[
    1200, 2400, 4800, 9600, 19_200, 38_400, 57_600, 115_200, 230_400,
];

/// Trim and validate a hostname or IPv4 literal. Returns the cleaned host.
pub fn validate_host(host: &str) -> Result<String, String> {
    let host = host.trim();
    if host.is_empty() {
        return Err("Host is required".to_string());
    }
    if host.len() > MAX_HOST_LEN {
        return Err(format!("Host is too long ({} > {} chars)", host.len(), MAX_HOST_LEN));
    }
    // Allow IPv4 literals, DNS labels, and bare hostnames. Reject any
    // characters that aren't valid in hostnames or IPv4 octets.
    let ok = host.chars().all(|c| c.is_ascii_alphanumeric() || c == '.' || c == '-' || c == '_');
    if !ok {
        return Err(format!("Host '{}' contains invalid characters", host));
    }
    Ok(host.to_string())
}

/// A non-zero TCP port. Caller's responsibility to constrain further.
pub fn validate_port(port: u16) -> Result<u16, String> {
    if port == 0 {
        return Err("Port must be non-zero".to_string());
    }
    Ok(port)
}

/// Validate a `/24` subnet prefix like "192.168.1" (no trailing octet).
pub fn validate_subnet_prefix(prefix: &str) -> Result<String, String> {
    let prefix = prefix.trim();
    if prefix.is_empty() {
        return Err("subnet_prefix is required (e.g. \"192.168.1\")".to_string());
    }
    let parts: Vec<&str> = prefix.split('.').collect();
    if parts.len() != 3 {
        return Err(format!(
            "Invalid subnet prefix '{}' - expected format: \"192.168.1\"",
            prefix
        ));
    }
    for p in &parts {
        if p.parse::<u8>().is_err() {
            return Err(format!("Invalid subnet prefix '{}' - octet '{}' is not 0-255", prefix, p));
        }
    }
    Ok(prefix.to_string())
}

/// Validate that an IPv4 literal is well-formed. Rejects multicast,
/// broadcast, and unspecified addresses for hardware targets.
#[allow(dead_code)] // exposed for upcoming commands that accept raw IPs
pub fn validate_ipv4_target(ip: &str) -> Result<Ipv4Addr, String> {
    let parsed: Ipv4Addr = ip
        .trim()
        .parse()
        .map_err(|_| format!("Invalid IPv4 address '{}'", ip))?;
    if parsed.is_unspecified() {
        return Err("0.0.0.0 is not a valid hardware target".to_string());
    }
    if parsed.is_broadcast() {
        return Err("255.255.255.255 is not a valid hardware target".to_string());
    }
    if parsed.is_multicast() {
        return Err(format!("Multicast address {} is not a valid hardware target", ip));
    }
    Ok(parsed)
}

/// Baud rate - accepts 0 as "use default 9600", else must be in whitelist.
pub fn validate_baud(baud: u32) -> Result<u32, String> {
    if baud == 0 {
        return Ok(9600);
    }
    if !SUPPORTED_BAUDS.contains(&baud) {
        return Err(format!(
            "Unsupported baud rate {} (allowed: {:?})",
            baud, SUPPORTED_BAUDS
        ));
    }
    Ok(baud)
}

/// Validate a serial port path / Windows COM name.
pub fn validate_serial_port_name(port: &str) -> Result<String, String> {
    let port = port.trim();
    if port.is_empty() {
        return Err("Serial port name is required".to_string());
    }
    if port.len() > 256 {
        return Err("Serial port name is unreasonably long".to_string());
    }
    Ok(port.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn validate_host_rejects_empty() {
        assert!(validate_host("   ").is_err());
    }

    #[test]
    fn validate_host_accepts_ipv4() {
        assert_eq!(validate_host("192.168.1.10").unwrap(), "192.168.1.10");
    }

    #[test]
    fn validate_host_accepts_dns_name() {
        assert_eq!(validate_host("printer-01.local").unwrap(), "printer-01.local");
    }

    #[test]
    fn validate_host_rejects_special_chars() {
        assert!(validate_host("evil;rm -rf").is_err());
        assert!(validate_host("a b").is_err());
    }

    #[test]
    fn validate_port_rejects_zero() {
        assert!(validate_port(0).is_err());
        assert_eq!(validate_port(9100).unwrap(), 9100);
    }

    #[test]
    fn validate_subnet_prefix_happy_path() {
        assert_eq!(validate_subnet_prefix("192.168.1").unwrap(), "192.168.1");
    }

    #[test]
    fn validate_subnet_prefix_rejects_full_ip() {
        assert!(validate_subnet_prefix("192.168.1.1").is_err());
    }

    #[test]
    fn validate_subnet_prefix_rejects_out_of_range_octet() {
        assert!(validate_subnet_prefix("192.999.1").is_err());
    }

    #[test]
    fn validate_subnet_prefix_rejects_garbage() {
        assert!(validate_subnet_prefix("abc.def.ghi").is_err());
        assert!(validate_subnet_prefix("").is_err());
    }

    #[test]
    fn validate_ipv4_target_rejects_unspecified_and_broadcast() {
        assert!(validate_ipv4_target("0.0.0.0").is_err());
        assert!(validate_ipv4_target("255.255.255.255").is_err());
    }

    #[test]
    fn validate_ipv4_target_rejects_multicast() {
        assert!(validate_ipv4_target("224.0.0.1").is_err());
    }

    #[test]
    fn validate_ipv4_target_accepts_normal() {
        assert!(validate_ipv4_target("192.168.1.50").is_ok());
    }

    #[test]
    fn validate_baud_defaults_zero_to_9600() {
        assert_eq!(validate_baud(0).unwrap(), 9600);
    }

    #[test]
    fn validate_baud_accepts_common_rates() {
        for &b in SUPPORTED_BAUDS {
            assert_eq!(validate_baud(b).unwrap(), b);
        }
    }

    #[test]
    fn validate_baud_rejects_garbage() {
        assert!(validate_baud(123).is_err());
        assert!(validate_baud(1_000_000).is_err());
    }

    #[test]
    fn validate_serial_port_name_rejects_empty_and_long() {
        assert!(validate_serial_port_name("   ").is_err());
        assert!(validate_serial_port_name(&"x".repeat(300)).is_err());
        assert_eq!(validate_serial_port_name("COM3").unwrap(), "COM3");
    }
}
