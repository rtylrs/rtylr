use std::io::{Read, Write};
use std::time::Duration;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DeviceStatus {
    pub name: String,
    pub status: String,      // "ok" | "unreachable" | "not_configured"
    pub message: String,
    pub device_type: String, // "thermal_escpos" | "serial_escpos" | "windows_spooler" | "network_lpr" | "network_ipp" | "unknown" | "none"
}

#[derive(Debug, Serialize, Deserialize)]
pub struct HardwareHealthReport {
    pub printer: DeviceStatus,
    pub drawer: DeviceStatus,
    pub kds: DeviceStatus,
}

#[derive(Debug, Deserialize)]
pub struct HealthCheckConfig {
    // Printer
    pub printer_transport: String, // "windows" | "serial" | "network" | ""
    pub printer_serial_port: Option<String>,
    pub printer_baud_rate: Option<u32>,
    pub printer_host: Option<String>,
    pub printer_port: Option<u16>,
    pub printer_name: Option<String>,
    // Drawer (if separate from printer)
    pub drawer_transport: Option<String>, // "printer" | "serial" | "network"
    pub drawer_serial_port: Option<String>,
    pub drawer_baud_rate: Option<u32>,
    pub drawer_host: Option<String>,
    pub drawer_port: Option<u16>,
    // KDS hardware
    pub kds_transport: String, // "none" | "serial" | "network"
    pub kds_serial_port: Option<String>,
    pub kds_baud_rate: Option<u32>,
    pub kds_host: Option<String>,
    pub kds_port: Option<u16>,
    // Retry count for flaky connections (0 = no retry, default 1)
    pub retry_count: Option<u32>,
}

/// Retry a probe up to `retries` extra times if the first attempt fails.
/// Returns immediately on the first `status == "ok"` result.
fn probe_with_retry<F>(probe: F, retries: u32) -> DeviceStatus
where
    F: Fn() -> DeviceStatus,
{
    let mut result = probe();
    for _ in 0..retries {
        if result.status == "ok" {
            break;
        }
        std::thread::sleep(Duration::from_millis(250));
        result = probe();
    }
    result
}

fn probe_serial(port_name: &str, baud_rate: u32) -> DeviceStatus {
    let name = port_name.to_string();
    if port_name.trim().is_empty() {
        return DeviceStatus {
            name: name.clone(),
            status: "not_configured".to_string(),
            message: "No serial port configured".to_string(),
            device_type: "none".to_string(),
        };
    }
    let baud = if baud_rate == 0 { 9600 } else { baud_rate };
    match serialport::new(port_name, baud)
        .timeout(Duration::from_millis(2000))
        .open()
    {
        Ok(_) => DeviceStatus {
            name,
            status: "ok".to_string(),
            message: format!("Serial port {} reachable at {} baud", port_name, baud),
            device_type: "serial_escpos".to_string(),
        },
        Err(e) => DeviceStatus {
            name,
            status: "unreachable".to_string(),
            message: format!("Cannot open {}: {}", port_name, e),
            device_type: "serial_escpos".to_string(),
        },
    }
}

// Zebra ZPL host identification command.
// A Zebra printer responds with STX + printer model info + ETX within ~200 ms.
const ZPL_HOST_IDENT: &[u8] = b"~HI\r\n";

/// Attempt to confirm ZPL protocol on an already-connected stream.
/// Returns `true` if the device responded with a Zebra STX header.
fn probe_zpl_protocol(stream: &mut std::net::TcpStream) -> bool {
    stream.set_write_timeout(Some(Duration::from_millis(200))).ok();
    stream.set_read_timeout(Some(Duration::from_millis(400))).ok();

    if stream.write_all(ZPL_HOST_IDENT).is_err() {
        return false;
    }

    let mut buf = [0u8; 4];
    if stream.read(&mut buf).unwrap_or(0) == 0 {
        return false;
    }
    // Zebra responses start with STX (0x02) or 'N' for model info
    buf[0] == 0x02 || buf[0] == b'N'
}

fn infer_tcp_device_type(port: u16) -> &'static str {
    match port {
        9100        => "thermal_escpos",
        9101 | 9102 => "star_micronics",
        4242        => "star_cloudprnt",
        6101        => "zpl_label",
        9999        => "kiosk_display",
        515         => "network_lpr",
        631         => "network_ipp",
        9200 | 4999 => "kds",
        8008 | 8043 => "epson_epos",
        _           => "unknown",
    }
}

/// Probe Epson ePOS-Print HTTP endpoint (port 8008/8043).
/// Epson ePOS printers expose an XML-over-HTTP API; a GET to /cgi-bin/epos/service.cgi
/// returns a 200 with an ePOS XML body when the printer is ready.
fn probe_epos_http(host: &str, port: u16) -> bool {
    let addr_str = format!("{}:{}", host, port);
    let Ok(addr) = addr_str.parse::<std::net::SocketAddr>() else {
        return false;
    };
    let Ok(mut stream) = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2)) else {
        return false;
    };
    stream.set_write_timeout(Some(Duration::from_millis(500))).ok();
    stream.set_read_timeout(Some(Duration::from_millis(800))).ok();

    let request = format!(
        "GET /cgi-bin/epos/service.cgi HTTP/1.0\r\nHost: {}:{}\r\nConnection: close\r\n\r\n",
        host, port
    );
    if stream.write_all(request.as_bytes()).is_err() {
        return false;
    }

    let mut buf = [0u8; 128];
    let n = stream.read(&mut buf).unwrap_or(0);
    if n < 7 {
        return false;
    }
    // Accept HTTP 200 or 400 — both confirm an ePOS HTTP server is present
    let resp = &buf[..n];
    resp.starts_with(b"HTTP/1") && (resp.contains(&b'2') || resp.contains(&b'4'))
}

// ESC/POS DLE EOT 1 — real-time status transmission request.
// A conforming ESC/POS printer responds with a 1-byte status packet within ~100 ms.
const ESCPOS_STATUS_REQUEST: &[u8] = &[0x10, 0x04, 0x01];

/// Attempt to confirm ESC/POS protocol on an already-connected TCP stream.
/// Returns `true` if the printer responded with a status byte.
fn probe_escpos_protocol(stream: &mut std::net::TcpStream) -> bool {
    stream
        .set_write_timeout(Some(Duration::from_millis(200)))
        .ok();
    stream
        .set_read_timeout(Some(Duration::from_millis(300)))
        .ok();

    if stream.write_all(ESCPOS_STATUS_REQUEST).is_err() {
        return false;
    }

    let mut buf = [0u8; 1];
    matches!(stream.read_exact(&mut buf), Ok(()))
}

fn probe_tcp(host: &str, port: u16) -> DeviceStatus {
    let addr_str = format!("{}:{}", host, port);
    let name = addr_str.clone();
    let device_type_hint = infer_tcp_device_type(port);
    if host.trim().is_empty() {
        return DeviceStatus {
            name,
            status: "not_configured".to_string(),
            message: "No host configured".to_string(),
            device_type: device_type_hint.to_string(),
        };
    }
    let addr = match addr_str.parse::<std::net::SocketAddr>() {
        Ok(a) => a,
        Err(e) => {
            return DeviceStatus {
                name,
                status: "unreachable".to_string(),
                message: format!("Invalid address {}: {}", addr_str, e),
                device_type: device_type_hint.to_string(),
            }
        }
    };
    match std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(3)) {
        Ok(mut stream) => {
            let (device_type, detail) = match port {
                9100 => {
                    if probe_escpos_protocol(&mut stream) {
                        ("thermal_escpos_confirmed", "ESC/POS status byte received — thermal printer confirmed")
                    } else {
                        ("thermal_escpos", "Connected on port 9100 — ESC/POS assumed (no status response)")
                    }
                }
                6101 => {
                    if probe_zpl_protocol(&mut stream) {
                        ("zpl_label_confirmed", "Zebra ZPL response received — label printer confirmed")
                    } else {
                        ("zpl_label", "Connected on port 6101 — ZPL assumed (no identification response)")
                    }
                }
                8008 | 8043 => {
                    // Drop the TCP stream first so the HTTP probe can make its own connection
                    drop(stream);
                    if probe_epos_http(host, port) {
                        ("epson_epos_confirmed", "Epson ePOS-Print HTTP response confirmed")
                    } else {
                        ("epson_epos", "Connected on ePOS port — Epson ePOS assumed (no HTTP response)")
                    }
                }
                // Star Micronics native ports and all other well-known device ports
                // are identified by port number alone (no probe needed)
                _ => (device_type_hint, "TCP connection established"),
            };
            DeviceStatus {
                name,
                status: "ok".to_string(),
                message: format!("{}:{} reachable — {}", host, port, detail),
                device_type: device_type.to_string(),
            }
        }
        Err(e) => DeviceStatus {
            name,
            status: "unreachable".to_string(),
            message: format!("Cannot connect to {}:{}: {}", host, port, e),
            device_type: device_type_hint.to_string(),
        },
    }
}

#[cfg(windows)]
fn probe_windows_printer(printer_name: &str) -> DeviceStatus {
    use std::ffi::OsStr;
    use std::os::windows::ffi::OsStrExt;
    use std::ptr;
    use winapi::um::winspool::{ClosePrinter, OpenPrinterW};

    let name = printer_name.to_string();
    if printer_name.trim().is_empty() {
        return DeviceStatus {
            name,
            status: "not_configured".to_string(),
            message: "No printer name configured".to_string(),
            device_type: "windows_spooler".to_string(),
        };
    }

    let wide: Vec<u16> = OsStr::new(printer_name)
        .encode_wide()
        .chain(std::iter::once(0))
        .collect();

    unsafe {
        let mut h = ptr::null_mut();
        if OpenPrinterW(wide.as_ptr() as *mut u16, &mut h, ptr::null_mut()) != 0 {
            ClosePrinter(h);
            DeviceStatus {
                name,
                status: "ok".to_string(),
                message: format!("Windows printer '{}' accessible", printer_name),
                device_type: "windows_spooler".to_string(),
            }
        } else {
            DeviceStatus {
                name,
                status: "unreachable".to_string(),
                message: format!("Cannot open Windows printer '{}'", printer_name),
                device_type: "windows_spooler".to_string(),
            }
        }
    }
}

#[cfg(not(windows))]
fn probe_windows_printer(printer_name: &str) -> DeviceStatus {
    DeviceStatus {
        name: printer_name.to_string(),
        status: "not_configured".to_string(),
        message: "Windows spooler not available on this platform".to_string(),
        device_type: "windows_spooler".to_string(),
    }
}

#[tauri::command]
pub fn check_hardware(config: HealthCheckConfig) -> HardwareHealthReport {
    let retries = config.retry_count.unwrap_or(1);

    let printer = match config.printer_transport.as_str() {
        "serial" => {
            let port = config.printer_serial_port.clone().unwrap_or_default();
            let baud = config.printer_baud_rate.unwrap_or(9600);
            probe_with_retry(|| probe_serial(&port, baud), retries)
        }
        "network" => {
            let host = config.printer_host.clone().unwrap_or_default();
            let port = config.printer_port.unwrap_or(9100);
            probe_with_retry(|| probe_tcp(&host, port), retries)
        }
        "windows" => probe_windows_printer(config.printer_name.as_deref().unwrap_or("")),
        _ => DeviceStatus {
            name: "printer".to_string(),
            status: "not_configured".to_string(),
            message: "No printer transport configured".to_string(),
            device_type: "none".to_string(),
        },
    };

    let drawer = match config.drawer_transport.as_deref().unwrap_or("printer") {
        "serial" => {
            let port = config.drawer_serial_port.clone().unwrap_or_default();
            let baud = config.drawer_baud_rate.unwrap_or(9600);
            probe_with_retry(|| probe_serial(&port, baud), retries)
        }
        "network" => {
            let host = config.drawer_host.clone().unwrap_or_default();
            let port = config.drawer_port.unwrap_or(9100);
            probe_with_retry(|| probe_tcp(&host, port), retries)
        }
        // "printer" means drawer shares the printer connection
        _ => DeviceStatus {
            name: "drawer".to_string(),
            status: printer.status.clone(),
            message: "Drawer shares printer connection".to_string(),
            device_type: printer.device_type.clone(),
        },
    };

    let kds = match config.kds_transport.as_str() {
        "serial" => {
            let port = config.kds_serial_port.clone().unwrap_or_default();
            let baud = config.kds_baud_rate.unwrap_or(9600);
            probe_with_retry(|| probe_serial(&port, baud), retries)
        }
        "network" => {
            let host = config.kds_host.clone().unwrap_or_default();
            let port = config.kds_port.unwrap_or(9100);
            probe_with_retry(|| probe_tcp(&host, port), retries)
        }
        _ => DeviceStatus {
            name: "kds".to_string(),
            status: "not_configured".to_string(),
            message: "No KDS hardware configured".to_string(),
            device_type: "none".to_string(),
        },
    };

    HardwareHealthReport { printer, drawer, kds }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn base_config() -> HealthCheckConfig {
        HealthCheckConfig {
            printer_transport: String::new(),
            printer_serial_port: None,
            printer_baud_rate: None,
            printer_host: None,
            printer_port: None,
            printer_name: None,
            drawer_transport: None,
            drawer_serial_port: None,
            drawer_baud_rate: None,
            drawer_host: None,
            drawer_port: None,
            kds_transport: "none".to_string(),
            kds_serial_port: None,
            kds_baud_rate: None,
            kds_host: None,
            kds_port: None,
            retry_count: None,
        }
    }

    #[test]
    fn unconfigured_returns_not_configured() {
        let report = check_hardware(base_config());
        assert_eq!(report.printer.status, "not_configured");
        assert_eq!(report.drawer.status, "not_configured");
        assert_eq!(report.kds.status, "not_configured");
    }

    #[test]
    fn bad_tcp_host_returns_unreachable() {
        let status = probe_tcp("192.0.2.1", 9100); // TEST-NET, guaranteed unreachable
        assert_eq!(status.status, "unreachable");
    }

    #[test]
    fn empty_serial_port_returns_not_configured() {
        let status = probe_serial("", 9600);
        assert_eq!(status.status, "not_configured");
    }

    #[test]
    fn nonexistent_serial_port_returns_unreachable() {
        let status = probe_serial("/dev/ttyNONEXISTENT999", 9600);
        assert_eq!(status.status, "unreachable");
    }

    #[test]
    fn kds_none_transport_returns_not_configured() {
        let config = HealthCheckConfig {
            kds_transport: "none".to_string(),
            ..base_config()
        };
        let report = check_hardware(config);
        assert_eq!(report.kds.status, "not_configured");
    }

    #[test]
    fn drawer_inherits_printer_status_when_transport_is_printer() {
        let config = HealthCheckConfig {
            printer_transport: "network".to_string(),
            printer_host: Some("192.0.2.1".to_string()),
            printer_port: Some(9100),
            drawer_transport: Some("printer".to_string()),
            ..base_config()
        };
        let report = check_hardware(config);
        assert_eq!(report.printer.status, "unreachable");
        assert_eq!(report.drawer.status, "unreachable");
        assert_eq!(report.drawer.device_type, report.printer.device_type);
    }

    #[test]
    fn tcp_port_9100_infers_thermal_escpos() {
        let status = probe_tcp("192.0.2.1", 9100);
        assert_eq!(status.device_type, "thermal_escpos");
    }

    #[test]
    fn tcp_port_515_infers_network_lpr() {
        let status = probe_tcp("192.0.2.1", 515);
        assert_eq!(status.device_type, "network_lpr");
    }

    #[test]
    fn tcp_port_631_infers_network_ipp() {
        let status = probe_tcp("192.0.2.1", 631);
        assert_eq!(status.device_type, "network_ipp");
    }

    #[test]
    fn serial_port_infers_serial_escpos() {
        let status = probe_serial("/dev/ttyNONEXISTENT999", 9600);
        assert_eq!(status.device_type, "serial_escpos");
    }

    #[test]
    fn unconfigured_printer_returns_none_device_type() {
        let report = check_hardware(base_config());
        assert_eq!(report.printer.device_type, "none");
        assert_eq!(report.kds.device_type, "none");
    }

    // ── ESC/POS protocol probe tests ─────────────────────────────────────────

    #[test]
    fn probe_escpos_returns_confirmed_when_printer_responds_with_status_byte() {
        use std::io::Write as IoWrite;
        use std::net::TcpListener;
        use std::thread;

        // Mock ESC/POS printer: accept connection, read request, reply with 1 status byte
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((mut client, _)) = listener.accept() {
                let mut buf = [0u8; 3];
                let _ = client.read(&mut buf);
                // Reply with ESC/POS printer status byte (0x12 = OK, online, no errors)
                let _ = client.write_all(&[0x12]);
            }
        });

        // Give the listener thread a moment to start
        std::thread::sleep(std::time::Duration::from_millis(10));

        let addr_str = format!("127.0.0.1:{}", port);
        let addr: std::net::SocketAddr = addr_str.parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(
            &addr,
            Duration::from_secs(2),
        ).expect("connect");

        let confirmed = probe_escpos_protocol(&mut stream);
        assert!(confirmed, "Should confirm ESC/POS when printer sends status byte");
    }

    #[test]
    fn probe_escpos_returns_assumed_when_no_status_byte_received() {
        use std::net::TcpListener;
        use std::thread;

        // Mock raw TCP listener that accepts but sends no response (dumb network printer)
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((_client, _)) = listener.accept() {
                // Accept but don't respond — simulates plain raw TCP printer
                std::thread::sleep(std::time::Duration::from_millis(600));
            }
        });

        std::thread::sleep(std::time::Duration::from_millis(10));

        let addr_str = format!("127.0.0.1:{}", port);
        let addr: std::net::SocketAddr = addr_str.parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(
            &addr,
            Duration::from_secs(2),
        ).expect("connect");

        let confirmed = probe_escpos_protocol(&mut stream);
        assert!(!confirmed, "Should not confirm ESC/POS when no status byte received");
    }

    // ── ZPL protocol probe tests ──────────────────────────────────────────────

    #[test]
    fn probe_zpl_returns_confirmed_when_zebra_responds_with_stx() {
        use std::io::Write as IoWrite;
        use std::net::TcpListener;
        use std::thread;

        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((mut client, _)) = listener.accept() {
                let mut buf = [0u8; 5];
                let _ = client.read(&mut buf);
                // Zebra ZPL host identification response starts with STX (0x02)
                let _ = client.write_all(b"\x02NXX,XX,1\x03");
            }
        });

        std::thread::sleep(std::time::Duration::from_millis(10));

        let addr_str = format!("127.0.0.1:{}", port);
        let addr: std::net::SocketAddr = addr_str.parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2))
            .expect("connect");

        assert!(probe_zpl_protocol(&mut stream), "Should confirm ZPL when Zebra responds with STX");
    }

    #[test]
    fn tcp_port_6101_infers_zpl_label() {
        let status = probe_tcp("192.0.2.1", 6101);
        assert_eq!(status.device_type, "zpl_label");
    }

    #[test]
    fn tcp_port_9101_infers_star_micronics() {
        let status = probe_tcp("192.0.2.1", 9101);
        assert_eq!(status.device_type, "star_micronics");
    }

    #[test]
    fn tcp_port_9999_infers_kiosk_display() {
        let status = probe_tcp("192.0.2.1", 9999);
        assert_eq!(status.device_type, "kiosk_display");
    }

    #[test]
    fn tcp_port_8008_infers_epson_epos() {
        assert_eq!(infer_tcp_device_type(8008), "epson_epos");
    }

    #[test]
    fn tcp_port_8043_infers_epson_epos() {
        assert_eq!(infer_tcp_device_type(8043), "epson_epos");
    }

    // ── Retry logic tests ─────────────────────────────────────────────────────

    #[test]
    fn probe_with_retry_returns_immediately_on_first_ok() {
        let call_count = std::sync::Arc::new(std::sync::atomic::AtomicU32::new(0));
        let count = std::sync::Arc::clone(&call_count);

        let result = probe_with_retry(
            move || {
                count.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
                DeviceStatus {
                    name: "test".to_string(),
                    status: "ok".to_string(),
                    message: "ok".to_string(),
                    device_type: "unknown".to_string(),
                }
            },
            3,
        );

        assert_eq!(result.status, "ok");
        // Should have called exactly once and stopped
        assert_eq!(call_count.load(std::sync::atomic::Ordering::SeqCst), 1);
    }

    #[test]
    fn probe_with_retry_retries_on_failure_and_returns_last() {
        let call_count = std::sync::Arc::new(std::sync::atomic::AtomicU32::new(0));
        let count = std::sync::Arc::clone(&call_count);

        let result = probe_with_retry(
            move || {
                count.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
                DeviceStatus {
                    name: "test".to_string(),
                    status: "unreachable".to_string(),
                    message: "failed".to_string(),
                    device_type: "unknown".to_string(),
                }
            },
            2,
        );

        assert_eq!(result.status, "unreachable");
        // 1 initial + 2 retries = 3 total calls
        assert_eq!(call_count.load(std::sync::atomic::Ordering::SeqCst), 3);
    }

    // ── Concurrent probe stress test ──────────────────────────────────────────

    /// Spawn 32 threads each calling probe_tcp against an unreachable address.
    /// Verifies: no panic, no deadlock, all results collected correctly,
    /// and the shared result set has exactly one entry per thread (no data races).
    #[test]
    fn concurrent_probes_do_not_panic_or_race() {
        use std::sync::{Arc, Mutex};

        let results: Arc<Mutex<Vec<DeviceStatus>>> = Arc::new(Mutex::new(Vec::new()));
        let mut handles = Vec::new();

        for i in 0..32u16 {
            let results = Arc::clone(&results);
            // Each thread targets a distinct port in the TEST-NET range (192.0.2.x)
            // so they all get a fast ECONNREFUSED or timeout without hitting real hardware
            handles.push(std::thread::spawn(move || {
                let status = probe_tcp("192.0.2.1", 9100 + i);
                results.lock().unwrap().push(status);
            }));
        }

        for h in handles {
            h.join().expect("probe thread panicked");
        }

        let collected = results.lock().unwrap();
        assert_eq!(collected.len(), 32, "expected exactly one result per thread");
        // Every probe to an unreachable host must report 'unreachable'
        for status in collected.iter() {
            assert_eq!(status.status, "unreachable", "unexpected status: {:?}", status);
        }
    }
}
