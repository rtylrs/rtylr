use std::io::Write;
use std::time::Duration;

// ESC/POS cash drawer kick sequences
// Pin 2 = 0x00, Pin 5 = 0x01
const DRAWER_PIN2: [u8; 5] = [0x1B, 0x70, 0x00, 0x19, 0xFA];
const DRAWER_PIN5: [u8; 5] = [0x1B, 0x70, 0x01, 0x19, 0xFA];

fn drawer_bytes(pin: u8) -> &'static [u8] {
    if pin == 5 {
        &DRAWER_PIN5
    } else {
        &DRAWER_PIN2
    }
}

/// Open cash drawer via a serial port (RJ11/RJ12 or USB-to-serial adapter).
/// `pin` is 2 or 5 (the cash drawer kick pin on the printer).
#[tauri::command]
pub fn open_cash_drawer_serial(
    port_name: String,
    baud_rate: u32,
    pin: u8,
) -> Result<String, String> {
    let port_name = port_name.trim().to_string();
    if port_name.is_empty() {
        return Err("Serial port name is required".to_string());
    }

    let baud = if baud_rate == 0 { 9600 } else { baud_rate };
    let pin = if pin == 5 { 5 } else { 2 };

    let mut port = serialport::new(&port_name, baud)
        .timeout(Duration::from_millis(2000))
        .open()
        .map_err(|e| format!("Cannot open port '{}': {}", port_name, e))?;

    port.write_all(drawer_bytes(pin))
        .map_err(|e| format!("Write to port '{}' failed: {}", port_name, e))?;

    Ok(format!("Drawer opened via serial {}", port_name))
}

/// Open cash drawer over the network (e.g. network-connected thermal printer).
/// Sends ESC/POS drawer kick to `host:port` via TCP.
#[tauri::command]
pub fn open_cash_drawer_network(host: String, port: u16, pin: u8) -> Result<String, String> {
    let host = host.trim().to_string();
    if host.is_empty() {
        return Err("Host is required".to_string());
    }
    if port == 0 {
        return Err("Port must be non-zero".to_string());
    }

    let pin = if pin == 5 { 5 } else { 2 };
    let addr = format!("{}:{}", host, port);

    let mut stream = std::net::TcpStream::connect_timeout(
        &addr.parse().map_err(|_| format!("Invalid address: {}", addr))?,
        Duration::from_secs(5),
    )
    .map_err(|e| format!("Cannot connect to {}: {}", addr, e))?;

    stream
        .write_all(drawer_bytes(pin))
        .map_err(|e| format!("Write to {} failed: {}", addr, e))?;

    Ok(format!("Drawer opened via network {}:{}", host, port))
}
