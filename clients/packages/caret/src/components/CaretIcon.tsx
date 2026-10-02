import type { SVGProps } from "react";

export interface CaretIconProps extends SVGProps<SVGSVGElement> {
  size?: number;
}

const ribbon =
  "M10.55 11.8C8.3 9.8 7.1 7.4 8.7 5.3C10.2 3.3 13.3 3 15.15 4.65C17 6.3 17.35 8.9 16 10.9L14 9.8C14.9 8.5 14.55 6.9 13.5 6.05C12.45 5.2 10.8 5.35 10 6.4C9.15 7.55 10.1 9.15 11.9 10.8Z";

export default function CaretIcon({ size = 24, ...props }: CaretIconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      {...props}
    >
      <g transform="translate(12 12) scale(1.12) translate(-12 -12)">
        <path d={ribbon} />
        <path d={ribbon} transform="rotate(120 12 12)" />
        <path d={ribbon} transform="rotate(240 12 12)" />
      </g>
    </svg>
  );
}
