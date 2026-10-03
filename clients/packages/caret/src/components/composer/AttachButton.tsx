import React, { useRef } from "react";
import { Paperclip } from "lucide-react";
import IconButton from "../../ui/IconButton";
import Tooltip from "../../ui/Tooltip";

export interface AttachButtonProps {
  onFiles?: (files: File[]) => void;
  accept?: string;
}

/** Attaches files to the composer. Uses a hidden file input. */
export const AttachButton: React.FC<AttachButtonProps> = ({
  onFiles,
  accept = "image/jpeg,image/png,image/webp",
}) => {
  const ref = useRef<HTMLInputElement>(null);

  const onChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const list = e.target.files;
    if (!list) return;
    const files = Array.from(list);
    onFiles?.(files);
    e.target.value = "";
  };

  return (
    <>
      <input
        ref={ref}
        type="file"
        multiple
        accept={accept}
        onChange={onChange}
        style={{ display: "none" }}
      />
      <Tooltip label="Attach files">
        <IconButton
          className="caret-composer-icon"
          size="md"
          label="Attach files"
          onClick={() => ref.current?.click()}
        >
          <Paperclip size={14} />
        </IconButton>
      </Tooltip>
    </>
  );
};

export default AttachButton;
