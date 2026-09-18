//! Hardware health probes (printer, drawer, KDS).
//!
//! Probe code uses transports for connection + hardware/probe.rs for
//! protocol-level identification. Each probe is wrapped with a retry
//! loop and returned as a `DeviceStatus`.

use serde::{Deserialize, Serialize};
use std::time::Duration;

use crate::hardware::probe::{infer_tcp_device_type, probe_epos_http_on, probe_escpos_protocol, probe_zpl_protocol};
use crate::transport::{network as net_tx, serial as serial_tx, windows_spooler};

#[derive(Debug, Serialize)]
pub struct DeviceStatus {
    pub name: String,
    pub status: String,
    pub message: String,
    pub device_type: String,
}

#[derive(Debug, Serialize)]
pub struct HardwareHealthReport {
    pub printer: DeviceStatus,
    pub drawer: DeviceStatus,
    pub kds: DeviceStatus,
}

#[derive(Debug, Deserialize)]
pub struct HealthCheckConfig {
    pub printer_transport: String,
    pub printer_serial_port: Option<String>,
    pub printer_baud_rate: Option<u32>,
    pub printer_host: Option<String>,
    pub printer_port: Option<u16>,
    pub printer_name: Option<String>,

    pub drawer_transport: Option<String>,
    pub drawer_serial_port: Option<String>,
    pub drawer_baud_rate: Option<u32>,
    pub drawer_host: Option<String>,
    pub drawer_port: Option<u16>,

    pub kds_transport: String,
    pub kds_serial_port: Option<String>,
    pub kds_baud_rate: Option<u32>,
    pub kds_host: Option<String>,
    pub kds_port: Option<u16>,

    pub retry_count: Option<u32>,
}

fn probe_with_retry<F: Fn() -> DeviceStatus>(probe: F, retries: u32) -> DeviceStatus {
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
            name,
            status: "not_configured".to_string(),
            message: "No serial port configured".to_string(),
            device_type: "none".to_string(),
        };
    }
    let baud = if baud_rate == 0 { 9600 } else { baud_rate };
    match serial_tx::probe_open(port_name, baud, 2000) {
        Ok(()) => DeviceStatus {
            name,
            status: "ok".to_string(),
            message: format!("Serial port {} reachable at {} baud", port_name, baud),
            device_type: "serial_escpos".to_string(),
        },
        Err(e) => DeviceStatus {
            name,
            status: "unreachable".to_string(),
            message: e,
            device_type: "serial_escpos".to_string(),
        },
    }
}

fn probe_tcp(host: &str, port: u16) -> DeviceStatus {
    let name = format!("{}:{}", host, port);
    let device_type_hint = infer_tcp_device_type(port);

    if host.trim().is_empty() {
        return DeviceStatus {
            name,
            status: "not_configured".to_string(),
            message: "No host configured".to_string(),
            device_type: device_type_hint.to_string(),
        };
    }

    match net_tx::probe_connect(host, port, 3) {
        Ok(mut stream) => {
            let (device_type, detail) = match port {
                9100 => {
                    if probe_escpos_protocol(&mut stream) {
                        ("thermal_escpos_confirmed", "ESC/POS status byte received - thermal printer confirmed")
                    } else {
                        ("thermal_escpos", "Connected on port 9100 - ESC/POS assumed (no status response)")
                    }
                }
                6101 => {
                    if probe_zpl_protocol(&mut stream) {
                        ("zpl_label_confirmed", "Zebra ZPL response received - label printer confirmed")
                    } else {
                        ("zpl_label", "Connected on port 6101 - ZPL assumed (no identification response)")
                    }
                }
                8008 | 8043 => {
                    // Reuse the already-open stream instead of dialing a
                    // second connection to the same host:port.
                    if probe_epos_http_on(&mut stream, host, port) {
                        ("epson_epos_confirmed", "Epson ePOS-Print HTTP response confirmed")
                    } else {
                        ("epson_epos", "Connected on ePOS port - Epson ePOS assumed (no HTTP response)")
                    }
                }
                _ => (device_type_hint, "TCP connection established"),
            };
            DeviceStatus {
                name,
                status: "ok".to_string(),
                message: format!("{}:{} reachable - {}", host, port, detail),
                device_type: device_type.to_string(),
            }
        }
        Err(e) => DeviceStatus {
            name,
            status: "unreachable".to_string(),
            message: e,
            device_type: device_type_hint.to_string(),
        },
    }
}

fn probe_windows_printer(printer_name: &str) -> DeviceStatus {
    if printer_name.trim().is_empty() {
        return DeviceStatus {
            name: printer_name.to_string(),
            status: "not_configured".to_string(),
            message: "No printer name configured".to_string(),
            device_type: "windows_spooler".to_string(),
        };
    }
    match windows_spooler::probe_open(printer_name) {
        Ok(()) => DeviceStatus {
            name: printer_name.to_string(),
            status: "ok".to_string(),
            message: format!("Windows printer '{}' accessible", printer_name),
            device_type: "windows_spooler".to_string(),
        },
        Err(e) => DeviceStatus {
            name: printer_name.to_string(),
            status: "unreachable".to_string(),
            message: e,
            device_type: "windows_spooler".to_string(),
        },
    }
}

fn probe_printer(config: &HealthCheckConfig, retries: u32) -> DeviceStatus {
    match config.printer_transport.as_str() {
        "serial" => {
            let port = config.printer_serial_port.as_deref().unwrap_or_default().to_string();
            let baud = config.printer_baud_rate.unwrap_or(9600);
            probe_with_retry(|| probe_serial(&port, baud), retries)
        }
        "network" => {
            let host = config.printer_host.as_deref().unwrap_or_default().to_string();
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
    }
}

fn probe_drawer(config: &HealthCheckConfig, retries: u32) -> DeviceStatus {
    match config.drawer_transport.as_deref().unwrap_or("printer") {
        "serial" => {
            let port = config.drawer_serial_port.as_deref().unwrap_or_default().to_string();
            let baud = config.drawer_baud_rate.unwrap_or(9600);
            probe_with_retry(|| probe_serial(&port, baud), retries)
        }
        "network" => {
            let host = config.drawer_host.as_deref().unwrap_or_default().to_string();
            let port = config.drawer_port.unwrap_or(9100);
            probe_with_retry(|| probe_tcp(&host, port), retries)
        }
        // Handled by the caller, which inherits the printer result.
        _ => DeviceStatus {
            name: "drawer".to_string(),
            status: "not_configured".to_string(),
            message: "Drawer shares printer connection".to_string(),
            device_type: "none".to_string(),
        },
    }
}

fn probe_kds(config: &HealthCheckConfig, retries: u32) -> DeviceStatus {
    match config.kds_transport.as_str() {
        "serial" => {
            let port = config.kds_serial_port.as_deref().unwrap_or_default().to_string();
            let baud = config.kds_baud_rate.unwrap_or(9600);
            probe_with_retry(|| probe_serial(&port, baud), retries)
        }
        "network" => {
            let host = config.kds_host.as_deref().unwrap_or_default().to_string();
            let port = config.kds_port.unwrap_or(9100);
            probe_with_retry(|| probe_tcp(&host, port), retries)
        }
        _ => DeviceStatus {
            name: "kds".to_string(),
            status: "not_configured".to_string(),
            message: "No KDS hardware configured".to_string(),
            device_type: "none".to_string(),
        },
    }
}

#[tauri::command(async)]
pub fn check_hardware(config: HealthCheckConfig) -> HardwareHealthReport {
    // Frontend-supplied: clamp so a bogus value cannot hang the terminal.
    let retries = config.retry_count.unwrap_or(1).clamp(0, 5);
    // Without an explicit drawer transport the drawer rides on the printer
    // connection and needs no probe of its own.
    let drawer_has_own_transport =
        matches!(config.drawer_transport.as_deref(), Some("serial") | Some("network"));

    // Probe independent devices concurrently - serial execution stacked
    // three multi-second probes (plus retries) into one long stall.
    std::thread::scope(|scope| {
        let printer_handle = scope.spawn(|| probe_printer(&config, retries));
        let drawer_handle = drawer_has_own_transport.then(|| scope.spawn(|| probe_drawer(&config, retries)));
        let kds_handle = scope.spawn(|| probe_kds(&config, retries));

        let printer = printer_handle.join().expect("printer probe thread panicked");
        let drawer = match drawer_handle {
            Some(handle) => handle.join().expect("drawer probe thread panicked"),
            None => DeviceStatus {
                name: "drawer".to_string(),
                status: printer.status.clone(),
                message: "Drawer shares printer connection".to_string(),
                device_type: printer.device_type.clone(),
            },
        };
        let kds = kds_handle.join().expect("kds probe thread panicked");

        HardwareHealthReport { printer, drawer, kds }
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::atomic::{AtomicU32, Ordering};
    use std::sync::Arc;

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
        assert_eq!(probe_tcp("192.0.2.1", 9100).status, "unreachable");
    }

    #[test]
    fn empty_serial_port_returns_not_configured() {
        assert_eq!(probe_serial("", 9600).status, "not_configured");
    }

    #[test]
    fn nonexistent_serial_port_returns_unreachable() {
        assert_eq!(probe_serial("/dev/ttyNONEXISTENT999", 9600).status, "unreachable");
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
    fn probe_with_retry_returns_immediately_on_first_ok() {
        let calls = Arc::new(AtomicU32::new(0));
        let counter = Arc::clone(&calls);
        let result = probe_with_retry(
            move || {
                counter.fetch_add(1, Ordering::SeqCst);
                DeviceStatus {
                    name: "x".to_string(),
                    status: "ok".to_string(),
                    message: "ok".to_string(),
                    device_type: "unknown".to_string(),
                }
            },
            3,
        );
        assert_eq!(result.status, "ok");
        assert_eq!(calls.load(Ordering::SeqCst), 1);
    }

    #[test]
    fn probe_with_retry_retries_on_failure_and_returns_last() {
        let calls = Arc::new(AtomicU32::new(0));
        let counter = Arc::clone(&calls);
        let result = probe_with_retry(
            move || {
                counter.fetch_add(1, Ordering::SeqCst);
                DeviceStatus {
                    name: "x".to_string(),
                    status: "unreachable".to_string(),
                    message: "fail".to_string(),
                    device_type: "unknown".to_string(),
                }
            },
            2,
        );
        assert_eq!(result.status, "unreachable");
        assert_eq!(calls.load(Ordering::SeqCst), 3);
    }

    #[test]
    fn concurrent_probes_do_not_panic_or_race() {
        use std::sync::Mutex;

        let results: Arc<Mutex<Vec<DeviceStatus>>> = Arc::new(Mutex::new(Vec::new()));
        let mut handles = Vec::new();

        for i in 0..32u16 {
            let results = Arc::clone(&results);
            handles.push(std::thread::spawn(move || {
                let status = probe_tcp("192.0.2.1", 9100 + i);
                results.lock().unwrap().push(status);
            }));
        }

        for h in handles {
            h.join().expect("probe thread panicked");
        }

        let collected = results.lock().unwrap();
        assert_eq!(collected.len(), 32);
        for s in collected.iter() {
            assert_eq!(s.status, "unreachable");
        }
    }
}
