import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function FlowIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M3.5 4h7v7h-7zM13.5 13h7v7h-7zM10.5 7.5H17V10M13.5 16.5H7V14" />
      <path d="m14.5 8.5 2.5 2 2.5-2M9.5 15.5l-2.5-2-2.5 2" />
      <path d="m7 4 3.5 3.5M17 13l3.5 3.5" strokeWidth="1.15" opacity=".5" />
    </AppIconFrame>
  );
}
