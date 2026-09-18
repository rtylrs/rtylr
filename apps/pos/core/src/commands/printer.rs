//! Legacy Windows-only drawer kick via the print spooler. Kept for
//! backwards compatibility with frontends that haven't migrated to the
//! cross-platform `open_cash_drawer_*` commands.

use crate::hardware::escpos::DRAWER_PIN2;
use crate::transport::windows_spooler;

#[tauri::command(async)]
pub fn open_cash_drawer(printer_name: String) -> Result<String, String> {
    let printer_name = printer_name.trim();
    if printer_name.is_empty() {
        return Err("Printer name is required".to_string());
    }
    windows_spooler::write_raw(printer_name, DRAWER_PIN2, "Open Drawer")?;
    Ok("Drawer opened successfully!".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_blank_printer_name() {
        assert!(open_cash_drawer("   ".to_string()).is_err());
    }
}
