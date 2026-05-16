//! LAN /24 subnet scanner for POS hardware (printers, KDS, kiosk displays).

use serde::Serialize;
use std::net::{SocketAddr, TcpStream};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;

use crate::validation::net::validate_subnet_prefix;

#[derive(Debug, Serialize)]
pub struct DiscoveredDevice {
    pub host: String,
    pub port: u16,
    pub device_type: String,
    pub label: String,
}

/// Well-known ports probed during subnet scan, mapped to device types.
///
/// Covers ESC/POS thermal printers, Star Micronics native ports, Zebra ZPL
/// label printers, customer-facing kiosk displays, LPR/IPP network printers,
/// kitchen display systems, and Epson ePOS-Print over HTTP(S).
pub const SCAN_PORTS: &[(u16, &str, &str)] = &[
    (9100, "thermal_escpos",  "Thermal/ESC-POS or raw receipt printer"),
    (9101, "star_micronics",  "Star Micronics receipt printer (port 2)"),
    (9102, "star_micronics",  "Star Micronics receipt printer (port 3)"),
    (4242, "star_cloudprnt",  "Star CloudPRNT kiosk/receipt printer"),
    (6101, "zpl_label",       "Zebra ZPL label printer"),
    (9999, "kiosk_display",   "Customer-facing kiosk or pole display"),
    (515,  "network_lpr",     "LPR/LPD line printer"),
    (631,  "network_ipp",     "IPP network printer"),
    (9200, "kds",             "Kitchen Display System (alt port)"),
    (4999, "kds",             "Kitchen Display System"),
    (8008, "epson_epos",      "Epson ePOS-Print thermal printer (HTTP)"),
    (8043, "epson_epos",      "Epson ePOS-Print thermal printer (HTTPS)"),
];

pub fn scan_host(host: &str, timeout_ms: u64) -> Vec<DiscoveredDevice> {
    let timeout = Duration::from_millis(timeout_ms.clamp(50, 5000));
    let mut found = Vec::new();
    for &(port, device_type, label) in SCAN_PORTS {
        let addr_str = format!("{}:{}", host, port);
        if let Ok(addr) = addr_str.parse::<SocketAddr>() {
            if TcpStream::connect_timeout(&addr, timeout).is_ok() {
                found.push(DiscoveredDevice {
                    host: host.to_string(),
                    port,
                    device_type: device_type.to_string(),
                    label: format!("{} at {}:{}", label, host, port),
                });
            }
        }
    }
    found
}

/// Scan a /24 subnet in parallel. Wall time ≈ `n_ports × timeout_ms`
/// regardless of subnet size.
pub fn discover(subnet_prefix: &str, timeout_ms: u64) -> Result<Vec<DiscoveredDevice>, String> {
    let prefix = validate_subnet_prefix(subnet_prefix)?;

    let results: Arc<Mutex<Vec<DiscoveredDevice>>> = Arc::new(Mutex::new(Vec::new()));

    let handles: Vec<_> = (1u8..=254)
        .map(|octet| {
            let host = format!("{}.{}", prefix, octet);
            let results = Arc::clone(&results);
            thread::spawn(move || {
                let found = scan_host(&host, timeout_ms);
                if !found.is_empty() {
                    results.lock().expect("mutex").extend(found);
                }
            })
        })
        .collect();

    for handle in handles {
        handle.join().map_err(|_| "Thread panicked during subnet scan".to_string())?;
    }

    let mut final_results = Arc::try_unwrap(results)
        .map_err(|_| "Internal scan error: arc still held".to_string())?
        .into_inner()
        .map_err(|_| "Mutex poisoned".to_string())?;

    final_results.sort_by(|a, b| {
        a.host
            .parse::<std::net::IpAddr>()
            .ok()
            .cmp(&b.host.parse::<std::net::IpAddr>().ok())
            .then(a.port.cmp(&b.port))
    });

    Ok(final_results)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn discover_rejects_empty_subnet_prefix() {
        let result = discover("", 100);
        assert!(result.is_err());
    }

    #[test]
    fn discover_rejects_invalid_subnet_format() {
        assert!(discover("not-a-subnet", 100).is_err());
    }

    #[test]
    fn discover_rejects_four_octet_full_ip() {
        assert!(discover("192.168.1.1", 100).is_err());
    }

    #[test]
    fn scan_host_returns_nothing_for_unreachable_host() {
        // TEST-NET (192.0.2.0/24) is guaranteed non-routable
        let found = scan_host("192.0.2.1", 50);
        assert!(found.is_empty());
    }
}
