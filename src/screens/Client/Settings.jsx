import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { auth, db } from "../../firebase";
import i18next from "../../services/i18next";
import { Switch, Button } from "@mui/material";

const SettingsPage = () => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({
    language: "en",
    notifications: true,
  });

  useEffect(() => {
    const language = localStorage.getItem("language");
    if (language) {
      setFormData({
        ...formData,
        language,
      });
    }

    // Fetch user's notification preference and other settings from a database or other source
    // and update the state accordingly
    // Example:
    // const userSettings = fetchUserSettings(); // Replace with actual fetch
    // setFormData({
    //   ...formData,
    //   language: userSettings.language,
    //   notifications: userSettings.notifications,
    // });

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const selectLanguage = (language) => {
    i18next.changeLanguage(language);
    localStorage.setItem("language", language);
  };

  const handleNotificationsToggle = () => {
    // Update user's notification preference in a database or other source
    // and update the state accordingly
    // Example:
    // const updatedNotifications = !formData.notifications;
    // updateNotificationsInDatabase(updatedNotifications); // Replace with actual update
    // setFormData({
    //   ...formData,
    //   notifications: updatedNotifications,
    // });
  };

  const handleExportData = () => {
    // Implement data export logic here (e.g., generate and download a data export file)
    // Example:
    // exportUserData(); // Replace with actual export logic
  };

  const handleDeleteAccount = () => {
    // Display a confirmation dialog to confirm account deletion
    const confirmDelete = window.confirm(t("Settings.confirmDeleteAccount"));
    if (confirmDelete) {
      // Perform the account deletion action (e.g., delete user's account and data)
      // Example:
      // deleteAccount(); // Replace with actual account deletion code
    }
  };

  return (
    <div className="settings-container">
      <h2>{t("Settings.title")}</h2>
      <div>
        <p>{t("Settings.language")}:</p>
        <select
          name="language"
          value={formData.language}
          onChange={(e) => {
            selectLanguage(e.target.value);
            setFormData({
              ...formData,
              language: e.target.value,
            });
          }}
        >
          <option value="en">English</option>
          <option value="ar">Arabic</option>
        </select>
      </div>
      <div>
        <p>{t("Settings.notifications")}:</p>
        <label>
          <Switch
            color="primary"
            checked={formData.notifications}
            onChange={handleNotificationsToggle}
          />
          {formData.notifications ? t("Settings.on") : t("Settings.off")}
        </label>
      </div>
      <div>
        <Button
          onClick={() => {
            auth.sendPasswordResetEmail(formData.email).then(() => {});
          }}
          variant="contained"
        >
          Forgot Password
        </Button>
      </div>
      <div>
        <Button variant="contained" color="primary" onClick={handleExportData}>
          {t("Settings.exportData")}
        </Button>
      </div>
      <div>
        <Button variant="contained" color="error" onClick={handleDeleteAccount}>
          {t("Settings.deleteAccount")}
        </Button>
      </div>
    </div>
  );
};

export default SettingsPage;
