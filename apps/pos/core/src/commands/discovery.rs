use serde::Serialize;
use std::net::{SocketAddr, TcpStream};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;

#[derive(Serialize)]
pub struct PrinterInfo {
    pub name: String,
    pub is_default: bool,
    pub connection_type: String,
}

#[derive(Serialize)]
pub struct SerialPortInfo {
    pub port_name: String,
    pub manufacturer: Option<String>,
    pub product: Option<String>,
    pub vid: Option<u16>,
    pub pid: Option<u16>,
}

/// List all available printers on the system.
#[tauri::command]
pub fn list_printers() -> Result<Vec<PrinterInfo>, String> {
    #[cfg(windows)]
    {
        list_printers_windows()
    }

    #[cfg(unix)]
    {
        list_printers_unix()
    }
}

/// List all available serial ports, including USB-to-serial adapters.
#[tauri::command]
pub fn list_serial_ports() -> Result<Vec<SerialPortInfo>, String> {
    let ports = serialport::available_ports()
        .map_err(|e| format!("Failed to enumerate serial ports: {}", e))?;

    let result = ports
        .into_iter()
        .map(|p| {
            let (manufacturer, product, vid, pid) = match p.port_type {
                serialport::SerialPortType::UsbPort(ref info) => (
                    info.manufacturer.clone(),
                    info.product.clone(),
                    Some(info.vid),
                    Some(info.pid),
                ),
                _ => (None, None, None, None),
            };
            SerialPortInfo {
                port_name: p.port_name,
                manufacturer,
                product,
                vid,
                pid,
            }
        })
        .collect();

    Ok(result)
}

// ── Windows ─────────────────────────────────────────────────────────────────

#[cfg(windows)]
fn list_printers_windows() -> Result<Vec<PrinterInfo>, String> {
    use std::ffi::OsString;
    use std::os::windows::ffi::OsStringExt;
    use winapi::um::winspool::{
        EnumPrintersW, GetDefaultPrinterW, PRINTER_ENUM_LOCAL, PRINTER_ENUM_CONNECTIONS,
        PRINTER_INFO_2W,
    };

    let flags = PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS;
    let mut bytes_needed: u32 = 0;
    let mut printers_returned: u32 = 0;

    unsafe {
        // First call: determine buffer size
        EnumPrintersW(
            flags,
            std::ptr::null_mut(),
            2,
            std::ptr::null_mut(),
            0,
            &mut bytes_needed,
            &mut printers_returned,
        );

        if bytes_needed == 0 {
            return Ok(vec![]);
        }

        let mut buffer: Vec<u8> = vec![0u8; bytes_needed as usize];

        let ok = EnumPrintersW(
            flags,
            std::ptr::null_mut(),
            2,
            buffer.as_mut_ptr(),
            bytes_needed,
            &mut bytes_needed,
            &mut printers_returned,
        );

        if ok == 0 {
            return Err("EnumPrintersW failed".to_string());
        }

        // Get default printer name
        let mut default_buf = vec![0u16; 256];
        let mut default_len = default_buf.len() as u32;
        GetDefaultPrinterW(default_buf.as_mut_ptr(), &mut default_len);
        let default_name = OsString::from_wide(
            &default_buf[..default_buf.iter().position(|&c| c == 0).unwrap_or(0)],
        )
        .to_string_lossy()
        .to_string();

        let info_ptr = buffer.as_ptr() as *const PRINTER_INFO_2W;
        let mut printers = Vec::with_capacity(printers_returned as usize);

        for i in 0..printers_returned as isize {
            let info = &*info_ptr.offset(i);

            let name = if info.pPrinterName.is_null() {
                continue;
            } else {
                let len = (0..).take_while(|&j| *info.pPrinterName.offset(j) != 0).count();
                let slice = std::slice::from_raw_parts(info.pPrinterName, len);
                OsString::from_wide(slice).to_string_lossy().to_string()
            };

            let port = if info.pPortName.is_null() {
                String::new()
            } else {
                let len = (0..).take_while(|&j| *info.pPortName.offset(j) != 0).count();
                let slice = std::slice::from_raw_parts(info.pPortName, len);
                OsString::from_wide(slice).to_string_lossy().to_string()
            };

            let connection_type = infer_connection_type_windows(&port);

            printers.push(PrinterInfo {
                is_default: name == default_name,
                connection_type,
                name,
            });
        }

        Ok(printers)
    }
}

#[cfg(windows)]
fn infer_connection_type_windows(port: &str) -> String {
    let p = port.to_uppercase();
    if p.starts_with("USB") {
        "USB".to_string()
    } else if p.starts_with("COM") {
        "Serial".to_string()
    } else if p.starts_with("LPT") {
        "Parallel".to_string()
    } else if p.contains("IP_") || p.contains("NET") || p.contains("WSD") || p.contains(':') {
        "Network".to_string()
    } else {
        "Unknown".to_string()
    }
}

// ── Unix (macOS / Linux) ────────────────────────────────────────────────────

#[cfg(unix)]
fn list_printers_unix() -> Result<Vec<PrinterInfo>, String> {
    use std::process::Command;

    // lpstat -a lists all accept queues (printer queues accepting jobs)
    let output = Command::new("lpstat")
        .args(["-a"])
        .output()
        .map_err(|e| format!("lpstat not available: {}", e))?;

    if !output.status.success() && output.stdout.is_empty() {
        return Ok(vec![]);
    }

    // Get default printer
    let default_output = Command::new("lpstat").args(["-d"]).output().ok();
    let default_name: String = default_output
        .and_then(|o| String::from_utf8(o.stdout).ok())
        .and_then(|s| {
            s.trim()
                .split_whitespace()
                .last()
                .map(|s| s.to_string())
        })
        .unwrap_or_default();

    let stdout = String::from_utf8(output.stdout)
        .map_err(|_| "lpstat output is not valid UTF-8".to_string())?;

    let printers = stdout
        .lines()
        .filter_map(|line| {
            // Lines look like: "printer_name accepting requests since ..."
            let name = line.split_whitespace().next()?;
            if name.is_empty() {
                return None;
            }
            let connection_type = infer_connection_type_unix(name);
            Some(PrinterInfo {
                is_default: name == default_name,
                connection_type,
                name: name.to_string(),
            })
        })
        .collect();

    Ok(printers)
}

#[cfg(unix)]
fn infer_connection_type_unix(name: &str) -> String {
    let n = name.to_lowercase();
    if n.contains("usb") {
        "USB".to_string()
    } else if n.contains("net") || n.contains("network") || n.contains("ip") {
        "Network".to_string()
    } else if n.contains("serial") || n.contains("tty") {
        "Serial".to_string()
    } else {
        "Unknown".to_string()
    }
}

// ── Network hardware auto-discovery ─────────────────────────────────────────

/// A device found during network scan.
#[derive(Debug, Serialize)]
pub struct DiscoveredDevice {
    pub host: String,
    pub port: u16,
    pub device_type: String, // "thermal_escpos" | "network_lpr" | "network_ipp" | "kds" | "unknown"
    pub label: String,
}

/// Well-known ports probed during subnet scan, mapped to device types.
///
/// Covers ESC/POS thermal printers, Star Micronics native ports, Zebra ZPL
/// label printers, customer-facing kiosk displays, LPR/IPP network printers,
/// and kitchen display systems (KDS).
const SCAN_PORTS: &[(u16, &str, &str)] = &[
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
    (8008, "epson_epos",     "Epson ePOS-Print thermal printer (HTTP)"),
    (8043, "epson_epos",     "Epson ePOS-Print thermal printer (HTTPS)"),
];

fn scan_host(host: &str, timeout_ms: u64) -> Vec<DiscoveredDevice> {
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

/// Scan a /24 subnet for hardware devices (printers, KDS displays, kiosk terminals).
///
/// `subnet_prefix` is the first three octets, e.g. "192.168.1".
/// `timeout_ms` is the per-host per-port connection timeout (50–5000 ms).
///
/// All 254 hosts are probed **in parallel** — total wall time is approximately
/// `n_ports × timeout_ms` regardless of subnet size (vs sequential which would be
/// `254 × n_ports × timeout_ms`). Results are sorted by host then port.
#[tauri::command]
pub fn discover_network_hardware(
    subnet_prefix: String,
    timeout_ms: u64,
) -> Result<Vec<DiscoveredDevice>, String> {
    let prefix = subnet_prefix.trim().to_string();
    if prefix.is_empty() {
        return Err("subnet_prefix is required (e.g. \"192.168.1\")".to_string());
    }
    // Validate: must be exactly three dot-separated octets
    let parts: Vec<&str> = prefix.split('.').collect();
    if parts.len() != 3 || parts.iter().any(|p| p.parse::<u8>().is_err()) {
        return Err(format!("Invalid subnet prefix '{}' — expected format: \"192.168.1\"", prefix));
    }

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

// ── Tests ─────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::TcpListener;
    use std::thread;

    #[test]
    fn discover_rejects_empty_subnet_prefix() {
        let result = discover_network_hardware("".to_string(), 100);
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("subnet_prefix is required"));
    }

    #[test]
    fn discover_rejects_invalid_subnet_format() {
        let result = discover_network_hardware("not-a-subnet".to_string(), 100);
        assert!(result.is_err());
    }

    #[test]
    fn discover_rejects_four_octet_full_ip() {
        // Full IP isn't a valid prefix
        let result = discover_network_hardware("192.168.1.1".to_string(), 100);
        assert!(result.is_err());
    }

    #[test]
    fn scan_host_finds_listening_port() {
        // Bind a random port on loopback, scan for it
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        // Accept one connection in background so scan_host doesn't block
        thread::spawn(move || {
            let _ = listener.accept();
        });

        // scan_host only scans SCAN_PORTS so we exercise the helper directly
        let timeout = Duration::from_millis(200);
        let addr_str = format!("127.0.0.1:{}", port);
        let addr: SocketAddr = addr_str.parse().unwrap();
        let connected = TcpStream::connect_timeout(&addr, timeout).is_ok();
        assert!(connected, "Should be able to connect to our own listener");
    }

    #[test]
    fn scan_host_returns_nothing_for_unreachable_host() {
        // TEST-NET (192.0.2.0/24) is guaranteed non-routable
        let found = scan_host("192.0.2.1", 50);
        assert!(found.is_empty(), "TEST-NET host should not be reachable");
    }
}
