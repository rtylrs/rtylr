//! Live-protocol probes used by the hardware health check.
//!
//! Each probe assumes a TCP connection is already established (or, for the
//! HTTP one, makes its own) and tries to elicit a vendor-specific response
//! so we can distinguish "port answered" from "device speaks the protocol
//! we expect".

use std::io::{Read, Write};
use std::time::Duration;

use crate::hardware::escpos::STATUS_REQUEST;

/// Zebra ZPL host identification command.
/// A Zebra printer responds with STX + model info within ~200 ms.
pub const ZPL_HOST_IDENT: &[u8] = b"~HI\r\n";

/// Confirm ESC/POS protocol on an already-connected stream. Returns
/// `true` if the printer replied with at least one status byte.
pub fn probe_escpos_protocol(stream: &mut std::net::TcpStream) -> bool {
    stream.set_write_timeout(Some(Duration::from_millis(200))).ok();
    stream.set_read_timeout(Some(Duration::from_millis(300))).ok();

    if stream.write_all(STATUS_REQUEST).is_err() {
        return false;
    }

    let mut buf = [0u8; 1];
    matches!(stream.read_exact(&mut buf), Ok(()))
}

/// Confirm Zebra ZPL protocol on an already-connected stream.
/// Returns `true` on a Zebra STX (`0x02`) or `'N'` model-info reply.
pub fn probe_zpl_protocol(stream: &mut std::net::TcpStream) -> bool {
    stream.set_write_timeout(Some(Duration::from_millis(200))).ok();
    stream.set_read_timeout(Some(Duration::from_millis(400))).ok();

    if stream.write_all(ZPL_HOST_IDENT).is_err() {
        return false;
    }

    let mut buf = [0u8; 4];
    if stream.read(&mut buf).unwrap_or(0) == 0 {
        return false;
    }
    buf[0] == 0x02 || buf[0] == b'N'
}

/// Probe an Epson ePOS-Print HTTP endpoint (port 8008/8043) over an
/// already-connected stream, so health checks can reuse the connection
/// opened by `probe_connect`. Returns `true` when the device answers with
/// HTTP 2xx or 4xx - either confirms an ePOS HTTP server is present.
pub fn probe_epos_http_on(stream: &mut std::net::TcpStream, host: &str, port: u16) -> bool {
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
    let Some(status) = parse_http_status(&buf[..n]) else { return false; };
    // Only the status CODE decides - substring matching matched digits in
    // headers/dates and reported HTTP 500s as confirmed printers.
    status == 2 || status == 4
}

/// Parse the three-digit status code out of an HTTP/1.x response's first
/// line. Returns its leading digit (2 for 2xx, 4 for 4xx, ...).
fn parse_http_status(resp: &[u8]) -> Option<u8> {
    if !resp.starts_with(b"HTTP/1") {
        return None;
    }
    // "HTTP/1.x NNN ..." - the status code starts at byte offset 9.
    let code = resp.get(9..12)?;
    if !code.iter().all(|b| b.is_ascii_digit()) {
        return None;
    }
    Some(code[0] - b'0')
}

/// Map a well-known TCP port to its expected device type label.
pub fn infer_tcp_device_type(port: u16) -> &'static str {
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

#[cfg(test)]
mod tests {
    use super::*;
    use std::net::TcpListener;
    use std::thread;

    #[test]
    fn probe_escpos_returns_confirmed_when_printer_responds_with_status_byte() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((mut client, _)) = listener.accept() {
                let mut buf = [0u8; 3];
                let _ = client.read(&mut buf);
                let _ = client.write_all(&[0x12]);
            }
        });

        std::thread::sleep(Duration::from_millis(10));

        let addr: std::net::SocketAddr = format!("127.0.0.1:{}", port).parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2)).expect("connect");

        assert!(probe_escpos_protocol(&mut stream));
    }

    #[test]
    fn probe_escpos_returns_assumed_when_no_status_byte_received() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((_client, _)) = listener.accept() {
                std::thread::sleep(Duration::from_millis(600));
            }
        });

        std::thread::sleep(Duration::from_millis(10));

        let addr: std::net::SocketAddr = format!("127.0.0.1:{}", port).parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2)).expect("connect");

        assert!(!probe_escpos_protocol(&mut stream));
    }

    #[test]
    fn probe_zpl_returns_confirmed_when_zebra_responds_with_stx() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((mut client, _)) = listener.accept() {
                let mut buf = [0u8; 5];
                let _ = client.read(&mut buf);
                let _ = client.write_all(b"\x02NXX,XX,1\x03");
            }
        });

        std::thread::sleep(Duration::from_millis(10));

        let addr: std::net::SocketAddr = format!("127.0.0.1:{}", port).parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2)).expect("connect");

        assert!(probe_zpl_protocol(&mut stream));
    }

    #[test]
    fn probe_epos_http_returns_true_when_printer_responds_with_http_200() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || {
            if let Ok((mut client, _)) = listener.accept() {
                let mut buf = [0u8; 512];
                let _ = client.read(&mut buf);
                let _ = client.write_all(b"HTTP/1.0 200 OK\r\nContent-Type: text/xml\r\n\r\n<epos-print/>");
            }
        });

        std::thread::sleep(Duration::from_millis(10));
        let addr: std::net::SocketAddr = format!("127.0.0.1:{}", port).parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2)).expect("connect");
        assert!(probe_epos_http_on(&mut stream, "127.0.0.1", port));
    }

    #[test]
    fn probe_epos_http_returns_false_when_no_http_response() {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        thread::spawn(move || { let _ = listener.accept(); });

        std::thread::sleep(Duration::from_millis(10));
        let addr: std::net::SocketAddr = format!("127.0.0.1:{}", port).parse().unwrap();
        let mut stream = std::net::TcpStream::connect_timeout(&addr, Duration::from_secs(2)).expect("connect");
        assert!(!probe_epos_http_on(&mut stream, "127.0.0.1", port));
    }

    #[test] fn port_9100_infers_thermal_escpos() { assert_eq!(infer_tcp_device_type(9100), "thermal_escpos"); }
    #[test] fn port_9101_infers_star_micronics() { assert_eq!(infer_tcp_device_type(9101), "star_micronics"); }
    #[test] fn port_6101_infers_zpl_label() { assert_eq!(infer_tcp_device_type(6101), "zpl_label"); }
    #[test] fn port_515_infers_network_lpr() { assert_eq!(infer_tcp_device_type(515), "network_lpr"); }
    #[test] fn port_631_infers_network_ipp() { assert_eq!(infer_tcp_device_type(631), "network_ipp"); }
    #[test] fn port_9999_infers_kiosk_display() { assert_eq!(infer_tcp_device_type(9999), "kiosk_display"); }
    #[test] fn port_8008_infers_epson_epos() { assert_eq!(infer_tcp_device_type(8008), "epson_epos"); }
    #[test] fn port_8043_infers_epson_epos() { assert_eq!(infer_tcp_device_type(8043), "epson_epos"); }
}
