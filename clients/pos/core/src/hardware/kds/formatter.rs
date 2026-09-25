//! Pure KDS ticket-text formatter.

use serde::Deserialize;

#[derive(Deserialize, Clone, Debug, PartialEq)]
pub struct KdsItem {
    pub name: String,
    pub quantity: f64,
    pub notes: Option<String>,
    pub modifiers: Vec<String>,
}

#[derive(Deserialize, Clone, Debug, PartialEq)]
pub struct KdsTicket {
    pub order_number: String,
    pub station: Option<String>,
    pub order_type: Option<String>,
    pub table_number: Option<String>,
    pub items: Vec<KdsItem>,
    /// ISO-8601 timestamp of order creation
    pub created_at: Option<String>,
}

pub const KDS_WIDTH: usize = 40;

pub fn format_ticket_text(ticket: &KdsTicket) -> String {
    let mut out = String::with_capacity(512);

    out.push_str(&"=".repeat(KDS_WIDTH));
    out.push_str("\r\n");

    let station = ticket.station.as_deref().unwrap_or("KITCHEN");
    let order_label = format!("ORDER #{}", ticket.order_number);
    let gap = KDS_WIDTH.saturating_sub(order_label.len() + station.len());
    out.push_str(&format!("{}{}{}\r\n", order_label, " ".repeat(gap), station));

    if let Some(ref ot) = ticket.order_type {
        let table = ticket.table_number.as_deref().unwrap_or("");
        let line = if table.is_empty() {
            ot.to_string()
        } else {
            format!("{} | TABLE {}", ot, table)
        };
        out.push_str(&line);
        out.push_str("\r\n");
    }

    if let Some(ref ts) = ticket.created_at {
        // Slice defensively: a short or non-ASCII timestamp from the frontend
        // must not panic the command (slicing could land mid-UTF-8 char).
        let time_part = ts
            .find('T')
            .and_then(|i| ts.get(i + 1..(i + 6).min(ts.len())))
            .unwrap_or("");
        if !time_part.is_empty() {
            out.push_str(&format!("Time: {}\r\n", time_part));
        }
    }

    out.push_str(&"-".repeat(KDS_WIDTH));
    out.push_str("\r\n");

    for item in &ticket.items {
        out.push_str(&format!("{:.0}x {}\r\n", item.quantity, item.name));

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

    #[test] fn ticket_contains_order_number() { assert!(format_ticket_text(&sample_ticket()).contains("ORDER #42")); }
    #[test] fn ticket_contains_station() { assert!(format_ticket_text(&sample_ticket()).contains("GRILL")); }
    #[test] fn ticket_contains_item_names() {
        let t = format_ticket_text(&sample_ticket());
        assert!(t.contains("Burger") && t.contains("Fries"));
    }
    #[test] fn ticket_contains_modifier() { assert!(format_ticket_text(&sample_ticket()).contains("Extra cheese")); }
    #[test] fn ticket_contains_notes() { assert!(format_ticket_text(&sample_ticket()).contains("No pickles")); }
    #[test] fn ticket_contains_table() { assert!(format_ticket_text(&sample_ticket()).contains("TABLE 7")); }
}
