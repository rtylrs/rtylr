import type { PropsWithChildren, SVGProps } from "react";

export type AppIconProps = SVGProps<SVGSVGElement> & { size?: number | string };

export function AppIconFrame({
  size = 24,
  children,
  ...props
}: PropsWithChildren<AppIconProps>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.65"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
      {...props}
    >
      {children}
    </svg>
  );
}
