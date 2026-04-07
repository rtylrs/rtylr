import { LucideIcon, Home, Settings, Palette } from "lucide-react";

export interface MenuItem {
  icon: LucideIcon;
  text: string;
  path: string;
  permissions?: string[];
}

export interface DropdownSection {
  icon: LucideIcon;
  title: string;
  sectionKey: string;
  items: MenuItem[];
}

export interface MenuConfig {
  topMenuItems: MenuItem[];
  dropdownSections: DropdownSection[];
  bottomMenuItems: MenuItem[];
}

export const vbitezMenuConfig: MenuConfig = {
  topMenuItems: [
    {
      icon: Home,
      text: "Dashboard",
      path: "/",
    },
  ],
  dropdownSections: [
    {
      icon: Palette,
      title: "Design",
      sectionKey: "design",
      items: [
        {
          icon: Palette,
          text: "Design Studio",
          path: "/design",
        },
      ],
    },
  ],
  bottomMenuItems: [{ icon: Settings, text: "Settings", path: "/settings" }],
};

export const getMenuConfig = (): MenuConfig => {
  return vbitezMenuConfig;
};
