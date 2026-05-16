//! TCP write transport for network-attached printers, KDS, drawers.

use std::io::Write;
use std::net::TcpStream;
use std::time::Duration;

pub fn write_all(host: &str, port: u16, connect_timeout_secs: u64, data: &[u8]) -> Result<(), String> {
    let addr = format!("{}:{}", host, port);
    let mut stream = TcpStream::connect_timeout(
        &addr.parse().map_err(|_| format!("Invalid address: {}", addr))?,
        Duration::from_secs(connect_timeout_secs),
    )
    .map_err(|e| format!("Cannot connect to {}: {}", addr, e))?;

    stream
        .write_all(data)
        .map_err(|e| format!("Network write to {} failed: {}", addr, e))?;

    Ok(())
}

pub fn probe_connect(host: &str, port: u16, timeout_secs: u64) -> Result<TcpStream, String> {
    let addr = format!("{}:{}", host, port);
    TcpStream::connect_timeout(
        &addr.parse().map_err(|_| format!("Invalid address: {}", addr))?,
        Duration::from_secs(timeout_secs),
    )
    .map_err(|e| format!("Cannot connect to {}:{}: {}", host, port, e))
}
