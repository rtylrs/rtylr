import { createContext, useContext, type Dispatch, type SetStateAction } from 'react';

interface SidebarCollapseContextValue {
  collapsed: boolean;
  setCollapsed: Dispatch<SetStateAction<boolean>>;
}

const SidebarCollapseContext = createContext<SidebarCollapseContextValue>({
  collapsed: false,
  setCollapsed: () => undefined,
});

export const SidebarCollapseProvider = SidebarCollapseContext.Provider;

export function useSidebarCollapse(): SidebarCollapseContextValue {
  return useContext(SidebarCollapseContext);
}
