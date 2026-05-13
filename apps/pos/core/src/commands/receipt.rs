use serde::Deserialize;
use std::io::Write as IoWrite;
use std::time::Duration;

// ── ESC/POS constants ────────────────────────────────────────────────────────

const INIT: &[u8] = &[0x1B, 0x40];
const CUT: &[u8] = &[0x1D, 0x56, 0x41, 0x03]; // partial cut
const FEED_4: &[u8] = &[0x1B, 0x64, 0x04]; // feed 4 lines
const LF: &[u8] = &[0x0A];
const BOLD_ON: &[u8] = &[0x1B, 0x45, 0x01];
const BOLD_OFF: &[u8] = &[0x1B, 0x45, 0x00];
const ALIGN_LEFT: &[u8] = &[0x1B, 0x61, 0x00];
const ALIGN_CENTER: &[u8] = &[0x1B, 0x61, 0x01];
#[allow(dead_code)]
const ALIGN_RIGHT: &[u8] = &[0x1B, 0x61, 0x02]; // available for future KDS / report printing
const SIZE_DOUBLE: &[u8] = &[0x1B, 0x21, 0x30]; // double height + width
const SIZE_NORMAL: &[u8] = &[0x1B, 0x21, 0x00];
const DRAWER_PIN2: &[u8] = &[0x1B, 0x70, 0x00, 0x19, 0xFA];
const DRAWER_PIN5: &[u8] = &[0x1B, 0x70, 0x01, 0x19, 0xFA];

const RECEIPT_WIDTH: usize = 42;

// ── Domain types ─────────────────────────────────────────────────────────────

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

#[derive(Deserialize)]
pub struct PrinterConfig {
    /// "windows", "serial", or "network"
    pub transport: String,
    /// Windows: printer name / Serial: port name / Network: host
    pub target: String,
    /// Serial: baud rate (default 9600) / Network: TCP port (default 9100)
    pub port: Option<u32>,
}

// ── ESC/POS builder ──────────────────────────────────────────────────────────

fn build_receipt(data: &ReceiptData) -> Vec<u8> {
    let mut buf: Vec<u8> = Vec::with_capacity(2048);

    let push = |buf: &mut Vec<u8>, bytes: &[u8]| buf.extend_from_slice(bytes);
    let text = |buf: &mut Vec<u8>, s: &str| buf.extend_from_slice(s.as_bytes());

    push(&mut buf, INIT);

    // ── Header ────────────────────────────────────────────────────────────────
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

    // ── Order meta ────────────────────────────────────────────────────────────
    push(&mut buf, ALIGN_LEFT);

    let order_label = format!("Order #{}", data.order_number);
    if let Some(ref ot) = data.order_type {
        let right_pad = RECEIPT_WIDTH.saturating_sub(order_label.len() + ot.len());
        text(
            &mut buf,
            &format!("{}{}{}", order_label, " ".repeat(right_pad), ot),
        );
    } else {
        text(&mut buf, &order_label);
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

    // Local timestamp
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    text(&mut buf, &format_timestamp(now));
    push(&mut buf, LF);

    // Divider
    text(&mut buf, &"-".repeat(RECEIPT_WIDTH));
    push(&mut buf, LF);

    // ── Items ─────────────────────────────────────────────────────────────────
    for item in &data.items {
        // Qty × Name | Total
        let qty_name = format!("{:.0}x {}", item.quantity, item.name);
        let total_str = format!("{:.2}", item.total);
        let gap = RECEIPT_WIDTH.saturating_sub(qty_name.len() + total_str.len());
        push(&mut buf, BOLD_ON);
        text(
            &mut buf,
            &format!("{}{}{}", qty_name, " ".repeat(gap), total_str),
        );
        push(&mut buf, BOLD_OFF);
        push(&mut buf, LF);

        // Unit price (only if qty > 1)
        if item.quantity > 1.0 {
            text(
                &mut buf,
                &format!("  @ {:.2} each", item.unit_price),
            );
            push(&mut buf, LF);
        }

        // Modifiers
        for m in &item.modifiers {
            text(&mut buf, &format!("  + {}", m));
            push(&mut buf, LF);
        }

        // Item notes
        if let Some(ref notes) = item.notes {
            if !notes.is_empty() {
                text(&mut buf, &format!("  * {}", notes));
                push(&mut buf, LF);
            }
        }
    }

    // ── Totals ────────────────────────────────────────────────────────────────
    text(&mut buf, &"-".repeat(RECEIPT_WIDTH));
    push(&mut buf, LF);

    row(&mut buf, "Subtotal", &format!("{:.2}", data.subtotal));

    if data.discount_amount > 0.0 {
        row(
            &mut buf,
            "Discount",
            &format!("-{:.2}", data.discount_amount),
        );
    }

    if data.tax_amount > 0.0 {
        row(&mut buf, "Tax", &format!("{:.2}", data.tax_amount));
    }

    if data.tip_amount > 0.0 {
        row(&mut buf, "Tip", &format!("{:.2}", data.tip_amount));
    }

    push(&mut buf, BOLD_ON);
    row(&mut buf, "TOTAL", &format!("{:.2}", data.total));
    push(&mut buf, BOLD_OFF);

    text(&mut buf, &"-".repeat(RECEIPT_WIDTH));
    push(&mut buf, LF);

    // ── Payment ───────────────────────────────────────────────────────────────
    row(
        &mut buf,
        &data.payment_method.to_uppercase(),
        &format!("{:.2}", data.amount_paid),
    );

    if data.change_due > 0.0 {
        row(&mut buf, "Change", &format!("{:.2}", data.change_due));
    }

    push(&mut buf, LF);

    // ── Footer ────────────────────────────────────────────────────────────────
    push(&mut buf, ALIGN_CENTER);
    if let Some(ref footer) = data.footer_message {
        text(&mut buf, footer);
        push(&mut buf, LF);
    } else {
        text(&mut buf, "Thank you for your visit!");
        push(&mut buf, LF);
    }
    push(&mut buf, LF);

    // ── Drawer kick (before cut so printer finishes print first) ──────────────
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

fn row(buf: &mut Vec<u8>, label: &str, value: &str) {
    let gap = RECEIPT_WIDTH.saturating_sub(label.len() + value.len());
    buf.extend_from_slice(format!("{}{}{}", label, " ".repeat(gap), value).as_bytes());
    buf.push(0x0A);
}

fn format_timestamp(unix_secs: u64) -> String {
    // Very lightweight timestamp formatter (no chrono dep)
    // Format: YYYY-MM-DD HH:MM — approximate, UTC
    let secs = unix_secs;
    let mins = secs / 60;
    let hours = mins / 60 % 24;
    let days = secs / 86400;

    // Days since epoch → approximate date (good enough for receipt)
    let mut year = 1970u32;
    let mut remaining = days;
    loop {
        let days_in_year = if is_leap(year) { 366 } else { 365 };
        if remaining < days_in_year {
            break;
        }
        remaining -= days_in_year;
        year += 1;
    }
    let month_days: [u32; 12] = [31, if is_leap(year) { 29 } else { 28 }, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
    let mut month = 1u32;
    for &md in &month_days {
        if remaining < md as u64 {
            break;
        }
        remaining -= md as u64;
        month += 1;
    }
    let day = remaining + 1;

    format!(
        "{:04}-{:02}-{:02} {:02}:{:02}",
        year,
        month,
        day,
        hours % 24,
        mins % 60
    )
}

fn is_leap(year: u32) -> bool {
    (year % 4 == 0 && year % 100 != 0) || year % 400 == 0
}

// ── Tauri commands ────────────────────────────────────────────────────────────

/// Print a receipt. The `config` specifies transport (windows / serial / network).
#[tauri::command]
pub fn print_receipt(config: PrinterConfig, receipt: ReceiptData) -> Result<String, String> {
    let data = build_receipt(&receipt);

    match config.transport.as_str() {
        "serial" => print_via_serial(&config.target, config.port.unwrap_or(9600), &data),
        "network" => print_via_network(&config.target, config.port.unwrap_or(9100) as u16, &data),
        "windows" => print_via_windows_spooler(&config.target, &data),
        other => Err(format!("Unknown transport: '{}'", other)),
    }
}

fn print_via_serial(port_name: &str, baud: u32, data: &[u8]) -> Result<String, String> {
    let mut port = serialport::new(port_name, baud)
        .timeout(Duration::from_millis(5000))
        .open()
        .map_err(|e| format!("Cannot open serial port '{}': {}", port_name, e))?;

    port.write_all(data)
        .map_err(|e| format!("Serial write failed: {}", e))?;

    Ok(format!("Receipt sent to serial {}", port_name))
}

fn print_via_network(host: &str, port: u16, data: &[u8]) -> Result<String, String> {
    let addr = format!("{}:{}", host, port);
    let mut stream = std::net::TcpStream::connect_timeout(
        &addr
            .parse()
            .map_err(|_| format!("Invalid address: {}", addr))?,
        Duration::from_secs(8),
    )
    .map_err(|e| format!("Cannot connect to {}: {}", addr, e))?;

    stream
        .write_all(data)
        .map_err(|e| format!("Network write to {} failed: {}", addr, e))?;

    Ok(format!("Receipt sent to network {}:{}", host, port))
}

fn print_via_windows_spooler(_printer_name: &str, _data: &[u8]) -> Result<String, String> {
    #[cfg(windows)]
    {
        use std::ffi::OsStr;
        use std::os::windows::ffi::OsStrExt;
        use winapi::um::winspool::{
            ClosePrinter, EndDocPrinter, EndPagePrinter, OpenPrinterW, StartDocPrinterW,
            StartPagePrinter, WritePrinter, DOC_INFO_1W,
        };

        let wide_name: Vec<u16> = OsStr::new(_printer_name)
            .encode_wide()
            .chain(std::iter::once(0))
            .collect();

        unsafe {
            let mut h_printer = std::ptr::null_mut();
            if OpenPrinterW(
                wide_name.as_ptr() as *mut u16,
                &mut h_printer,
                std::ptr::null_mut(),
            ) == 0
            {
                return Err(format!("Cannot open printer '{}'", _printer_name));
            }

            let mut doc_name: Vec<u16> = OsStr::new("Receipt")
                .encode_wide()
                .chain(std::iter::once(0))
                .collect();
            let mut data_type: Vec<u16> = OsStr::new("RAW")
                .encode_wide()
                .chain(std::iter::once(0))
                .collect();

            let doc_info = DOC_INFO_1W {
                pDocName: doc_name.as_mut_ptr(),
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
                _data.as_ptr() as *mut _,
                _data.len() as u32,
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

        Ok(format!("Receipt sent to Windows printer '{}'", _printer_name))
    }

    #[cfg(not(windows))]
    {
        Err("Windows spooler transport is only available on Windows. Use 'serial' or 'network'.".to_string())
    }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Read;
    use std::net::TcpListener;
    use std::thread;

    /// Spin up a one-shot TCP listener on a random local port, run `f` against
    /// that port, then return all bytes the listener received.
    fn with_tcp_mock<F: FnOnce(u16) -> Result<String, String>>(f: F) -> (Result<String, String>, Vec<u8>) {
        let listener = TcpListener::bind("127.0.0.1:0").expect("bind");
        let port = listener.local_addr().expect("addr").port();

        let recv_handle = thread::spawn(move || {
            let (mut stream, _) = listener.accept().expect("accept");
            let mut buf = Vec::new();
            stream.read_to_end(&mut buf).expect("read");
            buf
        });

        let result = f(port);
        let received = recv_handle.join().expect("join");
        (result, received)
    }

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

    #[test]
    fn build_receipt_is_non_empty() {
        let data = build_receipt(&sample_receipt());
        assert!(!data.is_empty());
    }

    #[test]
    fn build_receipt_contains_init() {
        let data = build_receipt(&sample_receipt());
        assert!(data.starts_with(INIT));
    }

    #[test]
    fn build_receipt_ends_with_cut() {
        let data = build_receipt(&sample_receipt());
        assert!(data.ends_with(CUT));
    }

    #[test]
    fn build_receipt_contains_business_name() {
        let data = build_receipt(&sample_receipt());
        let text = String::from_utf8_lossy(&data);
        assert!(text.contains("Test Cafe"));
    }

    #[test]
    fn build_receipt_contains_item_name() {
        let data = build_receipt(&sample_receipt());
        let text = String::from_utf8_lossy(&data);
        assert!(text.contains("Espresso"));
    }

    #[test]
    fn build_receipt_includes_drawer_kick_when_enabled() {
        let mut r = sample_receipt();
        r.open_drawer = true;
        r.drawer_pin = Some(2);
        let data = build_receipt(&r);
        assert!(data
            .windows(5)
            .any(|w| w == [0x1B, 0x70, 0x00, 0x19, 0xFA]));
    }

    #[test]
    fn build_receipt_no_drawer_kick_when_disabled() {
        let data = build_receipt(&sample_receipt());
        assert!(!data
            .windows(5)
            .any(|w| w == [0x1B, 0x70, 0x00, 0x19, 0xFA]));
    }

    // ── TCP wire protocol tests ───────────────────────────────────────────────
    // These tests spin up a real TcpListener on localhost, call print_via_network,
    // and verify the exact bytes that arrive on the wire — no mocks.

    #[test]
    fn network_print_sends_escpos_init_bytes_over_tcp() {
        let (result, received) = with_tcp_mock(|port| {
            print_via_network("127.0.0.1", port, &build_receipt(&sample_receipt()))
        });

        assert!(result.is_ok(), "print_via_network failed: {:?}", result);
        assert!(received.starts_with(&[0x1B, 0x40]), "Missing ESC/POS INIT at start of stream");
    }

    #[test]
    fn network_print_ends_with_cut_command_over_tcp() {
        let (result, received) = with_tcp_mock(|port| {
            print_via_network("127.0.0.1", port, &build_receipt(&sample_receipt()))
        });

        assert!(result.is_ok());
        assert!(
            received.ends_with(&[0x1D, 0x56, 0x41, 0x03]),
            "Missing CUT command at end of stream"
        );
    }

    #[test]
    fn network_print_contains_business_name_in_tcp_stream() {
        let (result, received) = with_tcp_mock(|port| {
            print_via_network("127.0.0.1", port, &build_receipt(&sample_receipt()))
        });

        assert!(result.is_ok());
        let text = String::from_utf8_lossy(&received);
        assert!(text.contains("Test Cafe"), "Business name not found in TCP stream");
    }

    #[test]
    fn network_print_embeds_drawer_kick_in_tcp_stream() {
        let mut r = sample_receipt();
        r.open_drawer = true;
        r.drawer_pin = Some(2);

        let (result, received) = with_tcp_mock(|port| {
            print_via_network("127.0.0.1", port, &build_receipt(&r))
        });

        assert!(result.is_ok());
        assert!(
            received.windows(5).any(|w| w == [0x1B, 0x70, 0x00, 0x19, 0xFA]),
            "Drawer kick ESC/POS bytes not found in TCP stream"
        );
    }

    #[test]
    fn network_print_returns_ok_with_host_port_confirmation() {
        let (result, _) = with_tcp_mock(|port| {
            print_via_network("127.0.0.1", port, &build_receipt(&sample_receipt()))
        });

        let msg = result.expect("should succeed");
        assert!(msg.contains("127.0.0.1"), "Result message missing host");
    }
}
