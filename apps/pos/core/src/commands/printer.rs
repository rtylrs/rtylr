#[tauri::command]
pub fn open_cash_drawer(printer_name: String) -> Result<String, String> {
    let printer_name = printer_name.trim();
    if printer_name.is_empty() {
        return Err("Printer name is required".to_string());
    }

    #[cfg(not(windows))]
    let _ = &printer_name;

    #[cfg(windows)]
    {
        use std::ffi::OsStr;
        use std::os::windows::ffi::OsStrExt;
        use std::ptr;
        use winapi::um::winspool::{
            ClosePrinter, EndDocPrinter, EndPagePrinter, OpenPrinterW,
            StartDocPrinterW, StartPagePrinter, WritePrinter, DOC_INFO_1W,
        };

        // ESC p 0 25 250  →  [1B 70 00 19 FA]
        let drawer_pulse: [u8; 5] = [0x1B, 0x70, 0x00, 0x19, 0xFA];

        // Convert printer name to wide string (UTF-16)
        let wide_name: Vec<u16> = OsStr::new(&printer_name)
            .encode_wide()
            .chain(std::iter::once(0))
            .collect();

        unsafe {
            let mut h_printer = ptr::null_mut();

            // Open printer
            if OpenPrinterW(
                wide_name.as_ptr() as *mut u16,
                &mut h_printer,
                ptr::null_mut(),
            ) == 0
            {
                return Err(format!("Failed to open printer '{}'", printer_name));
            }

            // Prepare DOC_INFO
            let mut doc_name: Vec<u16> = OsStr::new("Open Drawer")
                .encode_wide()
                .chain(std::iter::once(0))
                .collect();

            let mut data_type: Vec<u16> = OsStr::new("RAW")
                .encode_wide()
                .chain(std::iter::once(0))
                .collect();

            let doc_info = DOC_INFO_1W {
                pDocName: doc_name.as_mut_ptr(),
                pOutputFile: ptr::null_mut(),
                pDatatype: data_type.as_mut_ptr(),
            };

            // Start document
            let doc_id = StartDocPrinterW(h_printer, 1, &doc_info as *const _ as *mut u8);
            if doc_id == 0 {
                ClosePrinter(h_printer);
                return Err("Failed to start document".to_string());
            }

            // Start page
            if StartPagePrinter(h_printer) == 0 {
                EndDocPrinter(h_printer);
                ClosePrinter(h_printer);
                return Err("Failed to start page".to_string());
            }

            // Write data
            let mut bytes_written = 0u32;
            if WritePrinter(
                h_printer,
                drawer_pulse.as_ptr() as *mut _,
                drawer_pulse.len() as u32,
                &mut bytes_written,
            ) == 0
            {
                EndPagePrinter(h_printer);
                EndDocPrinter(h_printer);
                ClosePrinter(h_printer);
                return Err("Failed to write to printer".to_string());
            }

            // End page and document
            EndPagePrinter(h_printer);
            EndDocPrinter(h_printer);
            ClosePrinter(h_printer);
        }

        Ok("Drawer opened successfully!".to_string())
    }

    #[cfg(not(windows))]
    {
        Err("This function is only available on Windows".to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::open_cash_drawer;

    #[test]
    fn rejects_blank_printer_name() {
        let result = open_cash_drawer("   ".to_string());

        assert_eq!(result, Err("Printer name is required".to_string()));
    }
}
