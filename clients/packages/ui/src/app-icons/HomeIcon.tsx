import { createAppIcon } from "./createAppIcon";

/** A house; the door is the accent. */
export const HomeIcon = createAppIcon(
  "HomeIcon",
  <>
    <path d="M4 10.25 12 3.75l8 6.5v8.5a1.75 1.75 0 0 1-1.75 1.75H5.75A1.75 1.75 0 0 1 4 18.75z" />
  </>,
  <>
    <rect x="9.75" y="13.5" width="4.5" height="7" rx="1.25" />
  </>,
);
