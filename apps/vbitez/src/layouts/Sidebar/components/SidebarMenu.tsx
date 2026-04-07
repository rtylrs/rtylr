import { useLocation, useNavigate } from "react-router-dom";
import { useSidebarStore } from "../../../stores";
import SidebarMenuItem from "./SidebarMenuItem";
import SidebarDropdown from "./SidebarDropdown";
import { getMenuConfig, MenuItem } from "./menuConfig";

const SidebarMenu = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { sectionsVisible, toggleSection } = useSidebarStore();

  const handleNavigation = (path: string) => {
    navigate(path);
  };

  const menuConfig = getMenuConfig();
  const { topMenuItems, dropdownSections, bottomMenuItems } = menuConfig;

  return (
    <div className="vb-sidebar__menu">
      {/* Top menu items */}
      {topMenuItems.map((item, index) => (
        <SidebarMenuItem
          key={index}
          icon={item.icon}
          text={item.text}
          isActive={location.pathname === item.path}
          onClick={() => handleNavigation(item.path)}
        />
      ))}

      {/* Dropdown sections */}
      {dropdownSections.map((section, index) => (
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
      {bottomMenuItems.map((item, index) => (
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