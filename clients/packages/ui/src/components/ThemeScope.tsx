import React, {
  createContext,
  useContext,
  type CSSProperties,
  type ReactNode,
} from "react";

export type ThemeScopeStyle = CSSProperties & {
  [key: `--${string}`]: string | number | undefined;
};

const ThemeScopeContext = createContext<ThemeScopeStyle | undefined>(undefined);

// React context preserves app tokens when overlays render outside the app DOM.
export function ThemeScopeProvider({
  style,
  children,
}: {
  style: ThemeScopeStyle;
  children: ReactNode;
}) {
  return (
    <ThemeScopeContext.Provider value={style}>
      {children}
    </ThemeScopeContext.Provider>
  );
}

export function useThemeScope(): ThemeScopeStyle | undefined {
  return useContext(ThemeScopeContext);
}
