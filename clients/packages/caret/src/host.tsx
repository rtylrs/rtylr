import { createContext, useContext, type ReactNode } from "react";

/** A page of the host app that Caret points the person at, from a ```page
 *  block or an in-app link in a reply. */
export interface CaretPageSpec {
  path: string;
  /** The page's name, as Caret was told it. */
  title?: string;
  /** The button to press there, as the app labels it. */
  action?: string;
}

/** What the app around Caret lets it do. Without a host, page blocks show as
 *  plain links. */
export interface CaretHost {
  renderPage?: (spec: CaretPageSpec) => ReactNode;
  /** Opens a page of the app without leaving the chat. */
  openPage?: (spec: CaretPageSpec) => void;
}

const CaretHostContext = createContext<CaretHost>({});

export const CaretHostProvider = CaretHostContext.Provider;

export function useCaretHost(): CaretHost {
  return useContext(CaretHostContext);
}

/** Reads a ```page block. Only same-app paths are accepted. */
export function parsePageSpec(raw: string): CaretPageSpec | null {
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!value || typeof value !== "object") return null;
  const { path, title, action } = value as Record<string, unknown>;
  if (typeof path !== "string" || !isAppPath(path.trim())) return null;
  const text = (v: unknown) =>
    typeof v === "string" && v.trim() ? v.trim().slice(0, 80) : undefined;
  return {
    path: path.trim(),
    ...(text(title) ? { title: text(title) } : {}),
    ...(text(action) ? { action: text(action) } : {}),
  };
}

/** A path inside the app: "/team", not "//evil.example" or a full URL. */
export function isAppPath(path: string): boolean {
  return /^\/(?![/\\])/.test(path);
}
