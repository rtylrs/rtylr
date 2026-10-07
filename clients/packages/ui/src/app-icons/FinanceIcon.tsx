import { ReceiptText } from "lucide-react";
import { AppIconFrame, type AppIconProps } from "./AppIconFrame";

export function FinanceIcon(props: AppIconProps) {
  return (
    <AppIconFrame icon={ReceiptText} {...props} />
  );
}
