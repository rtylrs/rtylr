import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function MenuIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M5 3.5h11l3 3v14H5ZM15.5 3.5V7H19M8 16h8M8 18.5h5" />
      <path d="M8 13h8a4 4 0 0 0-8 0ZM12 8V7" />
    </AppIconFrame>
  );
}
