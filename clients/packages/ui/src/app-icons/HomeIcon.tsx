import { House } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function HomeIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={House} {...props} />
  );
}
