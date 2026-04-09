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
    items: [{ icon: User, title: "Account", path: "/settings/account" }],
  },
  {
    title: "Business",
    items: [
      { icon: Building2, title: "Business Type", path: "/settings/business-type" },
      { icon: Settings, title: "General", path: "/settings/general" },
      { icon: Palette, title: "Branding", path: "/settings/branding" },
    ],
  },
  {
    title: "Team",
    items: [
      { icon: ShieldCheck, title: "Permissions", path: "/settings/permissions" },
      { icon: Users, title: "Users", path: "/settings/users" },
    ],
  },
  {
    title: "Billing",
    items: [{ icon: CreditCard, title: "Payments", path: "/settings/payments" }],
  },
  {
    title: "Content",
    items: [
      { icon: File, title: "Files", path: "/settings/files" },
      { icon: Languages, title: "Languages", path: "/settings/languages" },
    ],
  },
  {
    title: "Legal",
    items: [{ icon: Gavel, title: "Legal Pages", path: "/settings/legal" }],
  },
];

const SettingsSidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <div className="h-full overflow-y-auto py-4 px-3">
      <nav className="space-y-5">
        {settingsSections.map((section, sectionIndex) => (
          <div key={sectionIndex}>
            <div className="px-2 mb-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted">
              {section.title}
            </div>
            <div className="space-y-0.5">
              {section.items.map((item) => {
                const Icon = item.icon;
                const isActive = location.pathname === item.path;
                return (
                  <div
                    key={item.path}
                    className={`flex items-center gap-2.5 px-2.5 py-2 rounded-lg cursor-pointer text-[13px] transition-colors ${
                      isActive
                        ? "bg-admin text-white font-medium"
                        : "text-primary hover:bg-black/5"
                    }`}
                    onClick={() => navigate(item.path)}
                  >
                    <Icon size={15} />
                    <span>{item.title}</span>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </nav>
    </div>
  );
};

export default SettingsSidebar;
