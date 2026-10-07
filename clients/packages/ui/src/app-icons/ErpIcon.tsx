import { Boxes } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function ErpIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={Boxes} {...props} />
  );
}
