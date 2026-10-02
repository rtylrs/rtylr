import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function HomeIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="m3.5 10 8.5-7 8.5 7v9a1.5 1.5 0 0 1-1.5 1.5H5A1.5 1.5 0 0 1 3.5 19Z" />
      <path d="M9 20.5v-7h6v7M15.5 6l5 4" />
    </AppIconFrame>
  );
}
