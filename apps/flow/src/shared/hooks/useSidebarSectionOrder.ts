import { useCallback, useState } from 'react';

export const SIDEBAR_SECTION_IDS = ['plan', 'team', 'time', 'projects'] as const;
export type SidebarSectionId = (typeof SIDEBAR_SECTION_IDS)[number];

const STORAGE_KEY = 'flow.sidebar.sectionOrder-v1';

function normalize(raw: unknown): SidebarSectionId[] {
  if (!Array.isArray(raw)) return [...SIDEBAR_SECTION_IDS];
  const allowed = new Set<string>(SIDEBAR_SECTION_IDS);
  const seen = new Set<string>();
  const out: SidebarSectionId[] = [];
  for (const x of raw) {
    if (typeof x === 'string' && allowed.has(x) && !seen.has(x)) {
      seen.add(x);
      out.push(x as SidebarSectionId);
    }
  }
  for (const id of SIDEBAR_SECTION_IDS) {
    if (!seen.has(id)) out.push(id);
  }
  return out;
}

function load(): SidebarSectionId[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return normalize(JSON.parse(raw));
  } catch {
    /* ignore */
  }
  return [...SIDEBAR_SECTION_IDS];
}

function save(order: SidebarSectionId[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(order));
  } catch {
    /* ignore */
  }
}

/**
 * User-defined order of the collapsible sidebar groups (localStorage).
 *
 * Adding an id needs no storage migration: normalize() drops ids it does not
 * know and appends any known id missing from the stored array, so an older
 * stored order simply gains the new group at the end.
 */
export function useSidebarSectionOrder() {
  const [order, setOrder] = useState<SidebarSectionId[]>(load);

  const moveSection = useCallback((id: SidebarSectionId, direction: 'up' | 'down') => {
    setOrder((prev) => {
      const i = prev.indexOf(id);
      if (i < 0) return prev;
      const j = direction === 'up' ? i - 1 : i + 1;
      if (j < 0 || j >= prev.length) return prev;
      const next = [...prev];
      [next[i], next[j]] = [next[j], next[i]];
      save(next);
      return next;
    });
  }, []);

  return { order, moveSection };
}
