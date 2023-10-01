import { createContext, useContext, useState } from "react";
import PropTypes from "prop-types";

// Create a context for managing sidebar section visibility
export const SidebarContext = createContext();

// Create a custom hook to access the context

//eslint-disable-next-line
export function useSidebar() {
  return useContext(SidebarContext);
}

// Create a provider component to wrap your application
export function SidebarProvider({ children }) {
  // Define state to track the visibility of each section
  const [customersVisible, setCustomersVisible] = useState(false);
  const [managementVisible, setManagementVisible] = useState(false);
  const [configurationVisible, setConfigurationVisible] = useState(false);
  const [blogVisible, setBlogVisible] = useState(false);
  const [reportsVisible, setReportsVisible] = useState(false);
  const [ordersVisible, setOrdersVisible] = useState(false);
  const [petitionsVisible, setPetitionsVisible] = useState(false);
  const [donationsVisible, setDonationsVisible] = useState(false);

  // Create a function to toggle the visibility of a section
  const toggleSection = (section) => {
    switch (section) {
      case "customers":
        setCustomersVisible(!customersVisible);
        break;
      case "management":
        setManagementVisible(!managementVisible);
        break;
      case "configuration":
        setConfigurationVisible(!configurationVisible);
        break;
      case "blog":
        setBlogVisible(!blogVisible);
        break;
      case "reports":
        setReportsVisible(!reportsVisible);
        break;
      case "orders":
        setOrdersVisible(!ordersVisible);
        break;
      case "petitions":
        setPetitionsVisible(!petitionsVisible);
        break;
      case "donations":
        setDonationsVisible(!donationsVisible);
        break;
      default:
        break;
    }
  };

  // Provide the state and functions to the context
  const contextValue = {
    customersVisible,
    managementVisible,
    configurationVisible,
    blogVisible,
    reportsVisible,
    ordersVisible,
    petitionsVisible,
    donationsVisible,
    toggleSection,
  };

  return (
    <SidebarContext.Provider value={contextValue}>
      {children}
    </SidebarContext.Provider>
  );
}

// Define prop types
SidebarProvider.propTypes = {
  children: PropTypes.node.isRequired,
};
