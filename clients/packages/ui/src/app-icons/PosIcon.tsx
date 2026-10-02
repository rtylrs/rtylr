import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function PosIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M4 4.5h16v10H4zM4 11.5h16M8.5 18h7M12 14.5V18M3 20.5h18" />
      <path d="M14.5 4.5v4H20" />
      <path d="m14.5 4.5 5.5 4" strokeWidth="1.15" opacity=".5" />
      <path d="M7 8h3" />
    </AppIconFrame>
  );
}
