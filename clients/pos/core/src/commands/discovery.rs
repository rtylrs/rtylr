//! Tauri commands for hardware discovery. Thin wrappers - all logic lives
//! in `crate::hardware`.

use crate::hardware::{network_scan, printer_listing, serial_listing};

#[tauri::command]
pub fn list_printers() -> Result<Vec<printer_listing::PrinterInfo>, String> {
    printer_listing::list_printers()
}

#[tauri::command]
pub fn list_serial_ports() -> Result<Vec<serial_listing::SerialPortInfo>, String> {
    serial_listing::list_serial_ports()
}

#[tauri::command(async)]
pub fn discover_network_hardware(
    subnet_prefix: String,
    timeout_ms: u64,
) -> Result<Vec<network_scan::DiscoveredDevice>, String> {
    network_scan::discover(&subnet_prefix, timeout_ms)
}
