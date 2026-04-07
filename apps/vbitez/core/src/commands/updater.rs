use tauri::{updater, Manager};

#[tauri::command]
pub async fn check_for_updates(app_handle: tauri::AppHandle) -> Result<String, String> {
    match updater::builder(app_handle.clone()).check().await {
        Ok(update) => {
            if update.is_update_available() {
                Ok(format!("Update available: {}", update.latest_version()))
            } else {
                Ok("No updates available".to_string())
            }
        }
        Err(e) => Err(format!("Failed to check for updates: {}", e))
    }
}

pub async fn handle_update_check(app_handle: tauri::AppHandle) {
    match check_for_updates(app_handle.clone()).await {
        Ok(message) => {
            let _ = tauri::api::dialog::message(
                Some(&app_handle.get_window("main").unwrap()),
                "Update Check",
                message
            );
        }
        Err(error) => {
            let _ = tauri::api::dialog::message(
                Some(&app_handle.get_window("main").unwrap()),
                "Update Check Error",
                error
            );
        }
    }
}