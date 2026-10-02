import { createContext, useContext, type ReactNode } from 'react';

const PortalHeaderActions = createContext<ReactNode>(null);

export const PortalHeaderActionsProvider = PortalHeaderActions.Provider;
export const usePortalHeaderActions = () => useContext(PortalHeaderActions);
