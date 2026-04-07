import type { IconProps } from "phosphor-react";
import { House, Palette, GearSix, User, Info } from "phosphor-react";

type SearchIcon = React.ComponentType<IconProps>;
export interface SearchItem {
  id: string;
  title: string;
  icon: SearchIcon;
  platform: "rtylr" | "global";
  path: string;
  action: () => void;
  description?: string;
}

type NavigateFunction = (path: string) => void;

export const createSearchItems = (navigate: NavigateFunction): SearchItem[] => [
  // Rtylr Platform Items
  {
    id: "rtylr-home",
    title: "Dashboard",
    icon: House,
    platform: "rtylr",
    path: "/",
    action: () => navigate("/"),
    description: "Rtylr Dashboard",
  },
  {
    id: "rtylr-design",
    title: "Design Studio",
    icon: Palette,
    platform: "rtylr",
    path: "/design",
    action: () => navigate("/design"),
    description: "Create and edit designs",
  },
  {
    id: "rtylr-settings",
    title: "Settings",
    icon: GearSix,
    platform: "rtylr",
    path: "/settings",
    action: () => navigate("/settings"),
    description: "Application settings",
  },

  // Global Items
  {
    id: "global-profile",
    title: "Profile",
    icon: User,
    platform: "global",
    path: "/profile",
    action: () => navigate("/profile"),
    description: "User profile settings",
  },
  {
    id: "global-help",
    title: "Help & Support",
    icon: Info,
    platform: "global",
    path: "/help",
    action: () => navigate("/help"),
    description: "Get help and support",
  },
];

export const searchItems = (
  items: SearchItem[],
  keyword: string,
  currentPlatform?: string
): SearchItem[] => {
  if (!keyword.trim()) return [];

  const searchTerm = keyword.toLowerCase();
  const filteredItems = items.filter(
    (item) =>
      item.title.toLowerCase().includes(searchTerm) ||
      item.description?.toLowerCase().includes(searchTerm) ||
      item.platform.toLowerCase().includes(searchTerm)
  );

  if (currentPlatform) {
    const platformItems = filteredItems.filter(
      (item) => item.platform === currentPlatform
    );
    const globalItems = filteredItems.filter(
      (item) => item.platform === "global"
    );

    return [...platformItems.slice(0, 5), ...globalItems.slice(0, 2)];
  }

  return filteredItems.slice(0, 5);
};
