import type { SVGProps } from "react";

export interface CaretIconProps extends SVGProps<SVGSVGElement> {
  size?: number;
}

export default function CaretIcon({ size = 24, ...props }: CaretIconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
      {...props}
    >
      <path
        d="M7 3.5h10a3.5 3.5 0 0 1 3.5 3.5v8a3.5 3.5 0 0 1-3.5 3.5H9l-5.5 3V7A3.5 3.5 0 0 1 7 3.5Z"
        stroke="currentColor"
        strokeWidth="1.65"
        strokeLinejoin="round"
      />
      <path
        d="m7.5 13 4.5-5 4.5 5"
        stroke="currentColor"
        strokeWidth="1.9"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
