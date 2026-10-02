import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function RecruitIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M8 6V3.5h8V6M3 6h18v14H3ZM3 11.5l4 2M21 11.5l-4 2" />
      <circle cx="12" cy="11" r="2" />
      <path d="M8.5 17a3.5 3.5 0 0 1 7 0" />
    </AppIconFrame>
  );
}
