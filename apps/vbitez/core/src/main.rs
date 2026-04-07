// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod services;
mod utils;

use commands::{printer, updater};
use utils::menu;
use tauri::Manager; 

fn main() {
    let app_menu = menu::create_app_menu();

    tauri::Builder::default()
        .menu(app_menu)
        .on_menu_event(|event| {
            match event.menu_item_id() {
                "check_updates" => {
                    let app_handle = event.window().app_handle();
                    tauri::async_runtime::spawn(async move {
                        updater::handle_update_check(app_handle).await;
                    });
                }
                "quit" => {
                    std::process::exit(0);
                }
                _ => {}
            }
        })
        .invoke_handler(tauri::generate_handler![
            updater::check_for_updates,
            printer::open_cash_drawer
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
