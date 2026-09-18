//! `print_receipt` Tauri command.

use serde::Deserialize;

use crate::hardware::receipt_builder::{build_receipt, ReceiptData};
use crate::transport::{network, serial, windows_spooler};
use crate::validation::{
    net::{validate_baud, validate_host, validate_port, validate_serial_port_name},
    receipt::validate_receipt,
};

#[derive(Deserialize)]
pub struct PrinterConfig {
    /// "windows", "serial", or "network"
    pub transport: String,
    /// Windows: printer name / Serial: port name / Network: host
    pub target: String,
    /// Serial: baud rate (default 9600) / Network: TCP port (default 9100)
    pub port: Option<u32>,
}

#[tauri::command(async)]
pub fn print_receipt(config: PrinterConfig, receipt: ReceiptData) -> Result<String, String> {
    validate_receipt(&receipt)?;
    let data = build_receipt(&receipt);

    match config.transport.as_str() {
        "serial" => {
            let port_name = validate_serial_port_name(&config.target)?;
            let baud = validate_baud(config.port.unwrap_or(9600))?;
            serial::write_all(&port_name, baud, 5000, &data)?;
            Ok(format!("Receipt sent to serial {}", port_name))
        }
        "network" => {
            let host = validate_host(&config.target)?;
            let raw_port = config.port.unwrap_or(9100);
            // u32 -> u16 truncation silently corrupted ports like 74636 into 9100.
            let port = validate_port(u16::try_from(raw_port).map_err(|_| {
                format!("Network port {} is out of range", raw_port)
            })?)?;
            network::write_all(&host, port, 8, &data)?;
            Ok(format!("Receipt sent to network {}:{}", host, port))
        }
        "windows" => {
            let name = config.target.trim();
            if name.is_empty() {
                return Err("Windows printer name is required".to_string());
            }
            windows_spooler::write_raw(name, &data, "Receipt")?;
            Ok(format!("Receipt sent to Windows printer '{}'", name))
        }
        other => Err(format!("Unknown transport: '{}'", other)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::hardware::receipt_builder::ReceiptItem;
    use std::io::Read;
    use std::net::TcpListener;
    use std::thread;

    fn sample_receipt() -> ReceiptData {
        ReceiptData {
            business_name: "Test Cafe".to_string(),
            business_address: Some("123 Main".to_string()),
            business_phone: Some("555-1234".to_string()),
            order_number: "ORD-1".to_string(),
            order_type: Some("Dine In".to_string()),
            cashier_name: Some("Alice".to_string()),
            table_number: Some("5".to_string()),
            items: vec![ReceiptItem {
                name: "Espresso".to_string(),
                quantity: 2.0,
                unit_price: 3.50,
                total: 7.00,
                modifiers: vec!["Extra shot".to_string()],
                notes: None,
            }],
            subtotal: 7.00,
            tax_amount: 0.63,
            discount_amount: 0.0,
            tip_amount: 1.40,
            total: 9.03,
            payment_method: "Cash".to_string(),
            amount_paid: 10.00,
            change_due: 0.97,
            footer_message: Some("Thanks!".to_string()),
            open_drawer: false,
            drawer_pin: None,
        }
    }

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
    fn rejects_unknown_transport() {
        let config = PrinterConfig {
            transport: "bluetooth".to_string(),
            target: "x".to_string(),
            port: None,
        };
        assert!(print_receipt(config, sample_receipt()).is_err());
    }

    #[test]
    fn rejects_invalid_receipt_before_attempting_print() {
        let config = PrinterConfig {
            transport: "network".to_string(),
            target: "127.0.0.1".to_string(),
            port: Some(9100),
        };
        let mut bad = sample_receipt();
        bad.total = 999.99; // wrong total
        assert!(print_receipt(config, bad).is_err());
    }

    #[test]
    fn network_receipt_sends_valid_escpos_over_tcp() {
        let (result, received) = with_tcp_mock(|port| {
            print_receipt(
                PrinterConfig {
                    transport: "network".to_string(),
                    target: "127.0.0.1".to_string(),
                    port: Some(port as u32),
                },
                sample_receipt(),
            )
        });
        assert!(result.is_ok(), "{:?}", result);
        assert!(received.starts_with(&[0x1B, 0x40]), "Missing ESC/POS INIT");
        assert!(received.ends_with(&[0x1D, 0x56, 0x41, 0x03]), "Missing CUT");
        let text = String::from_utf8_lossy(&received);
        assert!(text.contains("Test Cafe"));
    }
}
