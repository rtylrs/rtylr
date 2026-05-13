use serde::Deserialize;
use std::io::Write as IoWrite;
use std::time::Duration;

#[derive(Deserialize)]
pub struct KdsItem {
    pub name: String,
    pub quantity: f64,
    pub notes: Option<String>,
    pub modifiers: Vec<String>,
}

#[derive(Deserialize)]
pub struct KdsTicket {
    pub order_number: String,
    pub station: Option<String>,
    pub order_type: Option<String>,
    pub table_number: Option<String>,
    pub items: Vec<KdsItem>,
    /// ISO-8601 timestamp of order creation
    pub created_at: Option<String>,
}

/// Send a kitchen ticket to a hardware KDS display via serial port.
///
/// Most hardware KDS systems (VFD displays, kitchen printers) accept plain ASCII
/// lines terminated by CRLF. This command formats the ticket as text and writes
/// it to the serial port.
#[tauri::command]
pub fn send_kds_ticket_serial(
    port_name: String,
    baud_rate: u32,
    ticket: KdsTicket,
) -> Result<String, String> {
    let port_name = port_name.trim().to_string();
    if port_name.is_empty() {
        return Err("Serial port name is required".to_string());
    }

    let baud = if baud_rate == 0 { 9600 } else { baud_rate };
    let data = format_ticket_text(&ticket);

    let mut port = serialport::new(&port_name, baud)
        .timeout(Duration::from_millis(3000))
        .open()
        .map_err(|e| format!("Cannot open KDS serial port '{}': {}", port_name, e))?;

    port.write_all(data.as_bytes())
        .map_err(|e| format!("KDS serial write failed: {}", e))?;

    Ok(format!("KDS ticket sent to serial {}", port_name))
}

/// Send a kitchen ticket to a hardware KDS display over TCP (network-connected KDS).
///
/// Sends plain text ticket data terminated with form feed (0x0C) so the KDS
/// display knows to render the ticket. Standard for network KDS hardware.
#[tauri::command]
pub fn send_kds_ticket_network(
    host: String,
    port: u16,
    ticket: KdsTicket,
) -> Result<String, String> {
    let host = host.trim().to_string();
    if host.is_empty() {
        return Err("KDS host is required".to_string());
    }
    if port == 0 {
        return Err("KDS port must be non-zero".to_string());
    }

    let addr = format!("{}:{}", host, port);
    let mut data = format_ticket_text(&ticket);
    data.push('\x0C'); // form feed — signals end of ticket to most KDS hardware

    let mut stream = std::net::TcpStream::connect_timeout(
        &addr
            .parse()
            .map_err(|_| format!("Invalid KDS address: {}", addr))?,
        Duration::from_secs(5),
    )
    .map_err(|e| format!("Cannot connect to KDS {}: {}", addr, e))?;

    stream
        .write_all(data.as_bytes())
        .map_err(|e| format!("KDS network write failed: {}", e))?;

    Ok(format!("KDS ticket sent to {}", addr))
}

// ── Ticket formatter ──────────────────────────────────────────────────────────

const KDS_WIDTH: usize = 40;

fn format_ticket_text(ticket: &KdsTicket) -> String {
    let mut out = String::with_capacity(512);

    // Divider
    out.push_str(&"=".repeat(KDS_WIDTH));
    out.push_str("\r\n");

    // Header: order number + station
    let station = ticket.station.as_deref().unwrap_or("KITCHEN");
    let order_label = format!("ORDER #{}", ticket.order_number);
    let gap = KDS_WIDTH.saturating_sub(order_label.len() + station.len());
    out.push_str(&format!("{}{}{}\r\n", order_label, " ".repeat(gap), station));

    // Order type + table
    if let Some(ref ot) = ticket.order_type {
        let table = ticket.table_number.as_deref().unwrap_or("");
        let line = if table.is_empty() {
            ot.clone()
        } else {
            format!("{} | TABLE {}", ot, table)
        };
        out.push_str(&line);
        out.push_str("\r\n");
    }

    // Timestamp
    if let Some(ref ts) = ticket.created_at {
        // Show only the time portion (HH:MM) from ISO timestamp
        let time_part = ts.find('T').map(|i| &ts[i + 1..i + 6]).unwrap_or("");
        if !time_part.is_empty() {
            out.push_str(&format!("Time: {}\r\n", time_part));
        }
    }

    out.push_str(&"-".repeat(KDS_WIDTH));
    out.push_str("\r\n");

    // Items
    for item in &ticket.items {
        let qty_name = format!("{:.0}x {}", item.quantity, item.name);
        out.push_str(&qty_name);
        out.push_str("\r\n");

        for m in &item.modifiers {
            out.push_str(&format!("   + {}\r\n", m));
        }

        if let Some(ref notes) = item.notes {
            if !notes.is_empty() {
                out.push_str(&format!("   * {}\r\n", notes));
            }
        }
    }

    out.push_str(&"=".repeat(KDS_WIDTH));
    out.push_str("\r\n");

    out
}

// ── Tests ─────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    fn sample_ticket() -> KdsTicket {
        KdsTicket {
            order_number: "42".to_string(),
            station: Some("GRILL".to_string()),
            order_type: Some("Dine In".to_string()),
            table_number: Some("7".to_string()),
            created_at: Some("2026-05-13T14:30:00Z".to_string()),
            items: vec![
                KdsItem {
                    name: "Burger".to_string(),
                    quantity: 2.0,
                    notes: Some("No pickles".to_string()),
                    modifiers: vec!["Extra cheese".to_string()],
                },
                KdsItem {
                    name: "Fries".to_string(),
                    quantity: 1.0,
                    notes: None,
                    modifiers: vec![],
                },
            ],
        }
    }

    #[test]
    fn ticket_contains_order_number() {
        let text = format_ticket_text(&sample_ticket());
        assert!(text.contains("ORDER #42"));
    }

    #[test]
    fn ticket_contains_station() {
        let text = format_ticket_text(&sample_ticket());
        assert!(text.contains("GRILL"));
    }

    #[test]
    fn ticket_contains_item_names() {
        let text = format_ticket_text(&sample_ticket());
        assert!(text.contains("Burger"));
        assert!(text.contains("Fries"));
    }

    #[test]
    fn ticket_contains_modifier() {
        let text = format_ticket_text(&sample_ticket());
        assert!(text.contains("Extra cheese"));
    }

    #[test]
    fn ticket_contains_notes() {
        let text = format_ticket_text(&sample_ticket());
        assert!(text.contains("No pickles"));
    }

    #[test]
    fn ticket_contains_table() {
        let text = format_ticket_text(&sample_ticket());
        assert!(text.contains("TABLE 7"));
    }
}
