import { createContext, useContext } from 'react';
export const POSPortalContext = createContext(false);
export const usePOSPortal = () => useContext(POSPortalContext);
