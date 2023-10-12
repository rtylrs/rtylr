import { useState, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { auth, db, storage, firebase } from "../../firebase";
import i18next from "../../services/i18next";
import { Switch, Button } from "@mui/material";

const AdminSettings = () => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({
    language: "en",
    notifications: true,
  });
  const [config, setConfig] = useState({});
  const faviconRef = useRef(null);

  useEffect(() => {
    const language = localStorage.getItem("language");
    if (language) {
      setFormData({
        ...formData,
        language,
      });
    }

    db.collection("config")
      .doc("websiteData")
      .get()
      .then((doc) => {
        if (doc.exists) {
          setConfig(doc.data());
        }
      });

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

  const handleFaviconChange = (e) => {
    const file = e.target.files[0];
    const storageRef = storage.ref();
    const fileRef = storageRef.child(`config/${file.name}`);
    fileRef.put(file).then(() => {
      fileRef.getDownloadURL().then((url) => {
        db.collection("config")
          .doc("websiteData")
          .set(
            {
              favicon: url,
            },
            { merge: true }
          )
          .then(() => {
            setConfig({
              ...config,
              favicon: url,
            });
          });
      });
    });
  };

  return (
    <div className="settings-container">
      <h2>{t("Settings.title")}</h2>

      <div>
        <img
          src={
            config.favicon ? config.favicon : "https://picsum.photos/150/150"
          }
          alt="favicon"
          height={50}
          width={50}
          onClick={() => {
            faviconRef.current.click();
          }}
        />
        <input
          type="file"
          accept="image/*"
          style={{
            display: "none",
          }}
          ref={faviconRef}
          onChange={handleFaviconChange}
        />
        <p>{t("Settings.favicon")}</p>
      </div>

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
    </div>
  );
};

export default AdminSettings;
