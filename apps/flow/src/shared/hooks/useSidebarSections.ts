import { useCallback, useState } from 'react';
import type { SidebarSectionId } from './useSidebarSectionOrder';

const STORAGE_KEY = 'flow.sidebar.collapsedSections';

function load(): Record<string, boolean> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw) as Record<string, boolean>;
  } catch {
    /* ignore */
  }
  return {};
}

function save(next: Record<string, boolean>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  } catch {
    /* ignore */
  }
}

/** Collapsible sidebar groups; preference persisted in localStorage. */
export function useSidebarSections() {
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>(load);

  const toggle = useCallback((id: SidebarSectionId) => {
    setCollapsed((prev) => {
      const next = { ...prev, [id]: !prev[id] };
      save(next);
      return next;
    });
  }, []);

  const isCollapsed = useCallback((id: SidebarSectionId) => collapsed[id] === true, [collapsed]);

  return { toggle, isCollapsed };
}
