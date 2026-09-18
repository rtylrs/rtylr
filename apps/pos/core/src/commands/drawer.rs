//! Tauri commands for cash-drawer kick (serial / network).

use crate::hardware::escpos::drawer_bytes;
use crate::transport::{network, serial};
use crate::validation::{
    drawer::validate_drawer_pin,
    net::{validate_baud, validate_host, validate_port, validate_serial_port_name},
};

/// Open the cash drawer via a serial port (RJ11/RJ12 or USB-to-serial adapter).
#[tauri::command(async)]
pub fn open_cash_drawer_serial(
    port_name: String,
    baud_rate: u32,
    pin: u8,
) -> Result<String, String> {
    let port_name = validate_serial_port_name(&port_name)?;
    let baud = validate_baud(baud_rate)?;
    let pin = validate_drawer_pin(pin)?;

    serial::write_all(&port_name, baud, 2000, drawer_bytes(pin))?;
    Ok(format!("Drawer opened via serial {}", port_name))
}

/// Open the cash drawer over the network (network-attached thermal printer).
#[tauri::command(async)]
pub fn open_cash_drawer_network(host: String, port: u16, pin: u8) -> Result<String, String> {
    let host = validate_host(&host)?;
    let port = validate_port(port)?;
    let pin = validate_drawer_pin(pin)?;

    network::write_all(&host, port, 5, drawer_bytes(pin))?;
    Ok(format!("Drawer opened via network {}:{}", host, port))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn network_drawer_rejects_zero_port() {
        let result = open_cash_drawer_network("127.0.0.1".to_string(), 0, 2);
        assert!(result.is_err());
    }

    #[test]
    fn network_drawer_rejects_empty_host() {
        let result = open_cash_drawer_network("".to_string(), 9100, 2);
        assert!(result.is_err());
    }

    #[test]
    fn network_drawer_rejects_bad_pin() {
        let result = open_cash_drawer_network("127.0.0.1".to_string(), 9100, 3);
        assert!(result.is_err());
    }

    #[test]
    fn serial_drawer_rejects_empty_port() {
        let result = open_cash_drawer_serial("".to_string(), 9600, 2);
        assert!(result.is_err());
    }

    #[test]
    fn serial_drawer_rejects_bad_baud() {
        let result = open_cash_drawer_serial("COM3".to_string(), 123, 2);
        assert!(result.is_err());
    }

    // ── Wire-format integration: TCP mock receives ESC/POS drawer kick ─────────
    use std::io::Read;
    use std::net::TcpListener;
    use std::thread;

    fn with_tcp_mock<F: FnOnce(u16) -> Result<String, String>>(f: F) -> (Result<String, String>, Vec<u8>) {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();
        let recv_handle = thread::spawn(move || {
            let (mut stream, _) = listener.accept().expect("accept");
            let mut buf = Vec::new();
            stream.read_to_end(&mut buf).expect("read");
            buf
        });
        let result = f(port);
        let received = recv_handle.join().expect("join");
        (result, received)
    }

    #[test]
    fn network_drawer_sends_pin2_escpos_bytes_over_tcp() {
        let (result, received) = with_tcp_mock(|port| {
            open_cash_drawer_network("127.0.0.1".to_string(), port, 2)
        });
        assert!(result.is_ok());
        assert_eq!(received.as_slice(), &[0x1B, 0x70, 0x00, 0x19, 0xFA]);
    }

    #[test]
    fn network_drawer_sends_pin5_escpos_bytes_over_tcp() {
        let (result, received) = with_tcp_mock(|port| {
            open_cash_drawer_network("127.0.0.1".to_string(), port, 5)
        });
        assert!(result.is_ok());
        assert_eq!(received.as_slice(), &[0x1B, 0x70, 0x01, 0x19, 0xFA]);
    }
}
