//! Receipt-level invariants.
//!
//! Verifies that the receipt the frontend hands us is internally
//! consistent before we waste paper printing it. Catches frontend bugs
//! like negative tax, mismatched totals, and overflow-priced items.

use crate::hardware::receipt_builder::{ReceiptData, ReceiptItem};
use crate::validation::drawer::validate_drawer_pin;
use crate::validation::money::{equal_within_cents, validate_amount, validate_quantity};
use crate::validation::text::{validate_optional, validate_printable};

pub const MAX_ITEMS_PER_RECEIPT: usize = 500;
pub const MAX_MODIFIERS_PER_ITEM: usize = 64;
pub const MAX_BUSINESS_NAME_LEN: usize = 80;
pub const MAX_NAME_LEN: usize = 120;
pub const MAX_FREE_TEXT_LEN: usize = 200;

pub fn validate_receipt(data: &ReceiptData) -> Result<(), String> {
    // Header strings
    validate_printable("business_name", &data.business_name, MAX_BUSINESS_NAME_LEN)?;
    if data.business_name.trim().is_empty() {
        return Err("business_name is required".to_string());
    }
    validate_optional("business_address", &data.business_address, MAX_FREE_TEXT_LEN)?;
    validate_optional("business_phone", &data.business_phone, 40)?;
    validate_printable("order_number", &data.order_number, 64)?;
    validate_optional("order_type", &data.order_type, 40)?;
    validate_optional("cashier_name", &data.cashier_name, 80)?;
    validate_optional("table_number", &data.table_number, 20)?;
    validate_optional("footer_message", &data.footer_message, MAX_FREE_TEXT_LEN)?;
    validate_printable("payment_method", &data.payment_method, 40)?;

    // Item invariants
    if data.items.is_empty() {
        return Err("Receipt must contain at least one item".to_string());
    }
    if data.items.len() > MAX_ITEMS_PER_RECEIPT {
        return Err(format!(
            "Too many items ({} > {})",
            data.items.len(),
            MAX_ITEMS_PER_RECEIPT
        ));
    }
    for (i, item) in data.items.iter().enumerate() {
        validate_item(i, item)?;
    }

    // Totals
    validate_amount("subtotal", data.subtotal)?;
    validate_amount("tax_amount", data.tax_amount)?;
    validate_amount("discount_amount", data.discount_amount)?;
    validate_amount("tip_amount", data.tip_amount)?;
    validate_amount("total", data.total)?;
    validate_amount("amount_paid", data.amount_paid)?;
    validate_amount("change_due", data.change_due)?;

    // Totals math: subtotal - discount + tax + tip == total
    let expected_total = data.subtotal - data.discount_amount + data.tax_amount + data.tip_amount;
    if !equal_within_cents(expected_total, data.total) {
        return Err(format!(
            "Receipt totals do not reconcile: subtotal({:.2}) - discount({:.2}) + tax({:.2}) + tip({:.2}) = {:.2}, but total is {:.2}",
            data.subtotal, data.discount_amount, data.tax_amount, data.tip_amount, expected_total, data.total
        ));
    }

    // Discount cannot exceed subtotal
    if data.discount_amount > data.subtotal + 0.005 {
        return Err(format!(
            "discount_amount ({:.2}) exceeds subtotal ({:.2})",
            data.discount_amount, data.subtotal
        ));
    }

    // Payment must cover total
    if data.amount_paid + 0.005 < data.total {
        return Err(format!(
            "amount_paid ({:.2}) is less than total ({:.2})",
            data.amount_paid, data.total
        ));
    }

    // Change must equal amount_paid - total
    let expected_change = (data.amount_paid - data.total).max(0.0);
    if !equal_within_cents(expected_change, data.change_due) {
        return Err(format!(
            "change_due ({:.2}) does not match amount_paid({:.2}) - total({:.2}) = {:.2}",
            data.change_due, data.amount_paid, data.total, expected_change
        ));
    }

    // Drawer pin - only validate if a kick was requested.
    if data.open_drawer {
        let pin = data.drawer_pin.unwrap_or(2);
        validate_drawer_pin(pin)?;
    }

    Ok(())
}

fn validate_item(index: usize, item: &ReceiptItem) -> Result<(), String> {
    validate_printable(&format!("items[{}].name", index), &item.name, MAX_NAME_LEN)?;
    if item.name.trim().is_empty() {
        return Err(format!("items[{}].name is empty", index));
    }

    validate_quantity(item.quantity).map_err(|e| format!("items[{}]: {}", index, e))?;
    validate_amount(&format!("items[{}].unit_price", index), item.unit_price)?;
    validate_amount(&format!("items[{}].total", index), item.total)?;

    // quantity * unit_price ≈ total (allow cents-level rounding error)
    let expected = item.quantity * item.unit_price;
    if !equal_within_cents(expected, item.total) {
        return Err(format!(
            "items[{}] total ({:.2}) does not match qty({}) * unit_price({:.2}) = {:.2}",
            index, item.total, item.quantity, item.unit_price, expected
        ));
    }

    if item.modifiers.len() > MAX_MODIFIERS_PER_ITEM {
        return Err(format!(
            "items[{}] has too many modifiers ({} > {})",
            index,
            item.modifiers.len(),
            MAX_MODIFIERS_PER_ITEM
        ));
    }
    for (mi, m) in item.modifiers.iter().enumerate() {
        validate_printable(&format!("items[{}].modifiers[{}]", index, mi), m, MAX_FREE_TEXT_LEN)?;
    }

    if let Some(ref notes) = item.notes {
        validate_printable(&format!("items[{}].notes", index), notes, MAX_FREE_TEXT_LEN)?;
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn good_receipt() -> ReceiptData {
        ReceiptData {
            business_name: "Test Cafe".to_string(),
            business_address: Some("123 Main".to_string()),
            business_phone: Some("555-1234".to_string()),
            order_number: "ORD-1".to_string(),
            order_type: Some("Dine In".to_string()),
            cashier_name: Some("Alice".to_string()),
            table_number: Some("5".to_string()),
            items: vec![ReceiptItem {
                name: "Coffee".to_string(),
                quantity: 2.0,
                unit_price: 3.50,
                total: 7.00,
                modifiers: vec![],
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
            footer_message: None,
            open_drawer: false,
            drawer_pin: None,
        }
    }

    #[test] fn good_receipt_validates() { assert!(validate_receipt(&good_receipt()).is_ok()); }

    #[test]
    fn rejects_empty_business_name() {
        let mut r = good_receipt();
        r.business_name = "   ".to_string();
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_empty_items() {
        let mut r = good_receipt();
        r.items.clear();
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_mismatched_totals() {
        let mut r = good_receipt();
        r.total = 99.99;
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_underpayment() {
        let mut r = good_receipt();
        r.amount_paid = 5.00;
        r.change_due = 0.0;
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_wrong_change() {
        let mut r = good_receipt();
        r.change_due = 5.00;
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_discount_exceeding_subtotal() {
        let mut r = good_receipt();
        r.discount_amount = 99.00;
        // recompute total to satisfy the totals invariant - but discount > subtotal should still trip
        r.total = r.subtotal - r.discount_amount + r.tax_amount + r.tip_amount;
        r.amount_paid = r.total.max(0.0);
        r.change_due = (r.amount_paid - r.total).max(0.0);
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_negative_amount() {
        let mut r = good_receipt();
        r.tip_amount = -1.0;
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_nan_amount() {
        let mut r = good_receipt();
        r.tax_amount = f64::NAN;
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_item_total_mismatch() {
        let mut r = good_receipt();
        r.items[0].total = 99.99;
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_item_with_escape_in_name() {
        let mut r = good_receipt();
        r.items[0].name = "Coffee\x1Bevil".to_string();
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn rejects_invalid_drawer_pin_when_kick_requested() {
        let mut r = good_receipt();
        r.open_drawer = true;
        r.drawer_pin = Some(3);
        assert!(validate_receipt(&r).is_err());
    }

    #[test]
    fn accepts_default_drawer_pin_when_kick_requested() {
        let mut r = good_receipt();
        r.open_drawer = true;
        r.drawer_pin = None;
        assert!(validate_receipt(&r).is_ok());
    }

    #[test]
    fn rejects_too_many_items() {
        let mut r = good_receipt();
        r.items = (0..501)
            .map(|i| ReceiptItem {
                name: format!("Item {}", i),
                quantity: 1.0,
                unit_price: 0.0,
                total: 0.0,
                modifiers: vec![],
                notes: None,
            })
            .collect();
        // Make totals reconcile so we know we're tripping the item-count guard
        r.subtotal = 0.0;
        r.total = r.tax_amount + r.tip_amount;
        r.amount_paid = r.total;
        r.change_due = 0.0;
        assert!(validate_receipt(&r).is_err());
    }
}
