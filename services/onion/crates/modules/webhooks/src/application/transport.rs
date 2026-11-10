//! Signature, destination validation, and internal-address guards.
//! Mirrors `transport.go`.

use hmac::Hmac;
use sha2::Sha256;
use std::net::{IpAddr, Ipv4Addr, Ipv6Addr};

/// Carries the HMAC a subscriber verifies. Shaped `t=<unix seconds>,v1=<hex>`
/// and signed over `"<timestamp>.<body>"` rather than the body alone, so a
/// captured request cannot be replayed indefinitely.
pub const SIGNATURE_HEADER: &str = "X-Rtylr-Signature";
/// Headers identifying what is being delivered, so a subscriber can dedupe
/// on the event id rather than on payload equality.
pub const EVENT_TYPE_HEADER: &str = "X-Rtylr-Event";
pub const EVENT_ID_HEADER: &str = "X-Rtylr-Event-Id";
pub const DELIVERY_ID_HEADER: &str = "X-Rtylr-Delivery-Id";

/// Addresses the stdlib predicates do not already cover, and which must
/// not be reachable from a tenant-supplied URL.
const INTERNAL_RANGES: [&str; 7] = [
    "0.0.0.0/8",     // "this network"
    "100.64.0.0/10", // carrier-grade NAT
    "192.0.0.0/24",  // IETF protocol assignments
    "198.18.0.0/15", // benchmarking
    "240.0.0.0/4",   // reserved
    "::/128",        // unspecified
    "2001:db8::/32", // documentation
];

/// `isInternalIP`: loopback, private, unspecified, link-local, multicast,
/// plus the extra ranges above.
pub fn is_internal_ip(ip: IpAddr) -> bool {
    match ip {
        IpAddr::V4(v4) => {
            if v4.is_loopback()
                || v4.is_private()
                || v4.is_unspecified()
                || v4.is_link_local()
                || v4.is_broadcast()
                || v4.is_documentation()
                || v4.is_multicast()
            {
                return true;
            }
            internal_v4_ranges().iter().any(|net| net.contains(&v4))
        }
        IpAddr::V6(v6) => {
            // Unique-local (fc00::/7) and link-local (fe80::/10) are
            // covered by Go's IsPrivate + IsLinkLocalUnicast; the other
            // predicates plus the extra ranges below cover the rest.
            let first = v6.segments()[0];
            if v6.is_loopback()
                || v6.is_unspecified()
                || v6.is_multicast()
                || (0xfc00..=0xfdff).contains(&first)
                || (0xfe80..=0xfebf).contains(&first)
                || matches!(v6.segments(), [0, 0, 0, 0, 0, 0xffff, _, _])
            {
                return true;
            }
            internal_v6_ranges().iter().any(|net| net.contains(&v6))
        }
    }
}

struct V4Net {
    net: u32,
    mask: u32,
}

struct V6Net {
    net: u128,
    mask: u128,
}

fn internal_v4_ranges() -> Vec<V4Net> {
    INTERNAL_RANGES
        .iter()
        .filter_map(|cidr| {
            let (addr, bits) = cidr.split_once('/')?;
            let bits: u32 = bits.parse().ok()?;
            let ip: Ipv4Addr = addr.parse().ok()?;
            let net = u32::from(ip);
            let mask = if bits == 0 {
                0
            } else {
                u32::MAX << (32 - bits)
            };
            Some(V4Net { net, mask })
        })
        .collect()
}

impl V4Net {
    fn contains(&self, ip: &Ipv4Addr) -> bool {
        (u32::from(*ip) & self.mask) == (self.net & self.mask)
    }
}

fn internal_v6_ranges() -> Vec<V6Net> {
    INTERNAL_RANGES
        .iter()
        .filter_map(|cidr| {
            let (addr, bits) = cidr.split_once('/')?;
            let bits: u32 = bits.parse().ok()?;
            let ip: Ipv6Addr = addr.parse().ok()?;
            let net = u128::from(ip);
            let mask = if bits == 0 {
                0
            } else {
                u128::MAX << (128 - bits)
            };
            Some(V6Net { net, mask })
        })
        .collect()
}

impl V6Net {
    fn contains(&self, ip: &Ipv6Addr) -> bool {
        (u128::from(*ip) & self.mask) == (self.net & self.mask)
    }
}

/// `isInternalHost`: localhost suffixes and internal IP literals. The
/// url crate returns IPv6 hosts bracketed (`[::1]`), so strip the
/// brackets before parsing — Go's `url.Hostname()` does the same.
pub fn is_internal_host(host: &str) -> bool {
    let lower = host.trim_end_matches('.').to_ascii_lowercase();
    if lower == "localhost" || lower.ends_with(".localhost") {
        return true;
    }
    let bare = lower
        .strip_prefix('[')
        .and_then(|value| value.strip_suffix(']'))
        .unwrap_or(&lower);
    if let Ok(ip) = bare.parse::<IpAddr>() {
        return is_internal_ip(ip);
    }
    false
}

/// The error text the transport's control hook produces; `isInternalAddressRefusal`
/// recognises these so a refused internal address is classified permanent.
pub const INTERNAL_REFUSAL_MARKERS: [&str; 4] = [
    "refusing to connect to internal address",
    "refusing unresolved address",
    "refusing unparseable address",
    "redirects are not followed",
];

/// `isInternalAddressRefusal`: did the transport refuse this connection as
/// internal (or a redirect)? Those are permanent failures.
pub fn is_internal_address_refusal(message: &str) -> bool {
    INTERNAL_REFUSAL_MARKERS
        .iter()
        .any(|marker| message.contains(marker))
}

/// `sign`: `t=<timestamp>,v1=<hex>` over `"<timestamp>.<body>"`.
pub fn sign(secret: &str, timestamp: i64, body: &[u8]) -> String {
    use hmac::Mac as _;
    let mut mac = Hmac::<Sha256>::new_from_slice(secret.as_bytes())
        .unwrap_or_else(|_| Hmac::<Sha256>::new_from_slice(&[]).expect("hmac accepts empty key"));
    mac.update(timestamp.to_string().as_bytes());
    mac.update(b".");
    mac.update(body);
    format!(
        "t={timestamp},v1={}",
        hex::encode(mac.finalize().into_bytes())
    )
}

/// `validateDestination`: cheap checks before anything is dialled.
pub fn validate_destination(raw: &str) -> Result<(), String> {
    let parsed =
        url::Url::parse(raw).map_err(|error| format!("registration url is not a url: {error}"))?;
    let scheme = parsed.scheme();
    if scheme != "http" && scheme != "https" {
        return Err(format!(
            "registration url scheme {scheme:?} is not http or https"
        ));
    }
    if parsed.host_str().is_none() {
        return Err("registration url has no host".into());
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::{
        is_internal_address_refusal, is_internal_host, is_internal_ip, sign, validate_destination,
    };

    #[test]
    fn internal_hosts() {
        assert!(is_internal_host("localhost"));
        assert!(is_internal_host("foo.localhost."));
        assert!(is_internal_host("10.1.2.3"));
        assert!(is_internal_host("169.254.169.254"));
        assert!(is_internal_host("100.64.0.1"));
        assert!(is_internal_host("198.18.0.1"));
        assert!(!is_internal_host("example.com"));
        assert!(!is_internal_host("8.8.8.8"));
    }

    #[test]
    fn internal_ips() {
        assert!(is_internal_ip("127.0.0.1".parse().unwrap()));
        assert!(is_internal_ip("192.168.1.1".parse().unwrap()));
        assert!(is_internal_ip("0.0.0.0".parse().unwrap()));
        assert!(is_internal_ip("240.0.0.1".parse().unwrap()));
        assert!(is_internal_ip("::1".parse().unwrap()));
        // Unique-local + link-local (Go IsPrivate/IsLinkLocalUnicast).
        assert!(is_internal_ip("fd00::1".parse().unwrap()));
        assert!(is_internal_ip("fc00::1".parse().unwrap()));
        assert!(is_internal_ip("fe80::1".parse().unwrap()));
        assert!(!is_internal_ip("1.1.1.1".parse().unwrap()));
        assert!(!is_internal_ip("2606:4700::1111".parse().unwrap()));
    }

    #[test]
    fn internal_ipv6_hosts_are_bracketed() {
        // The url crate returns `[::1]` for IPv6 hosts.
        assert!(is_internal_host("[::1]"));
        assert!(is_internal_host("[fd00::1]"));
        assert!(is_internal_host("[fe80::1]"));
        assert!(!is_internal_host("[2606:4700::1111]"));
    }

    #[test]
    fn signature_shape() {
        let value = sign("secret", 123, b"body");
        assert!(value.starts_with("t=123,v1="));
        assert_eq!(value.len(), "t=123,v1=".len() + 64);
        // Same inputs → same signature.
        assert_eq!(value, sign("secret", 123, b"body"));
        assert_ne!(value, sign("secret", 124, b"body"));
    }

    #[test]
    fn destination_checks() {
        assert!(validate_destination("https://example.com/hook").is_ok());
        assert!(validate_destination("ftp://example.com").is_err());
        assert!(validate_destination("http://").is_err());
        assert!(is_internal_address_refusal(
            "dial: refusing to connect to internal address 10.0.0.1"
        ));
        assert!(is_internal_address_refusal("redirects are not followed"));
        assert!(!is_internal_address_refusal("connection refused"));
    }
}

mod hex {
    pub fn encode(bytes: impl AsRef<[u8]>) -> String {
        bytes
            .as_ref()
            .iter()
            .map(|byte| format!("{byte:02x}"))
            .collect()
    }
}
