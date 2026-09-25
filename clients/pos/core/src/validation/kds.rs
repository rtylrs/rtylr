//! KDS-ticket validation.

use crate::hardware::kds::formatter::{KdsItem, KdsTicket};
use crate::validation::money::validate_quantity;
use crate::validation::text::{validate_optional, validate_printable};

pub const MAX_KDS_ITEMS_PER_TICKET: usize = 200;
pub const MAX_KDS_MODIFIERS_PER_ITEM: usize = 64;

pub fn validate_kds_ticket(ticket: &KdsTicket) -> Result<(), String> {
    validate_printable("order_number", &ticket.order_number, 64)?;
    validate_optional("station", &ticket.station, 40)?;
    validate_optional("order_type", &ticket.order_type, 40)?;
    validate_optional("table_number", &ticket.table_number, 20)?;
    validate_optional("created_at", &ticket.created_at, 40)?;

    if ticket.items.is_empty() {
        return Err("KDS ticket must contain at least one item".to_string());
    }
    if ticket.items.len() > MAX_KDS_ITEMS_PER_TICKET {
        return Err(format!(
            "Too many items on KDS ticket ({} > {})",
            ticket.items.len(),
            MAX_KDS_ITEMS_PER_TICKET
        ));
    }
    for (i, item) in ticket.items.iter().enumerate() {
        validate_kds_item(i, item)?;
    }
    Ok(())
}

fn validate_kds_item(index: usize, item: &KdsItem) -> Result<(), String> {
    validate_printable(&format!("items[{}].name", index), &item.name, 120)?;
    if item.name.trim().is_empty() {
        return Err(format!("items[{}].name is empty", index));
    }
    validate_quantity(item.quantity).map_err(|e| format!("items[{}]: {}", index, e))?;

    if item.modifiers.len() > MAX_KDS_MODIFIERS_PER_ITEM {
        return Err(format!(
            "items[{}] has too many modifiers",
            index
        ));
    }
    for (mi, m) in item.modifiers.iter().enumerate() {
        validate_printable(&format!("items[{}].modifiers[{}]", index, mi), m, 200)?;
    }
    if let Some(ref notes) = item.notes {
        validate_printable(&format!("items[{}].notes", index), notes, 200)?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn good_ticket() -> KdsTicket {
        KdsTicket {
            order_number: "42".to_string(),
            station: Some("GRILL".to_string()),
            order_type: Some("Dine In".to_string()),
            table_number: Some("7".to_string()),
            created_at: Some("2026-05-13T14:30:00Z".to_string()),
            items: vec![KdsItem {
                name: "Burger".to_string(),
                quantity: 1.0,
                notes: None,
                modifiers: vec![],
            }],
        }
    }

    #[test] fn good_ticket_validates() { assert!(validate_kds_ticket(&good_ticket()).is_ok()); }

    #[test]
    fn rejects_no_items() {
        let mut t = good_ticket();
        t.items.clear();
        assert!(validate_kds_ticket(&t).is_err());
    }

    #[test]
    fn rejects_form_feed_in_item_name() {
        let mut t = good_ticket();
        t.items[0].name = "Bur\x0Cger".to_string();
        // Form feed in body would prematurely end the KDS ticket
        assert!(validate_kds_ticket(&t).is_err());
    }

    #[test]
    fn rejects_zero_quantity() {
        let mut t = good_ticket();
        t.items[0].quantity = 0.0;
        assert!(validate_kds_ticket(&t).is_err());
    }
}
