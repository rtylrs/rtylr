//! Pure ESC/POS receipt-byte builder.
//!
//! All I/O lives in `crate::transport`; this module is `no_std`-compatible in
//! spirit - just inputs in, bytes out, no side effects.

use serde::Deserialize;

use crate::hardware::escpos::{
    ALIGN_CENTER, ALIGN_LEFT, BOLD_OFF, BOLD_ON, CUT, DRAWER_PIN2, DRAWER_PIN5, FEED_4, INIT, LF,
    SIZE_DOUBLE, SIZE_NORMAL,
};
use crate::util::fmt::{write_padded_pair, write_row, RECEIPT_WIDTH};
use crate::util::time::format_timestamp;

#[derive(Deserialize)]
pub struct ReceiptItem {
    pub name: String,
    pub quantity: f64,
    pub unit_price: f64,
    pub total: f64,
    pub modifiers: Vec<String>,
    pub notes: Option<String>,
}

#[derive(Deserialize)]
pub struct ReceiptData {
    pub business_name: String,
    pub business_address: Option<String>,
    pub business_phone: Option<String>,
    pub order_number: String,
    pub order_type: Option<String>,
    pub cashier_name: Option<String>,
    pub table_number: Option<String>,
    pub items: Vec<ReceiptItem>,
    pub subtotal: f64,
    pub tax_amount: f64,
    pub discount_amount: f64,
    pub tip_amount: f64,
    pub total: f64,
    pub payment_method: String,
    pub amount_paid: f64,
    pub change_due: f64,
    pub footer_message: Option<String>,
    /// If true, send drawer kick immediately before the cut.
    pub open_drawer: bool,
    /// Which pin to kick: 2 or 5.
    pub drawer_pin: Option<u8>,
}

pub fn build_receipt(data: &ReceiptData) -> Vec<u8> {
    let mut buf: Vec<u8> = Vec::with_capacity(2048);

    let push = |buf: &mut Vec<u8>, bytes: &[u8]| buf.extend_from_slice(bytes);
    let text = |buf: &mut Vec<u8>, s: &str| buf.extend_from_slice(s.as_bytes());

    push(&mut buf, INIT);

    // Header
    push(&mut buf, ALIGN_CENTER);
    push(&mut buf, BOLD_ON);
    push(&mut buf, SIZE_DOUBLE);
    text(&mut buf, &data.business_name);
    push(&mut buf, LF);
    push(&mut buf, SIZE_NORMAL);
    push(&mut buf, BOLD_OFF);

    if let Some(ref addr) = data.business_address {
        text(&mut buf, addr);
        push(&mut buf, LF);
    }
    if let Some(ref phone) = data.business_phone {
        text(&mut buf, phone);
        push(&mut buf, LF);
    }
    push(&mut buf, LF);

    // Order meta
    push(&mut buf, ALIGN_LEFT);

    if let Some(ref ot) = data.order_type {
        write_padded_pair(&mut buf, &format!("Order #{}", data.order_number), ot, RECEIPT_WIDTH);
    } else {
        text(&mut buf, &format!("Order #{}", data.order_number));
    }
    push(&mut buf, LF);

    if let Some(ref table) = data.table_number {
        text(&mut buf, &format!("Table: {}", table));
        push(&mut buf, LF);
    }
    if let Some(ref cashier) = data.cashier_name {
        text(&mut buf, &format!("Cashier: {}", cashier));
        push(&mut buf, LF);
    }

    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    text(&mut buf, &format_timestamp(now));
    push(&mut buf, LF);

    text(&mut buf, &"-".repeat(RECEIPT_WIDTH));
    push(&mut buf, LF);

    // Items
    for item in &data.items {
        push(&mut buf, BOLD_ON);
        // Two small format! calls instead of four intermediate Strings
        // (qty_name + " ".repeat + concat) per item row.
        write_padded_pair(
            &mut buf,
            &format!("{:.0}x {}", item.quantity, item.name),
            &format!("{:.2}", item.total),
            RECEIPT_WIDTH,
        );
        push(&mut buf, BOLD_OFF);
        push(&mut buf, LF);

        if item.quantity > 1.0 {
            text(&mut buf, &format!("  @ {:.2} each", item.unit_price));
            push(&mut buf, LF);
        }

        for m in &item.modifiers {
            text(&mut buf, &format!("  + {}", m));
            push(&mut buf, LF);
        }

        if let Some(ref notes) = item.notes {
            if !notes.is_empty() {
                text(&mut buf, &format!("  * {}", notes));
                push(&mut buf, LF);
            }
        }
    }

    // Totals
    text(&mut buf, &"-".repeat(RECEIPT_WIDTH));
    push(&mut buf, LF);

    write_row(&mut buf, "Subtotal", &format!("{:.2}", data.subtotal), RECEIPT_WIDTH);

    if data.discount_amount > 0.0 {
        write_row(&mut buf, "Discount", &format!("-{:.2}", data.discount_amount), RECEIPT_WIDTH);
    }
    if data.tax_amount > 0.0 {
        write_row(&mut buf, "Tax", &format!("{:.2}", data.tax_amount), RECEIPT_WIDTH);
    }
    if data.tip_amount > 0.0 {
        write_row(&mut buf, "Tip", &format!("{:.2}", data.tip_amount), RECEIPT_WIDTH);
    }

    push(&mut buf, BOLD_ON);
    write_row(&mut buf, "TOTAL", &format!("{:.2}", data.total), RECEIPT_WIDTH);
    push(&mut buf, BOLD_OFF);

    text(&mut buf, &"-".repeat(RECEIPT_WIDTH));
    push(&mut buf, LF);

    // Payment
    write_row(
        &mut buf,
        &data.payment_method.to_uppercase(),
        &format!("{:.2}", data.amount_paid),
        RECEIPT_WIDTH,
    );

    if data.change_due > 0.0 {
        write_row(&mut buf, "Change", &format!("{:.2}", data.change_due), RECEIPT_WIDTH);
    }

    push(&mut buf, LF);

    // Footer
    push(&mut buf, ALIGN_CENTER);
    if let Some(ref footer) = data.footer_message {
        text(&mut buf, footer);
        push(&mut buf, LF);
    } else {
        text(&mut buf, "Thank you for your visit!");
        push(&mut buf, LF);
    }
    push(&mut buf, LF);

    // Drawer kick before cut so the printer finishes printing first
    if data.open_drawer {
        let pin = data.drawer_pin.unwrap_or(2);
        if pin == 5 {
            push(&mut buf, DRAWER_PIN5);
        } else {
            push(&mut buf, DRAWER_PIN2);
        }
    }

    push(&mut buf, FEED_4);
    push(&mut buf, CUT);

    buf
}

#[cfg(test)]
mod tests {
    use super::*;

    fn sample_receipt() -> ReceiptData {
        ReceiptData {
            business_name: "Test Cafe".to_string(),
            business_address: Some("123 Main St".to_string()),
            business_phone: Some("555-1234".to_string()),
            order_number: "ORD-001".to_string(),
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

    #[test] fn build_receipt_is_non_empty() { assert!(!build_receipt(&sample_receipt()).is_empty()); }
    #[test] fn build_receipt_starts_with_init() { assert!(build_receipt(&sample_receipt()).starts_with(INIT)); }
    #[test] fn build_receipt_ends_with_cut() { assert!(build_receipt(&sample_receipt()).ends_with(CUT)); }

    #[test]
    fn build_receipt_contains_business_name() {
        let t = String::from_utf8_lossy(&build_receipt(&sample_receipt())).to_string();
        assert!(t.contains("Test Cafe"));
    }

    #[test]
    fn build_receipt_contains_item_name() {
        let t = String::from_utf8_lossy(&build_receipt(&sample_receipt())).to_string();
        assert!(t.contains("Espresso"));
    }

    #[test]
    fn build_receipt_includes_drawer_kick_when_enabled() {
        let mut r = sample_receipt();
        r.open_drawer = true;
        r.drawer_pin = Some(2);
        let data = build_receipt(&r);
        assert!(data.windows(5).any(|w| w == [0x1B, 0x70, 0x00, 0x19, 0xFA]));
    }

    #[test]
    fn build_receipt_no_drawer_kick_when_disabled() {
        let data = build_receipt(&sample_receipt());
        assert!(!data.windows(5).any(|w| w == [0x1B, 0x70, 0x00, 0x19, 0xFA]));
    }
}
