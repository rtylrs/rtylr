import { useLocation, useNavigate } from "react-router-dom";
import { useSidebarStore } from "../../../stores";
import SidebarMenuItem from "./SidebarMenuItem";
import SidebarDropdown from "./SidebarDropdown";
import { getMenuConfigByBusinessType, MenuItem } from "./menuConfig";
import { useBusinessStore } from "../../../stores/businessStore";
import { usePermissionsStore } from "../../../stores/permissionsStore";
import styles from "../Sidebar.module.scss";

const SidebarMenu = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { sectionsVisible, toggleSection } = useSidebarStore();
  const { businessType } = useBusinessStore();
  const { hasPermission } = usePermissionsStore();

  const handleNavigation = (path: string) => {
    navigate(path);
  };

  const menuConfig = getMenuConfigByBusinessType(businessType);
  const { topMenuItems, sections, dropdownSections, bottomMenuItems } = menuConfig;

  // Filter menu items based on permissions
  const filterMenuItems = (items: MenuItem[]): MenuItem[] => {
    return items.filter((item) => {
      if (!item.permissions || item.permissions.length === 0) return true;
      return item.permissions.some((permission) => hasPermission(permission as any));
    });
  };

  const filteredTopMenuItems = filterMenuItems(topMenuItems);
  const filteredBottomMenuItems = filterMenuItems(bottomMenuItems);
  const filteredDropdownSections = dropdownSections.map((section) => ({
    ...section,
    items: filterMenuItems(section.items),
  })).filter((section) => section.items.length > 0);

  return (
    <div className={styles.menu}>
      {/* Top menu items */}
      {filteredTopMenuItems.map((item, index) => (
        <SidebarMenuItem
          key={index}
          icon={item.icon}
          text={item.text}
          isActive={location.pathname === item.path}
          onClick={() => handleNavigation(item.path)}
        />
      ))}

      {/* New sections structure */}
      {sections?.map((section, sectionIndex) => {
        const filteredItems = section.items ? filterMenuItems(section.items) : [];
        const filteredDropdownSections = section.dropdownSections?.map((dropdownSection) => ({
          ...dropdownSection,
          items: filterMenuItems(dropdownSection.items),
        })).filter((dropdownSection) => dropdownSection.items.length > 0) || [];

        if (filteredItems.length === 0 && filteredDropdownSections.length === 0) {
          return null;
        }

        return (
          <div key={sectionIndex} className={styles.section}>
            <div className={styles.sectionHeader}>{section.header}</div>

            {/* Regular items in this section */}
            {filteredItems.map((item, index) => (
              <SidebarMenuItem
                key={index}
                icon={item.icon}
                text={item.text}
                isActive={location.pathname === item.path}
                onClick={() => handleNavigation(item.path)}
              />
            ))}

            {/* Dropdown sections in this section */}
            {filteredDropdownSections.map((dropdownSection, index) => (
              <SidebarDropdown
                key={index}
                icon={dropdownSection.icon}
                title={dropdownSection.title}
                sectionKey={dropdownSection.sectionKey}
                items={dropdownSection.items}
                isVisible={sectionsVisible[dropdownSection.sectionKey] || false}
                currentPath={location.pathname}
                onToggle={toggleSection}
                onNavigate={handleNavigation}
              />
            ))}
          </div>
        );
      })}

      {/* Fallback: Legacy dropdown sections if no sections defined */}
      {!sections && filteredDropdownSections.map((section, index) => (
        <SidebarDropdown
          key={index}
          icon={section.icon}
          title={section.title}
          sectionKey={section.sectionKey}
          items={section.items}
          isVisible={sectionsVisible[section.sectionKey] || false}
          currentPath={location.pathname}
          onToggle={toggleSection}
          onNavigate={handleNavigation}
        />
      ))}

      {/* Bottom menu items */}
      {filteredBottomMenuItems.map((item, index) => (
        <SidebarMenuItem
          key={index}
          icon={item.icon}
          text={item.text}
          isActive={location.pathname === item.path}
          onClick={() => handleNavigation(item.path)}
        />
      ))}
    </div>
  );
};

export default SidebarMenu;