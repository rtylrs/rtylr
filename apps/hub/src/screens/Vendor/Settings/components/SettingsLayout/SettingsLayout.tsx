import React from "react";
import { Outlet } from "react-router-dom";
import SettingsSidebar from "../SettingsSidebar";
import SettingsNavbar from "../SettingsNavbar";
import styles from "./SettingsLayout.module.scss";

const SettingsLayout: React.FC = () => {
  return (
    <div className={styles.settingsLayout}>
      <div className={styles.settingsNavbar}>
        <SettingsNavbar />
      </div>
      <div className={styles.settingsSidebar}>
        <SettingsSidebar />
      </div>
      <div className={styles.settingsContent}>
        <Outlet />
      </div>
    </div>
  );
};

export default SettingsLayout;

