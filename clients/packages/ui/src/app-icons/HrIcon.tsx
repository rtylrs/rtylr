import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function HrIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M8 4h8l4 4v12H4V4Z" />
      <path d="M16 4v4h4M9 2v4" />
      <circle cx="12" cy="11" r="2" />
      <path d="M8.5 17a3.5 3.5 0 0 1 7 0" />
    </AppIconFrame>
  );
}
