/*
  File Created By: Mohamad Homsi
  Last Edited: Sunday 12 May 2024
  Last Edited by: Mohamad Homsi
*/

import { useState, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import i18next from "../../services/i18next";
import { Switch, Button } from "@mui/material";
import { useSelector } from "react-redux";

const VendorSettings = () => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({
    notifications: true,
  });
  const faviconRef = useRef(null);
  const iconRef = useRef(null);
  const user = useSelector((state) => state.user);

  useEffect(() => {
    const language = localStorage.getItem("language");
    if (language) {
      setFormData({
        ...formData,
        language,
      });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const selectLanguage = (language) => {
    i18next.changeLanguage(language);
    localStorage.setItem("language", language);
  };

  const handleNotificationsToggle = () => {};

  const handleImageChange = (e, type) => {};

  return (
    <div className="settings-container">
      <h2>{t("Settings.title")}</h2>
      <div className="InnerSettings">
        <section>
          <h3>Business Info</h3>

          <div className="NameAndSlogan">
            <lable>{t("Settings.BusinessName")}</lable>
            <input type="text" value={user?.user?.name} onChange={(e) => {}} />
            <lable>{t("Settings.slogan")}</lable>
            <input
              type="text"
              value={user?.user?.slogan}
              onChange={(e) => {}}
            />
            <Button
              variant="contained"
              color="primary"
              onClick={() => {
                // TODO: fix this
              }}
            >
              Update
            </Button>
          </div>
          <div>
            <p>{t("Settings.favicon")}</p>
            <img
              src={user?.user?.favicon ? user?.user?.favicon : "/favicon.ico"}
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
              onChange={(e) => {
                handleImageChange(e, "favicon");
              }}
            />
          </div>
          <div>
            <p>{t("Settings.logo")}</p>
            <img
              src={user?.user?.logo ? user?.user?.logo : "/logo.png"}
              alt="Icon"
              height={100}
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
              <option value="fr">French</option>
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

export default VendorSettings;
