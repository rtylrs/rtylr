//! Serial-port write transport for ESC/POS / KDS / drawer hardware.

use std::io::Write;
use std::time::Duration;

pub fn write_all(port_name: &str, baud: u32, timeout_ms: u64, data: &[u8]) -> Result<(), String> {
    let mut port = serialport::new(port_name, baud)
        .timeout(Duration::from_millis(timeout_ms))
        .open()
        .map_err(|e| format!("Cannot open serial port '{}': {}", port_name, e))?;

    port.write_all(data)
        .map_err(|e| format!("Serial write to '{}' failed: {}", port_name, e))?;

    // Some USB-serial adapters buffer internally; without an explicit flush
    // the fd can close before the bytes reach the device.
    port.flush()
        .map_err(|e| format!("Serial flush on '{}' failed: {}", port_name, e))?;

    Ok(())
}

/// Opens the port without writing - used by hardware health checks.
pub fn probe_open(port_name: &str, baud: u32, timeout_ms: u64) -> Result<(), String> {
    serialport::new(port_name, baud)
        .timeout(Duration::from_millis(timeout_ms))
        .open()
        .map(|_| ())
        .map_err(|e| format!("Cannot open {}: {}", port_name, e))
}
