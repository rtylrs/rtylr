import React, {
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

  useEffect(() => {
    localStorage.setItem("customersVisible", customersVisible.toString());
  }, [customersVisible]);

  useEffect(() => {
    localStorage.setItem("managementVisible", managementVisible.toString());
  }, [managementVisible]);

  useEffect(() => {
    localStorage.setItem(
      "configurationVisible",
      configurationVisible.toString()
    );
  }, [configurationVisible]);

  useEffect(() => {
    localStorage.setItem("blogVisible", blogVisible.toString());
  }, [blogVisible]);

  useEffect(() => {
    localStorage.setItem("reportsVisible", reportsVisible.toString());
  }, [reportsVisible]);

  useEffect(() => {
    localStorage.setItem("ordersVisible", ordersVisible.toString());
  }, [ordersVisible]);

  useEffect(() => {
    localStorage.setItem("petitionsVisible", petitionsVisible.toString());
  }, [petitionsVisible]);

  useEffect(() => {
    localStorage.setItem("donationsVisible", donationsVisible.toString());
  }, [donationsVisible]);

  useEffect(() => {
    localStorage.setItem("menuVisible", menuVisible.toString());
  }, [menuVisible]);

  const toggleSection = (section: string) => {
    switch (section) {
      case "customers":
        setCustomersVisible((prev) => !prev);
        break;
      case "management":
        setManagementVisible((prev) => !prev);
        break;
      case "configuration":
        setConfigurationVisible((prev) => !prev);
        break;
      case "blog":
        setBlogVisible((prev) => !prev);
        break;
      case "reports":
        setReportsVisible((prev) => !prev);
        break;
      case "orders":
        setOrdersVisible((prev) => !prev);
        break;
      case "petitions":
        setPetitionsVisible((prev) => !prev);
        break;
      case "donations":
        setDonationsVisible((prev) => !prev);
        break;
      case "menu":
        setMenuVisible((prev) => !prev);
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
