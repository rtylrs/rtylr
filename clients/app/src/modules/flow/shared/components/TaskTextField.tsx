import React, { useEffect, useRef, useState } from 'react';

/**
 * A text input (or textarea) that saves once, on blur, and only when the text
 * changed. While it has focus, a new value from the server does not overwrite
 * what is being typed. Used by the task panel's text properties and by custom
 * fields wherever they are edited.
 */
export function TaskTextField({
  value,
  onSave,
  multiline,
  wrap,
  ...props
}: {
  value: string;
  onSave: (value: string) => void;
  multiline?: boolean;
  /** One logical line that wraps instead of being cut off. */
  wrap?: boolean;
  type?: React.HTMLInputTypeAttribute;
  min?: number;
  step?: string;
  required?: boolean;
  disabled?: boolean;
  placeholder?: string;
  rows?: number;
  className?: string;
  style?: React.CSSProperties;
  'aria-label'?: string;
}) {
  const [draft, setDraft] = useState(value);
  const focused = useRef(false);
  useEffect(() => {
    if (!focused.current) setDraft(value);
  }, [value]);
  const handlers = {
    value: draft,
    onFocus: () => {
      focused.current = true;
    },
    onChange: (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
      setDraft(event.target.value),
    onBlur: () => {
      focused.current = false;
      if (props.required && !draft.trim()) {
        setDraft(value);
      } else if (draft !== value) {
        onSave(draft);
      }
    },
  };
  if (wrap) {
    // The hidden copy shares the textarea's grid cell and sets its height, so
    // the field is always as tall as its wrapped text.
    return (
      <div className="grid grid-cols-1">
        <div
          aria-hidden
          className={`${props.className ?? ''} flow-wrap-mirror invisible col-start-1 row-start-1 whitespace-pre-wrap break-words`}
        >
          {draft}{' '}
        </div>
        <textarea
          {...props}
          {...handlers}
          rows={1}
          onChange={(event) => setDraft(event.target.value.replace(/\s*\n\s*/g, ' '))}
          onKeyDown={(event) => {
            if (event.key === 'Enter') {
              event.preventDefault();
              event.currentTarget.blur();
            }
          }}
          className={`${props.className ?? ''} col-start-1 row-start-1 overflow-hidden`}
        />
      </div>
    );
  }
  return multiline ? <textarea {...props} {...handlers} /> : <input {...props} {...handlers} />;
}

export default TaskTextField;
