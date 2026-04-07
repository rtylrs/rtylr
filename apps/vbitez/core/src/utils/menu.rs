use tauri::{CustomMenuItem, Menu, MenuItem, Submenu};

pub fn create_app_menu() -> Menu {
    let check_updates = CustomMenuItem::new("check_updates".to_string(), "Check for Updates");
    let quit = CustomMenuItem::new("quit".to_string(), "Quit");

    let app_menu = Submenu::new(
        "Vbitez Dashboard",
        Menu::new()
            .add_item(check_updates)
            .add_native_item(MenuItem::Separator)
            .add_item(quit),
    );

    Menu::new().add_submenu(app_menu)
}
