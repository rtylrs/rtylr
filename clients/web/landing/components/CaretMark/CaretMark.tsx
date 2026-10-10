import React from 'react';

// The Caret knot. Same geometry as CaretIcon in @voxire/caret and the
// installed-app icon (clients/app/public/caret/icon.svg); copied here because
// the landing app doesn't depend on the Caret package.
const ribbon =
  'M10.55 11.8C8.3 9.8 7.1 7.4 8.7 5.3C10.2 3.3 13.3 3 15.15 4.65C17 6.3 17.35 8.9 16 10.9L14 9.8C14.9 8.5 14.55 6.9 13.5 6.05C12.45 5.2 10.8 5.35 10 6.4C9.15 7.55 10.1 9.15 11.9 10.8Z';

interface CaretMarkProps extends React.SVGProps<SVGSVGElement> {
  size?: number;
}

export const CaretMark: React.FC<CaretMarkProps> = ({ size = 16, ...props }) => (
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

export default CaretMark;
