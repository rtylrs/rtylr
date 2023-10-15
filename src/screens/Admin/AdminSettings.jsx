import { useState, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { auth, db, storage, firebase } from "../../firebase";
import i18next from "../../services/i18next";
import { Switch, Button } from "@mui/material";

const AdminSettings = () => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({
    notifications: true,
  });
  const [config, setConfig] = useState({});
  const faviconRef = useRef(null);
  const iconRef = useRef(null);

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

  const handleImageChange = (e, type) => {
    const file = e.target.files[0];
    const storageRef = storage.ref();
    const fileRef = storageRef.child(`config/${file.name}`);
    fileRef.put(file).then(() => {
      fileRef.getDownloadURL().then((url) => {
        const updateField = type === "favicon" ? "favicon" : "icon";
        const updatedConfig = {
          ...config,
          [updateField]: url,
        };

        db.collection("config")
          .doc("websiteData")
          .set(
            {
              [updateField]: url,
            },
            { merge: true }
          )
          .then(() => {
            setConfig(updatedConfig);
          });
      });
    });
  };

  return (
    <div className="settings-container">
      <h2>{t("Settings.title")}</h2>
      <div className="InnerSettings">
        <section>
          <h3>Business Info</h3>
          <div>
            <p>{t("Settings.favicon")}</p>
            <img
              src={
                config.favicon
                  ? config.favicon
                  : "https://picsum.photos/150/150"
              }
              alt="favicon"
              height={70}
              width={70}
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
              onChange={(e) => {
                handleImageChange(e, "favicon");
              }}
            />
          </div>
          <div>
            <p>{t("Settings.icon")}</p>
            <img
              src={config.icon ? config.icon : "https://picsum.photos/150/150"}
              alt="Icon"
              height={70}
              width={70}
              onClick={() => {
                iconRef.current.click();
              }}
            />
            <input
              type="file"
              accept="image/*"
              style={{
                display: "none",
              }}
              ref={iconRef}
              onChange={(e) => {
                handleImageChange(e, "icon");
              }}
            />
          </div>
        </section>

        <section>
          <h3>Custom Settings</h3>
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
        </section>
      </div>
    </div>
  );
};

export default AdminSettings;
