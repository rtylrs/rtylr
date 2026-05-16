//! Serial-port enumeration with USB metadata + thermal brand classification.

use serde::Serialize;

use super::brand::classify_thermal_brand;

#[derive(Serialize)]
pub struct SerialPortInfo {
    pub port_name: String,
    pub manufacturer: Option<String>,
    pub product: Option<String>,
    pub vid: Option<u16>,
    pub pid: Option<u16>,
    /// Best-guess thermal printer brand derived from USB VID,
    /// e.g. "epson", "star", "bematech", "elgin", "ncr", "citizen",
    /// "custom", "unknown".
    pub thermal_brand: String,
}

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
            let thermal_brand =
                classify_thermal_brand(vid, manufacturer.as_deref()).to_string();
            SerialPortInfo {
                port_name: p.port_name,
                manufacturer,
                product,
                vid,
                pid,
                thermal_brand,
            }
        })
        .collect();

    Ok(result)
}
