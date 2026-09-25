// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod hardware;
mod transport;
mod util;
mod validation;

use commands::{discovery, drawer, health, kds, printer, receipt};

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_process::init())
        .plugin(tauri_plugin_updater::Builder::new().build())
        .invoke_handler(tauri::generate_handler![
            // Legacy Windows-spooler drawer kick (kept for backward compat)
            printer::open_cash_drawer,
            // Cross-platform hardware discovery
            discovery::list_printers,
            discovery::list_serial_ports,
            discovery::discover_network_hardware,
            // Cross-platform drawer via serial or network
            drawer::open_cash_drawer_serial,
            drawer::open_cash_drawer_network,
            // ESC/POS receipt printing (serial / network / Windows spooler)
            receipt::print_receipt,
            // KDS hardware display (serial / network)
            kds::send_kds_ticket_serial,
            kds::send_kds_ticket_network,
            // Hardware health check
            health::check_hardware,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
