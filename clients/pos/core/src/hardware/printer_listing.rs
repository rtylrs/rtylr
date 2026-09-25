//! OS-specific enumeration of installed printers.

use serde::Serialize;

#[derive(Serialize)]
pub struct PrinterInfo {
    pub name: String,
    pub is_default: bool,
    pub connection_type: String,
}

pub fn list_printers() -> Result<Vec<PrinterInfo>, String> {
    #[cfg(windows)]
    { list_printers_windows() }

    #[cfg(unix)]
    { list_printers_unix() }
}

#[cfg(windows)]
fn list_printers_windows() -> Result<Vec<PrinterInfo>, String> {
    use std::ffi::OsString;
    use std::os::windows::ffi::OsStringExt;
    use winapi::um::winspool::{
        EnumPrintersW, GetDefaultPrinterW, PRINTER_ENUM_LOCAL, PRINTER_ENUM_CONNECTIONS,
        PRINTER_INFO_2W,
    };

    let flags = PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS;
    let mut bytes_needed: u32 = 0;
    let mut printers_returned: u32 = 0;

    unsafe {
        EnumPrintersW(
            flags,
            std::ptr::null_mut(),
            2,
            std::ptr::null_mut(),
            0,
            &mut bytes_needed,
            &mut printers_returned,
        );

        if bytes_needed == 0 {
            return Ok(vec![]);
        }

        let mut buffer: Vec<u8> = vec![0u8; bytes_needed as usize];

        let ok = EnumPrintersW(
            flags,
            std::ptr::null_mut(),
            2,
            buffer.as_mut_ptr(),
            bytes_needed,
            &mut bytes_needed,
            &mut printers_returned,
        );

        if ok == 0 {
            return Err("EnumPrintersW failed".to_string());
        }

        let mut default_buf = vec![0u16; 256];
        let mut default_len = default_buf.len() as u32;
        GetDefaultPrinterW(default_buf.as_mut_ptr(), &mut default_len);
        let default_name = OsString::from_wide(
            &default_buf[..default_buf.iter().position(|&c| c == 0).unwrap_or(0)],
        )
        .to_string_lossy()
        .to_string();

        let info_ptr = buffer.as_ptr() as *const PRINTER_INFO_2W;
        let mut printers = Vec::with_capacity(printers_returned as usize);

        for i in 0..printers_returned as isize {
            let info = &*info_ptr.offset(i);

            let name = if info.pPrinterName.is_null() {
                continue;
            } else {
                let len = (0..).take_while(|&j| *info.pPrinterName.offset(j) != 0).count();
                let slice = std::slice::from_raw_parts(info.pPrinterName, len);
                OsString::from_wide(slice).to_string_lossy().to_string()
            };

            let port = if info.pPortName.is_null() {
                String::new()
            } else {
                let len = (0..).take_while(|&j| *info.pPortName.offset(j) != 0).count();
                let slice = std::slice::from_raw_parts(info.pPortName, len);
                OsString::from_wide(slice).to_string_lossy().to_string()
            };

            let connection_type = infer_connection_type_windows(&port);

            printers.push(PrinterInfo {
                is_default: name == default_name,
                connection_type,
                name,
            });
        }

        Ok(printers)
    }
}

#[cfg(windows)]
fn infer_connection_type_windows(port: &str) -> String {
    let p = port.to_uppercase();
    if p.starts_with("USB") {
        "USB".to_string()
    } else if p.starts_with("COM") {
        "Serial".to_string()
    } else if p.starts_with("LPT") {
        "Parallel".to_string()
    } else if p.contains("IP_") || p.contains("NET") || p.contains("WSD") || p.contains(':') {
        "Network".to_string()
    } else {
        "Unknown".to_string()
    }
}

#[cfg(unix)]
fn list_printers_unix() -> Result<Vec<PrinterInfo>, String> {
    use std::process::Command;

    let output = Command::new("lpstat")
        .args(["-a"])
        .output()
        .map_err(|e| format!("lpstat not available: {}", e))?;

    if !output.status.success() && output.stdout.is_empty() {
        return Ok(vec![]);
    }

    let default_output = Command::new("lpstat").args(["-d"]).output().ok();
    let default_name: String = default_output
        .and_then(|o| String::from_utf8(o.stdout).ok())
        .and_then(|s| s.trim().split_whitespace().last().map(|s| s.to_string()))
        .unwrap_or_default();

    let stdout = String::from_utf8(output.stdout)
        .map_err(|_| "lpstat output is not valid UTF-8".to_string())?;

    let printers = stdout
        .lines()
        .filter_map(|line| {
            let name = line.split_whitespace().next()?;
            if name.is_empty() {
                return None;
            }
            let connection_type = infer_connection_type_unix(name);
            Some(PrinterInfo {
                is_default: name == default_name,
                connection_type,
                name: name.to_string(),
            })
        })
        .collect();

    Ok(printers)
}

#[cfg(unix)]
fn infer_connection_type_unix(name: &str) -> String {
    let n = name.to_lowercase();
    if n.contains("usb") {
        "USB".to_string()
    } else if n.contains("net") || n.contains("network") || n.contains("ip") {
        "Network".to_string()
    } else if n.contains("serial") || n.contains("tty") {
        "Serial".to_string()
    } else {
        "Unknown".to_string()
    }
}
