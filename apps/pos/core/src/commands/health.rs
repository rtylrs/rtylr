use std::time::Duration;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DeviceStatus {
    pub name: String,
    pub status: String, // "ok" | "unreachable" | "not_configured"
    pub message: String,
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
}

fn probe_serial(port_name: &str, baud_rate: u32) -> DeviceStatus {
    let name = port_name.to_string();
    if port_name.trim().is_empty() {
        return DeviceStatus {
            name: name.clone(),
            status: "not_configured".to_string(),
            message: "No serial port configured".to_string(),
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
        },
        Err(e) => DeviceStatus {
            name,
            status: "unreachable".to_string(),
            message: format!("Cannot open {}: {}", port_name, e),
        },
    }
}

fn probe_tcp(host: &str, port: u16) -> DeviceStatus {
    let addr_str = format!("{}:{}", host, port);
    let name = addr_str.clone();
    if host.trim().is_empty() {
        return DeviceStatus {
            name,
            status: "not_configured".to_string(),
            message: "No host configured".to_string(),
        };
    }
    let addr = match addr_str.parse::<std::net::SocketAddr>() {
        Ok(a) => a,
        Err(e) => {
            return DeviceStatus {
                name,
                status: "unreachable".to_string(),
                message: format!("Invalid address {}: {}", addr_str, e),
            }
        }
    };
    match std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(3)) {
        Ok(_) => DeviceStatus {
            name,
            status: "ok".to_string(),
            message: format!("TCP {}:{} reachable", host, port),
        },
        Err(e) => DeviceStatus {
            name,
            status: "unreachable".to_string(),
            message: format!("Cannot connect to {}:{}: {}", host, port, e),
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
            }
        } else {
            DeviceStatus {
                name,
                status: "unreachable".to_string(),
                message: format!("Cannot open Windows printer '{}'", printer_name),
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
    }
}

#[tauri::command]
pub fn check_hardware(config: HealthCheckConfig) -> HardwareHealthReport {
    let printer = match config.printer_transport.as_str() {
        "serial" => probe_serial(
            config.printer_serial_port.as_deref().unwrap_or(""),
            config.printer_baud_rate.unwrap_or(9600),
        ),
        "network" => probe_tcp(
            config.printer_host.as_deref().unwrap_or(""),
            config.printer_port.unwrap_or(9100),
        ),
        "windows" => probe_windows_printer(config.printer_name.as_deref().unwrap_or("")),
        _ => DeviceStatus {
            name: "printer".to_string(),
            status: "not_configured".to_string(),
            message: "No printer transport configured".to_string(),
        },
    };

    let drawer = match config.drawer_transport.as_deref().unwrap_or("printer") {
        "serial" => probe_serial(
            config.drawer_serial_port.as_deref().unwrap_or(""),
            config.drawer_baud_rate.unwrap_or(9600),
        ),
        "network" => probe_tcp(
            config.drawer_host.as_deref().unwrap_or(""),
            config.drawer_port.unwrap_or(9100),
        ),
        // "printer" means drawer shares the printer connection
        _ => DeviceStatus {
            name: "drawer".to_string(),
            status: printer.status.clone(),
            message: "Drawer shares printer connection".to_string(),
        },
    };

    let kds = match config.kds_transport.as_str() {
        "serial" => probe_serial(
            config.kds_serial_port.as_deref().unwrap_or(""),
            config.kds_baud_rate.unwrap_or(9600),
        ),
        "network" => probe_tcp(
            config.kds_host.as_deref().unwrap_or(""),
            config.kds_port.unwrap_or(9100),
        ),
        _ => DeviceStatus {
            name: "kds".to_string(),
            status: "not_configured".to_string(),
            message: "No KDS hardware configured".to_string(),
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
        // Both should be unreachable (TEST-NET IP) and drawer mirrors printer
        assert_eq!(report.printer.status, "unreachable");
        assert_eq!(report.drawer.status, "unreachable");
    }
}
