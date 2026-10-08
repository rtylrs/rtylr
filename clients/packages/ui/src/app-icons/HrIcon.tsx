import { UsersRound } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function HrIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={UsersRound} {...props} />
  );
}
