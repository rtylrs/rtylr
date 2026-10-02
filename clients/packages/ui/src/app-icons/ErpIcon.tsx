import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function ErpIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="m3.5 7.5 8.5-5 8.5 5v9L12 21.5l-8.5-5Z" />
      <path d="m3.5 7.5 8.5 5 8.5-5M12 12.5v9M8 5l8.5 5v4" />
    </AppIconFrame>
  );
}
