//! Windows print-spooler (winspool) raw-data transport.
//!
//! On non-Windows platforms every function returns an error so the
//! compile unit is identical across targets.

#[allow(unused_variables)]
pub fn write_raw(printer_name: &str, data: &[u8], doc_name: &str) -> Result<(), String> {
    #[cfg(windows)]
    {
        use std::ffi::OsStr;
        use std::os::windows::ffi::OsStrExt;
        use winapi::um::winspool::{
            ClosePrinter, EndDocPrinter, EndPagePrinter, OpenPrinterW, StartDocPrinterW,
            StartPagePrinter, WritePrinter, DOC_INFO_1W,
        };

        let wide_name: Vec<u16> = OsStr::new(printer_name)
            .encode_wide()
            .chain(std::iter::once(0))
            .collect();

        unsafe {
            let mut h_printer = std::ptr::null_mut();
            if OpenPrinterW(wide_name.as_ptr() as *mut u16, &mut h_printer, std::ptr::null_mut()) == 0 {
                return Err(format!("Cannot open printer '{}'", printer_name));
            }

            let mut doc_name_w: Vec<u16> = OsStr::new(doc_name)
                .encode_wide()
                .chain(std::iter::once(0))
                .collect();
            let mut data_type: Vec<u16> = OsStr::new("RAW")
                .encode_wide()
                .chain(std::iter::once(0))
                .collect();

            let doc_info = DOC_INFO_1W {
                pDocName: doc_name_w.as_mut_ptr(),
                pOutputFile: std::ptr::null_mut(),
                pDatatype: data_type.as_mut_ptr(),
            };

            let doc_id = StartDocPrinterW(h_printer, 1, &doc_info as *const _ as *mut u8);
            if doc_id == 0 {
                ClosePrinter(h_printer);
                return Err("StartDocPrinter failed".to_string());
            }

            if StartPagePrinter(h_printer) == 0 {
                EndDocPrinter(h_printer);
                ClosePrinter(h_printer);
                return Err("StartPagePrinter failed".to_string());
            }

            let mut bytes_written = 0u32;
            if WritePrinter(
                h_printer,
                data.as_ptr() as *mut _,
                data.len() as u32,
                &mut bytes_written,
            ) == 0
            {
                EndPagePrinter(h_printer);
                EndDocPrinter(h_printer);
                ClosePrinter(h_printer);
                return Err("WritePrinter failed".to_string());
            }

            EndPagePrinter(h_printer);
            EndDocPrinter(h_printer);
            ClosePrinter(h_printer);
        }

        Ok(())
    }

    #[cfg(not(windows))]
    {
        Err("Windows spooler transport is only available on Windows. Use 'serial' or 'network'.".to_string())
    }
}

/// Open + close the printer handle (no writes) - used by health probes.
#[allow(unused_variables)]
pub fn probe_open(printer_name: &str) -> Result<(), String> {
    #[cfg(windows)]
    {
        use std::ffi::OsStr;
        use std::os::windows::ffi::OsStrExt;
        use winapi::um::winspool::{ClosePrinter, OpenPrinterW};

        let wide: Vec<u16> = OsStr::new(printer_name)
            .encode_wide()
            .chain(std::iter::once(0))
            .collect();

        unsafe {
            let mut h = std::ptr::null_mut();
            if OpenPrinterW(wide.as_ptr() as *mut u16, &mut h, std::ptr::null_mut()) != 0 {
                ClosePrinter(h);
                Ok(())
            } else {
                Err(format!("Cannot open Windows printer '{}'", printer_name))
            }
        }
    }

    #[cfg(not(windows))]
    {
        Err("Windows spooler not available on this platform".to_string())
    }
}
