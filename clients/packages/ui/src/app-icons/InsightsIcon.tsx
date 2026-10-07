import { ChartNoAxesCombined } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function InsightsIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={ChartNoAxesCombined} {...props} />
  );
}
