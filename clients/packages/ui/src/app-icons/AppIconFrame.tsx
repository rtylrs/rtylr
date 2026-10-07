import type { LucideIcon, LucideProps } from "lucide-react";

export type AppIconProps = LucideProps;

export function AppIconFrame({
  icon: Icon,
  size = 24,
  strokeWidth = 1.8,
  ...props
}: AppIconProps & { icon: LucideIcon }) {
  return (
    <Icon
      size={size}
      strokeWidth={strokeWidth}
      aria-hidden="true"
      focusable="false"
      {...props}
    />
  );
}
