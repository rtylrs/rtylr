import type { IconProps } from "phosphor-react";
import {
  House,
  GearSix,
  Palette,
  List,
  ShoppingBag,
  ChartBar,
  QrCode,
  CreditCard,
  PuzzlePiece,
  UsersThree,
  Barbell,
  Package,
  Scissors,
  Coffee,
  ForkKnife,
  Buildings,
} from "phosphor-react";
import { BusinessType } from "../../../stores/businessStore";
import { PERMISSIONS } from "../../../utils/permissions";

type SidebarIcon = React.ComponentType<IconProps>;
export interface MenuItem {
  icon: SidebarIcon;
  text: string;
  path: string;
  permissions?: string[];
}

export interface DropdownSection {
  icon: SidebarIcon;
  title: string;
  sectionKey: string;
  items: MenuItem[];
}

export interface MenuSection {
  header: string;
  items: MenuItem[];
  dropdownSections?: DropdownSection[];
}

export interface MenuConfig {
  topMenuItems: MenuItem[];
  dropdownSections: DropdownSection[];
  bottomMenuItems: MenuItem[];
  sections?: MenuSection[];
}

// Restaurant menu configuration
export const restaurantMenuConfig: MenuConfig = {
  topMenuItems: [
    {
      icon: House,
      text: "Account home",
      path: "/",
      permissions: [],
    },
  ],
  sections: [
    {
      header: "RECENTS",
      items: [
        {
          icon: ChartBar,
          text: "Analytics & logs",
          path: "/app/business/analytics",
          permissions: [PERMISSIONS.VIEW_ANALYTICS],
        },
      ],
    },
    {
      header: "BUILD",
      dropdownSections: [
        {
          icon: List,
          title: "Menu",
          sectionKey: "menu",
          items: [
            {
              icon: List,
              text: "Menus",
              path: "/app/business/menus",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
            {
              icon: List,
              text: "Menu Items",
              path: "/app/business/menu",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
          ],
        },
        {
          icon: Palette,
          title: "Design",
          sectionKey: "design",
          items: [
            {
              icon: Palette,
              text: "Design Studio",
              path: "/app/business/design",
              permissions: [],
            },
          ],
        },
      ],
    },
    {
      header: "MANAGE",
      items: [
        {
          icon: QrCode,
          text: "QR Codes",
          path: "/app/business/qr-codes",
          permissions: [],
        },
        {
          icon: PuzzlePiece,
          text: "Addons",
          path: "/app/business/addons",
          permissions: [],
        },
        {
          icon: CreditCard,
          text: "Billing",
          path: "/app/business/billing",
          permissions: [PERMISSIONS.MANAGE_BILLING],
        },
        {
          icon: GearSix,
          text: "Settings",
          path: "/settings",
          permissions: [PERMISSIONS.MANAGE_SETTINGS],
        },
      ],
    },
  ],
  dropdownSections: [],
  bottomMenuItems: [],
};

// Gym menu configuration
export const gymMenuConfig: MenuConfig = {
  topMenuItems: [
    {
      icon: House,
      text: "Account home",
      path: "/",
      permissions: [],
    },
  ],
  sections: [
    {
      header: "BUILD",
      dropdownSections: [
        {
          icon: Barbell,
          title: "Memberships",
          sectionKey: "memberships",
          items: [
            {
              icon: UsersThree,
              text: "Members",
              path: "/app/business/members",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
            {
              icon: Package,
              text: "Plans",
              path: "/app/business/plans",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
          ],
        },
        {
          icon: Palette,
          title: "Design",
          sectionKey: "design",
          items: [
            {
              icon: Palette,
              text: "Design Studio",
              path: "/app/business/design",
              permissions: [],
            },
          ],
        },
      ],
    },
    {
      header: "MANAGE",
      items: [
        {
          icon: ChartBar,
          text: "Analytics",
          path: "/app/business/analytics",
          permissions: [PERMISSIONS.VIEW_ANALYTICS],
        },
        {
          icon: QrCode,
          text: "QR Codes",
          path: "/app/business/qr-codes",
          permissions: [],
        },
        {
          icon: PuzzlePiece,
          text: "Addons",
          path: "/app/business/addons",
          permissions: [],
        },
        {
          icon: CreditCard,
          text: "Billing",
          path: "/app/business/billing",
          permissions: [PERMISSIONS.MANAGE_BILLING],
        },
        {
          icon: GearSix,
          text: "Settings",
          path: "/settings",
          permissions: [PERMISSIONS.MANAGE_SETTINGS],
        },
      ],
    },
  ],
  dropdownSections: [],
  bottomMenuItems: [],
};

// Retail menu configuration
export const retailMenuConfig: MenuConfig = {
  topMenuItems: [
    {
      icon: House,
      text: "Account home",
      path: "/",
      permissions: [],
    },
  ],
  sections: [
    {
      header: "BUILD",
      dropdownSections: [
        {
          icon: ShoppingBag,
          title: "Products",
          sectionKey: "products",
          items: [
            {
              icon: Package,
              text: "Products",
              path: "/app/business/products",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
            {
              icon: Package,
              text: "Inventory",
              path: "/app/business/inventory",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
          ],
        },
        {
          icon: Palette,
          title: "Design",
          sectionKey: "design",
          items: [
            {
              icon: Palette,
              text: "Design Studio",
              path: "/app/business/design",
              permissions: [],
            },
          ],
        },
      ],
    },
    {
      header: "MANAGE",
      items: [
        {
          icon: ChartBar,
          text: "Analytics",
          path: "/app/business/analytics",
          permissions: [PERMISSIONS.VIEW_ANALYTICS],
        },
        {
          icon: QrCode,
          text: "QR Codes",
          path: "/app/business/qr-codes",
          permissions: [],
        },
        {
          icon: PuzzlePiece,
          text: "Addons",
          path: "/app/business/addons",
          permissions: [],
        },
        {
          icon: CreditCard,
          text: "Billing",
          path: "/app/business/billing",
          permissions: [PERMISSIONS.MANAGE_BILLING],
        },
        {
          icon: GearSix,
          text: "Settings",
          path: "/settings",
          permissions: [PERMISSIONS.MANAGE_SETTINGS],
        },
      ],
    },
  ],
  dropdownSections: [],
  bottomMenuItems: [],
};

// Default/Generic menu configuration
export const defaultMenuConfig: MenuConfig = {
  topMenuItems: [
    {
      icon: House,
      text: "Account home",
      path: "/",
      permissions: [],
    },
  ],
  sections: [
    {
      header: "BUILD",
      dropdownSections: [
        {
          icon: Package,
          title: "Products",
          sectionKey: "products",
          items: [
            {
              icon: Package,
              text: "Products",
              path: "/app/business/products",
              permissions: [PERMISSIONS.MANAGE_MENU],
            },
          ],
        },
        {
          icon: Palette,
          title: "Design",
          sectionKey: "design",
          items: [
            {
              icon: Palette,
              text: "Design Studio",
              path: "/app/business/design",
              permissions: [],
            },
          ],
        },
      ],
    },
    {
      header: "MANAGE",
      items: [
        {
          icon: ChartBar,
          text: "Analytics",
          path: "/app/business/analytics",
          permissions: [PERMISSIONS.VIEW_ANALYTICS],
        },
        {
          icon: QrCode,
          text: "QR Codes",
          path: "/app/business/qr-codes",
          permissions: [],
        },
        {
          icon: PuzzlePiece,
          text: "Addons",
          path: "/app/business/addons",
          permissions: [],
        },
        {
          icon: CreditCard,
          text: "Billing",
          path: "/app/business/billing",
          permissions: [PERMISSIONS.MANAGE_BILLING],
        },
        {
          icon: GearSix,
          text: "Settings",
          path: "/settings",
          permissions: [PERMISSIONS.MANAGE_SETTINGS],
        },
      ],
    },
  ],
  dropdownSections: [],
  bottomMenuItems: [],
};

export const getMenuConfigByBusinessType = (businessType: BusinessType | null): MenuConfig => {
  switch (businessType) {
    case "restaurant":
      return restaurantMenuConfig;
    case "gym":
      return gymMenuConfig;
    case "retail":
      return retailMenuConfig;
    default:
      return defaultMenuConfig;
  }
};

// Legacy function for backward compatibility
export const getMenuConfig = (): MenuConfig => {
  return restaurantMenuConfig;
};
