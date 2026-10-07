import { ContactRound } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function CrmIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={ContactRound} {...props} />
  );
}
