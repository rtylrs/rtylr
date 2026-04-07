import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ASSETS } from "../../../../../lib/constants";
import styles from "./SettingsNavbar.module.scss";

const SettingsNavbar: React.FC = () => {
  const navigate = useNavigate();
  const [logoError, setLogoError] = useState(false);

  const customLogo = "/emblem.png";
  const fallbackLogo = ASSETS.logos.light;

  const handleLogoError = () => {
    setLogoError(true);
  };

  const handleLogoClick = () => {
    navigate("/");
  };

  return (
    <div className={styles.settingsNavbar}>
      <div className={styles.navbarContent}>
        <div className={styles.logo} onClick={handleLogoClick}>
          <img
            src={logoError ? fallbackLogo : customLogo}
            alt="Logo"
            className={styles.logoImage}
            onError={handleLogoError}
          />
        </div>
      </div>
    </div>
  );
};

export default SettingsNavbar;
