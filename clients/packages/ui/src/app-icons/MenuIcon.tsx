import { Utensils } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function MenuIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={Utensils} {...props} />
  );
}
