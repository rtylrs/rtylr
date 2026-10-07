import { ShoppingCart } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function PosIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={ShoppingCart} {...props} />
  );
}
