import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function CrmIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M3 4.5h18v13H9l-4 4v-4H3Z" />
      <circle cx="10" cy="9" r="2" />
      <path d="M6.5 14a3.5 3.5 0 0 1 7 0M16 8h2M16 11h2" />
    </AppIconFrame>
  );
}
