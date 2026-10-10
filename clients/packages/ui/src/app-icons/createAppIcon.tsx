import { forwardRef, type CSSProperties, type ReactNode } from "react";
import type { LucideProps } from "lucide-react";

/** Lucide-compatible, so an app icon fits anywhere a Lucide icon does. */
export interface AppIconProps extends LucideProps {
  /**
   * Colour of the icon's accent part (the "fold"). Without it the accent is
   * a light tint of the icon colour, unless a parent sets the accent through
   * appIconAccentStyle().
   */
  accent?: string;
}

type AccentKind = "fill" | "stroke";

/** CSS variables that colour the accent part of every app icon inside. */
export function appIconAccentStyle(color: string): CSSProperties & Record<string, string> {
  return {
    "--app-icon-accent": color,
    "--app-icon-accent-stroke": color,
    "--app-icon-accent-opacity": "1",
  };
}

/**
 * Builds one app icon on the shared 24px grid: a line outline plus one accent
 * part. The accent picks up the app's colour where the app is shown (the
 * active rail item, launchers) and stays a quiet tint everywhere else, which
 * echoes the coloured fold of the rtylr mark.
 */
export function createAppIcon(name: string, outline: ReactNode, accentPart: ReactNode, kind: AccentKind = "fill") {
  const AppIcon = forwardRef<SVGSVGElement, AppIconProps>(function AppIcon(
    { size = 24, strokeWidth = 1.75, accent, absoluteStrokeWidth: _absolute, style, ...props },
    ref,
  ) {
    const accentStyle: CSSProperties =
      kind === "stroke"
        ? { stroke: accent ?? "var(--app-icon-accent, currentColor)" }
        : {
            fill: accent ?? "var(--app-icon-accent, currentColor)",
            fillOpacity: accent ? 1 : "var(--app-icon-accent-opacity, 0.22)",
            stroke: accent ?? "var(--app-icon-accent-stroke, currentColor)",
          };
    return (
      <svg
        ref={ref}
        width={size}
        height={size}
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
        focusable="false"
        style={style}
        {...props}
      >
        {outline}
        <g style={accentStyle}>{accentPart}</g>
      </svg>
    );
  });
  AppIcon.displayName = name;
  return AppIcon;
}
