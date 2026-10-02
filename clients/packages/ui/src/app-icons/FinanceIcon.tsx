import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function FinanceIcon(props: AppIconProps) {
  return (
    <AppIconFrame {...props}>
      <path d="M5 3.5h11l3 3V21l-3-1.5-4 1.5-4-1.5L5 21Z" />
      <path d="M15.5 3.5V7H19M8.5 9h3M8.5 12h7M8.5 15h7" />
      <path d="M14.5 9h1" strokeWidth="2.5" />
    </AppIconFrame>
  );
}
