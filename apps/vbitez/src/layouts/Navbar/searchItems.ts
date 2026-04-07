import { LucideIcon, Home, Palette, Settings, User, Info } from "lucide-react";

export interface SearchItem {
  id: string;
  title: string;
  icon: LucideIcon;
  platform: "vbitez" | "global";
  path: string;
  action: () => void;
  description?: string;
}

type NavigateFunction = (path: string) => void;

export const createSearchItems = (navigate: NavigateFunction): SearchItem[] => [
  // Vbitez Platform Items
  {
    id: "vbitez-home",
    title: "Dashboard",
    icon: Home,
    platform: "vbitez",
    path: "/",
    action: () => navigate("/"),
    description: "Vbitez Dashboard",
  },
  {
    id: "vbitez-design",
    title: "Design Studio",
    icon: Palette,
    platform: "vbitez",
    path: "/design",
    action: () => navigate("/design"),
    description: "Create and edit designs",
  },
  {
    id: "vbitez-settings",
    title: "Settings",
    icon: Settings,
    platform: "vbitez",
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
