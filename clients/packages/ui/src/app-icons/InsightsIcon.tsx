import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function InsightsIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M4 4v16h16M8 16v-3M12 16v-6M16 16V7" />
      <path d="m7 8 5-3 4 2 4-4M17 3h3v3" />
    </AppIconFrame>
  );
}
