//! TCP write transport for network-attached printers, KDS, drawers.

use std::io::Write;
use std::net::{SocketAddr, TcpStream, ToSocketAddrs};
use std::time::Duration;

/// Resolve `host:port` (IP literal OR DNS hostname - validate_host accepts
/// both) into concrete socket addresses. SocketAddr parsing alone rejected
/// every hostname at runtime.
fn resolve(host: &str, port: u16) -> Result<Vec<SocketAddr>, String> {
    (host, port)
        .to_socket_addrs()
        .map_err(|e| format!("Cannot resolve '{}:{}': {}", host, port, e))
        .map(|addrs| addrs.collect())
}

pub fn write_all(host: &str, port: u16, connect_timeout_secs: u64, data: &[u8]) -> Result<(), String> {
    let addrs = resolve(host, port)?;
    let mut last_err: Option<String> = None;
    for addr in addrs {
        match TcpStream::connect_timeout(&addr, Duration::from_secs(connect_timeout_secs)) {
            Ok(mut stream) => {
                return stream.write_all(data).map_err(|e| {
                    format!("Network write to {}:{} failed: {}", host, port, e)
                });
            }
            Err(e) => last_err = Some(format!("Cannot connect to {}:{}: {}", host, port, e)),
        }
    }
    Err(last_err.unwrap_or_else(|| format!("No addresses resolved for '{}:{}'", host, port)))
}

pub fn probe_connect(host: &str, port: u16, timeout_secs: u64) -> Result<TcpStream, String> {
    let addrs = resolve(host, port)?;
    let mut last_err: Option<String> = None;
    for addr in addrs {
        match TcpStream::connect_timeout(&addr, Duration::from_secs(timeout_secs)) {
            Ok(stream) => return Ok(stream),
            Err(e) => {
                last_err = Some(format!("Cannot connect to {}:{}: {}", host, port, e));
            }
        }
    }
    Err(last_err.unwrap_or_else(|| format!("No addresses resolved for '{}:{}'", host, port)))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn write_all_rejects_unresolvable_host() {
        let result = write_all("nonexistent.invalid", 9100, 1, b"x");
        assert!(result.is_err());
    }

    #[test]
    fn probe_connect_reports_resolution_failure_clearly() {
        let err = probe_connect("nonexistent.invalid", 9100, 1).unwrap_err();
        assert!(err.contains("resolve"), "unexpected error: {}", err);
    }
}
