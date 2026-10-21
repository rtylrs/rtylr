import React from "react";
import { FileText } from "lucide-react";
import { caretTokens } from "../../ui/tokens";

/** An attached text file, shown by its name. */
export const FileChip: React.FC<{ name: string }> = ({ name }) => (
  <div
    title={name}
    style={{
      display: "inline-flex",
      alignItems: "center",
      gap: 6,
      maxWidth: 200,
      height: 40,
      padding: "0 10px",
      borderRadius: 9,
      border: `1px solid ${caretTokens.border}`,
      background: caretTokens.surfaceHigh,
      color: caretTokens.textMuted,
      fontSize: 12,
    }}
  >
    <FileText size={14} style={{ flex: "0 0 auto" }} />
    <span
      style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}
    >
      {name}
    </span>
  </div>
);

export default FileChip;
