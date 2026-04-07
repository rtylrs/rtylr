import { useState } from "react";
import { ASSETS } from "../../../lib/constants";
import styles from "../Sidebar.module.scss";

const SidebarHeader = () => {
  const [logoError, setLogoError] = useState(false);

  const customLogo = "/emblem.png";
  const fallbackLogo = ASSETS.logos.light;

  const handleLogoError = () => {
    setLogoError(true);
  };

  return (
    <div className={styles.header}>
      <div className={styles.logo}>
        <img
          src={logoError ? fallbackLogo : customLogo}
          alt="Logo"
          className={styles.logoImage}
          onError={handleLogoError}
        />
      </div>
    </div>
  );
};

export default SidebarHeader;