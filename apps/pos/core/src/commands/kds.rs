//! KDS-ticket Tauri commands.

use crate::hardware::escpos::FORM_FEED;
use crate::hardware::kds::formatter::{format_ticket_text, KdsTicket};
use crate::transport::{network, serial};
use crate::validation::{
    kds::validate_kds_ticket,
    net::{validate_baud, validate_host, validate_port, validate_serial_port_name},
};

#[tauri::command(async)]
pub fn send_kds_ticket_serial(
    port_name: String,
    baud_rate: u32,
    ticket: KdsTicket,
) -> Result<String, String> {
    validate_kds_ticket(&ticket)?;
    let port_name = validate_serial_port_name(&port_name)?;
    let baud = validate_baud(baud_rate)?;

    let data = format_ticket_text(&ticket);
    serial::write_all(&port_name, baud, 3000, data.as_bytes())?;
    Ok(format!("KDS ticket sent to serial {}", port_name))
}

#[tauri::command(async)]
pub fn send_kds_ticket_network(
    host: String,
    port: u16,
    ticket: KdsTicket,
) -> Result<String, String> {
    validate_kds_ticket(&ticket)?;
    let host = validate_host(&host)?;
    let port = validate_port(port)?;

    let mut data = format_ticket_text(&ticket);
    data.push(FORM_FEED as char);

    network::write_all(&host, port, 5, data.as_bytes())?;
    Ok(format!("KDS ticket sent to {}:{}", host, port))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::hardware::kds::formatter::KdsItem;
    use std::io::Read;
    use std::net::TcpListener;
    use std::thread;

    fn sample_ticket() -> KdsTicket {
        KdsTicket {
            order_number: "42".to_string(),
            station: Some("GRILL".to_string()),
            order_type: Some("Dine In".to_string()),
            table_number: Some("7".to_string()),
            created_at: Some("2026-05-13T14:30:00Z".to_string()),
            items: vec![KdsItem {
                name: "Burger".to_string(),
                quantity: 2.0,
                notes: Some("No pickles".to_string()),
                modifiers: vec!["Extra cheese".to_string()],
            }],
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
    fn network_kds_ends_with_form_feed() {
        let (result, received) = with_tcp_mock(|port| {
            send_kds_ticket_network("127.0.0.1".to_string(), port, sample_ticket())
        });
        assert!(result.is_ok());
        assert_eq!(received.last().copied(), Some(0x0C));
    }

    #[test]
    fn network_kds_sends_order_and_items() {
        let (result, received) = with_tcp_mock(|port| {
            send_kds_ticket_network("127.0.0.1".to_string(), port, sample_ticket())
        });
        assert!(result.is_ok());
        let text = String::from_utf8_lossy(&received);
        assert!(text.contains("ORDER #42") && text.contains("Burger") && text.contains("Extra cheese"));
    }

    #[test]
    fn rejects_invalid_ticket() {
        let mut t = sample_ticket();
        t.items.clear();
        let result = send_kds_ticket_network("127.0.0.1".to_string(), 9200, t);
        assert!(result.is_err());
    }
}
