import React from "react";
import { useNavigate, useLocation } from "react-router-dom";
import {
  GearSix as Settings,
  ShieldCheckered as ShieldCheck,
  CreditCard,
  File,
  Translate as Languages,
  Scales as Gavel,
  Buildings as Building2,
  User,
  UsersThree as Users,
  Palette,
} from "phosphor-react";
import styles from "./SettingsSidebar.module.scss";

interface SettingsMenuItem {
  icon: React.ComponentType<any>;
  title: string;
  path: string;
}

interface SettingsSection {
  title: string;
  items: SettingsMenuItem[];
}

const settingsSections: SettingsSection[] = [
  {
    title: "Personal",
    items: [
      {
        icon: User,
        title: "Account",
        path: "/settings/account",
      },
    ],
  },
  {
    title: "Business",
    items: [
      {
        icon: Building2,
        title: "Business Type",
        path: "/settings/business-type",
      },
      {
        icon: Settings,
        title: "General",
        path: "/settings/general",
      },
      {
        icon: Palette,
        title: "Branding",
        path: "/settings/branding",
      },
    ],
  },
  {
    title: "Team",
    items: [
      {
        icon: ShieldCheck,
        title: "Permissions",
        path: "/settings/permissions",
      },
      {
        icon: Users,
        title: "Users",
        path: "/settings/users",
      },
    ],
  },
  {
    title: "Billing",
    items: [
      {
        icon: CreditCard,
        title: "Payments",
        path: "/settings/payments",
      },
    ],
  },
  {
    title: "Content",
    items: [
      {
        icon: File,
        title: "Files",
        path: "/settings/files",
      },
      {
        icon: Languages,
        title: "Languages",
        path: "/settings/languages",
      },
    ],
  },
  {
    title: "Legal",
    items: [
      {
        icon: Gavel,
        title: "Legal Pages",
        path: "/settings/legal",
      },
    ],
  },
];

const SettingsSidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <div className={styles.settingsSidebar}>
      <nav className={styles.sidebarNav}>
        {settingsSections.map((section, sectionIndex) => {
          return (
            <div key={sectionIndex} className={styles.section}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionTitle}>{section.title}</span>
              </div>
              <div className={styles.sectionItems}>
                {section.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = location.pathname === item.path;
                  return (
                    <div
                      key={item.path}
                      className={`${styles.navItem} ${
                        isActive ? styles.active : ""
                      }`}
                      onClick={() => navigate(item.path)}
                    >
                      <Icon size={16} className={styles.navItemIcon} />
                      <span className={styles.navItemText}>{item.title}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          );
        })}
      </nav>
    </div>
  );
};

export default SettingsSidebar;

