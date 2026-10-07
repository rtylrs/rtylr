import { Workflow } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function FlowIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={Workflow} {...props} />
  );
}
