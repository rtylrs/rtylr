import React, { useEffect, useLayoutEffect, useRef, useState } from "react";
import { CalendarDays, X } from "lucide-react";
import { cn } from "../lib/cn";
import { Calendar, type CalendarLabels } from "./Calendar";

export interface DatePickerLabels extends CalendarLabels {
  placeholder?: string;
  clear?: string;
}

export interface DatePickerProps {
  id?: string;
  value?: string;
  onChange: (date: string) => void;
  minDate?: string;
  maxDate?: string;
  weekStartsOn?: 0 | 1;
  locale?: string;
  disabled?: boolean;
  readOnly?: boolean;
  error?: boolean;
  clearable?: boolean;
  labels?: DatePickerLabels;
  formatValue?: (date: string) => string;
  className?: string;
  wrapperClassName?: string;
  "aria-label"?: string;
  "aria-labelledby"?: string;
  "aria-describedby"?: string;
  "aria-invalid"?: boolean;
}

const POPUP_GAP = 4;
const VIEWPORT_GUTTER = 12;

/**
 * Where the calendar panel goes, in viewport px: below the field, or above it
 * when the room below would cut it off, kept on screen and aligned to the
 * field's start edge.
 */
export function placePopup(
  anchor: { top: number; bottom: number; left: number; right: number },
  size: { width: number; height: number },
  viewport: { width: number; height: number },
  direction: "ltr" | "rtl",
): { top: number; left: number } {
  const aligned = direction === "rtl" ? anchor.right - size.width : anchor.left;
  const left = Math.max(
    VIEWPORT_GUTTER,
    Math.min(aligned, viewport.width - size.width - VIEWPORT_GUTTER),
  );
  const below = anchor.bottom + POPUP_GAP;
  const top =
    below + size.height <= viewport.height - VIEWPORT_GUTTER
      ? below
      : Math.max(VIEWPORT_GUTTER, anchor.top - size.height - POPUP_GAP);
  return { top, left };
}

const DatePicker = ({
  id,
  value,
  onChange,
  minDate,
  maxDate,
  weekStartsOn,
  locale,
  disabled,
  readOnly,
  error,
  clearable,
  labels,
  formatValue,
  className,
  wrapperClassName,
  "aria-label": ariaLabel,
  "aria-labelledby": ariaLabelledBy,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
}: DatePickerProps): React.JSX.Element => {
  const inert = disabled || readOnly;
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const popupRef = useRef<HTMLDivElement>(null);

  // The panel opens in the top layer, positioned against the field, so a
  // scrolling dialog or a footer can no longer clip it.
  useLayoutEffect(() => {
    const popup = popupRef.current;
    const trigger = triggerRef.current;
    if (!open || !popup || !trigger) return;
    popup.showPopover?.();
    const position = (): void => {
      const { top, left } = placePopup(
        trigger.getBoundingClientRect(),
        { width: popup.offsetWidth, height: popup.offsetHeight },
        { width: window.innerWidth, height: window.innerHeight },
        getComputedStyle(trigger).direction === "rtl" ? "rtl" : "ltr",
      );
      popup.style.top = `${top}px`;
      popup.style.left = `${left}px`;
    };
    position();
    window.addEventListener("resize", position);
    window.addEventListener("scroll", position, true);
    return () => {
      window.removeEventListener("resize", position);
      window.removeEventListener("scroll", position, true);
    };
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: MouseEvent | TouchEvent): void => {
      const target = event.target;
      const insidePicker =
        target instanceof Node && containerRef.current?.contains(target);
      if (!insidePicker) setOpen(false);
    };
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape") {
        setOpen(false);
        triggerRef.current?.focus();
        requestAnimationFrame(() => triggerRef.current?.focus());
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("touchstart", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("touchstart", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  const display = value
    ? formatValue
      ? formatValue(value)
      : new Intl.DateTimeFormat(locale, {
          dateStyle: "medium",
          timeZone: "UTC",
        }).format(new Date(`${value}T00:00:00Z`))
    : (labels?.placeholder ?? "Select date");

  return (
    <div
      ref={containerRef}
      className={cn("relative inline-block", wrapperClassName)}
    >
      <button
        ref={triggerRef}
        id={id}
        type="button"
        disabled={inert}
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-label={ariaLabel}
        aria-labelledby={ariaLabelledBy}
        aria-describedby={ariaDescribedBy}
        aria-invalid={ariaInvalid}
        data-value={value ?? ""}
        onClick={() => setOpen((o) => !o)}
        className={cn(
          "flex h-10 min-h-[44px] w-full items-center justify-between gap-2 rounded-lg border px-3.5 text-sm outline-none transition-colors",
          "border-black/10 bg-white/[0.06] text-inherit dark:border-white/10",
          "hover:border-accent/60 focus-visible:border-accent",
          "disabled:cursor-not-allowed disabled:opacity-50",
          error && "border-destructive focus-visible:border-destructive",
          !value && "text-muted/70",
          className,
        )}
      >
        <span className="flex items-center gap-2 truncate">
          <CalendarDays size={15} className="shrink-0 text-muted" />
          {display}
        </span>
        {clearable && value && !inert && (
          <span
            role="button"
            tabIndex={0}
            aria-label={labels?.clear ?? "Clear date"}
            onClick={(e) => {
              e.stopPropagation();
              onChange("");
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                e.stopPropagation();
                onChange("");
              }
            }}
            className="shrink-0 rounded p-0.5 text-muted hover:text-inherit"
          >
            <X size={13} />
          </span>
        )}
      </button>
      {open && !inert && (
        <div
          ref={popupRef}
          role="dialog"
          aria-modal="false"
          popover="manual"
          className="fixed inset-auto z-50 m-0 overflow-visible rounded-xl border border-black/10 bg-white p-0 text-inherit shadow-dropdown dark:border-white/10 dark:bg-admin"
        >
          <Calendar
            value={value}
            onChange={(date) => {
              onChange(date);
              setOpen(false);
              triggerRef.current?.focus();
              requestAnimationFrame(() => triggerRef.current?.focus());
            }}
            minDate={minDate}
            maxDate={maxDate}
            weekStartsOn={weekStartsOn}
            locale={locale}
            labels={labels}
          />
        </div>
      )}
    </div>
  );
};

export { DatePicker };
