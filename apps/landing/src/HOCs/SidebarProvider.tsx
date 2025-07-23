import {
  createContext,
  useContext,
  useState,
  ReactNode,
  useEffect,
} from "react";

interface SidebarContextType {
  sectionsVisible: { [key: string]: boolean };
  toggleSection: (section: string) => void;
}

const defaultState: SidebarContextType = {
  sectionsVisible: {},
  toggleSection: () => {},
};

export const SidebarContext = createContext<SidebarContextType>(defaultState);

export function useSidebar() {
  return useContext(SidebarContext);
}

interface SidebarProviderProps {
  children: ReactNode;
  initialSections: string[];
}

export function SidebarProvider({
  children,
  initialSections,
}: SidebarProviderProps) {
  const getInitialState = (key: string, defaultValue: boolean = false) => {
    const savedValue = localStorage.getItem(key);
    return savedValue === null ? defaultValue : savedValue === "true";
  };

  const initialVisibility = initialSections.reduce((acc, section) => {
    acc[section] = getInitialState(section + "Visible");
    return acc;
  }, {} as { [key: string]: boolean });

  const [sectionsVisible, setSectionsVisible] = useState(initialVisibility);

  useEffect(() => {
    Object.entries(sectionsVisible).forEach(([key, value]) => {
      localStorage.setItem(key + "Visible", value.toString());
    });
  }, [sectionsVisible]);

  const toggleSection = (section: string) => {
    setSectionsVisible((prev) => ({
      ...prev,
      [section]: !prev[section],
    }));
  };

  const contextValue: SidebarContextType = {
    sectionsVisible,
    toggleSection,
  };

  return (
    <SidebarContext.Provider value={contextValue}>
      {children}
    </SidebarContext.Provider>
  );
}
