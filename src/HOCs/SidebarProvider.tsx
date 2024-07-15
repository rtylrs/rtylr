import {
  createContext,
  useContext,
  useState,
  ReactNode,
  useEffect,
} from "react";

interface SidebarContextType {
  customersVisible: boolean;
  managementVisible: boolean;
  configurationVisible: boolean;
  blogVisible: boolean;
  reportsVisible: boolean;
  ordersVisible: boolean;
  petitionsVisible: boolean;
  donationsVisible: boolean;
  menuVisible: boolean;
  toggleSection: (section: string) => void;
}

const defaultState: SidebarContextType = {
  customersVisible: false,
  managementVisible: false,
  configurationVisible: false,
  blogVisible: false,
  reportsVisible: false,
  ordersVisible: false,
  petitionsVisible: false,
  donationsVisible: false,
  menuVisible: false,
  toggleSection: () => {},
};

export const SidebarContext = createContext<SidebarContextType>(defaultState);

export function useSidebar() {
  return useContext(SidebarContext);
}

interface SidebarProviderProps {
  children: ReactNode;
}

export function SidebarProvider({ children }: SidebarProviderProps) {
  const getInitialState = (key: string, defaultValue: boolean = false) => {
    const savedValue = localStorage.getItem(key);
    return savedValue === null ? defaultValue : savedValue === "true";
  };

  const [customersVisible, setCustomersVisible] = useState(() =>
    getInitialState("customersVisible")
  );
  const [managementVisible, setManagementVisible] = useState(() =>
    getInitialState("managementVisible")
  );
  const [configurationVisible, setConfigurationVisible] = useState(() =>
    getInitialState("configurationVisible")
  );
  const [blogVisible, setBlogVisible] = useState(() =>
    getInitialState("blogVisible")
  );
  const [reportsVisible, setReportsVisible] = useState(() =>
    getInitialState("reportsVisible")
  );
  const [ordersVisible, setOrdersVisible] = useState(() =>
    getInitialState("ordersVisible")
  );
  const [petitionsVisible, setPetitionsVisible] = useState(() =>
    getInitialState("petitionsVisible")
  );
  const [donationsVisible, setDonationsVisible] = useState(() =>
    getInitialState("donationsVisible")
  );
  const [menuVisible, setMenuVisible] = useState(() =>
    getInitialState("menuVisible")
  );

  const toggleSection = (section: string) => {
    switch (section) {
      case "customers":
        setCustomersVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("customersVisible", newState.toString());
          return newState;
        });
        break;
      case "management":
        setManagementVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("managementVisible", newState.toString());
          return newState;
        });
        break;
      case "configuration":
        setConfigurationVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("configurationVisible", newState.toString());
          return newState;
        });
        break;
      case "blog":
        setBlogVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("blogVisible", newState.toString());
          return newState;
        });
        break;
      case "reports":
        setReportsVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("reportsVisible", newState.toString());
          return newState;
        });
        break;
      case "orders":
        setOrdersVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("ordersVisible", newState.toString());
          return newState;
        });
        break;
      case "petitions":
        setPetitionsVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("petitionsVisible", newState.toString());
          return newState;
        });
        break;
      case "donations":
        setDonationsVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("donationsVisible", newState.toString());
          return newState;
        });
        break;
      case "menu":
        setMenuVisible((prev) => {
          const newState = !prev;
          localStorage.setItem("menuVisible", newState.toString());
          return newState;
        });
        break;
      default:
        break;
    }
  };

  const contextValue: SidebarContextType = {
    customersVisible,
    managementVisible,
    configurationVisible,
    blogVisible,
    reportsVisible,
    ordersVisible,
    petitionsVisible,
    donationsVisible,
    menuVisible,
    toggleSection,
  };

  return (
    <SidebarContext.Provider value={contextValue}>
      {children}
    </SidebarContext.Provider>
  );
}
