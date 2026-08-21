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

/// Scan a /24 subnet in parallel with a bounded worker pool. Spawning 254
/// raw threads per scan starved low-spec registers; a pool sized to the CPU
/// keeps wall time reasonable without the thread explosion. Results are
/// sorted by pre-parsed IP then port (no re-parsing inside the comparator).
pub fn discover(subnet_prefix: &str, timeout_ms: u64) -> Result<Vec<DiscoveredDevice>, String> {
    let prefix = validate_subnet_prefix(subnet_prefix)?;

    let results: Arc<Mutex<Vec<DiscoveredDevice>>> = Arc::new(Mutex::new(Vec::new()));
    let hosts: Arc<Vec<String>> = Arc::new((1u8..=254).map(|octet| format!("{}.{}", prefix, octet)).collect());
    let next = Arc::new(std::sync::atomic::AtomicUsize::new(0));
    let workers = thread::available_parallelism()
        .map(|n| n.get())
        .unwrap_or(4)
        .min(hosts.len());

    let handles: Vec<_> = (0..workers)
        .map(|_| {
            let results = Arc::clone(&results);
            let hosts = Arc::clone(&hosts);
            let next = Arc::clone(&next);
            thread::spawn(move || loop {
                let index = next.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
                if index >= hosts.len() {
                    break;
                }
                let found = scan_host(&hosts[index], timeout_ms);
                if found.is_empty() {
                    continue;
                }
                // A poisoned mutex (a panicking worker) must not panic the
                // whole command; recover the guarded data and continue.
                match results.lock() {
                    Ok(mut guard) => guard.extend(found),
                    Err(poisoned) => poisoned.into_inner().extend(found),
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
        .unwrap_or_else(|poisoned| poisoned.into_inner());

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
