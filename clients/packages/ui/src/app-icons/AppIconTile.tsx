import type { ComponentType, CSSProperties } from "react";
import { appIconAccentStyle, type AppIconProps } from "./createAppIcon";

export interface AppIconTileProps {
  icon: ComponentType<AppIconProps>;
  /** The app's colour, e.g. var(--app-crm). */
  color: string;
  size?: number;
  className?: string;
}

/** An app icon as a launcher tile: a white glyph on the app's colour. */
export function AppIconTile({ icon: Icon, color, size = 34, className }: AppIconTileProps) {
  const style: CSSProperties = {
    ...appIconAccentStyle("#fff"),
    width: size,
    height: size,
    flexShrink: 0,
    display: "inline-grid",
    placeItems: "center",
    borderRadius: Math.round(size * 0.27),
    background: color,
    color: "#fff",
  };
  return (
    <span className={className} style={style} aria-hidden="true">
      <Icon size={Math.round(size * 0.59)} />
    </span>
  );
}
