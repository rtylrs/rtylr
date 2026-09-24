import React, { useEffect, useMemo, useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "../lib/cn";

/**
 * All date-only values in this component are 'YYYY-MM-DD' strings, parsed and
 * formatted through UTC getters/setters exclusively (never local-timezone
 * ones). A calendar date has no time component; parsing it through a local
 * constructor shifts it by the viewer's UTC offset and renders the wrong day
 * for roughly half the world - the exact off-by-one this convention exists
 * to prevent.
 */
function parseISODate(date: string): Date {
  return new Date(`${date}T00:00:00Z`);
}

function toISODate(year: number, month: number, day: number): string {
  return `${String(year).padStart(4, "0")}-${String(month + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
}

function todayISODate(): string {
  const now = new Date();
  return toISODate(now.getFullYear(), now.getMonth(), now.getDate());
}

function addMonths(
  year: number,
  month: number,
  delta: number,
): { year: number; month: number } {
  const total = year * 12 + month + delta;
  return { year: Math.floor(total / 12), month: ((total % 12) + 12) % 12 };
}

function addDaysISO(date: string, delta: number): string {
  const d = parseISODate(date);
  d.setUTCDate(d.getUTCDate() + delta);
  return toISODate(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate());
}

function normalizeDateInput(value: string | undefined): string {
  return value?.trim() || todayISODate();
}

function isDateDisabled(
  date: string,
  minDate: string | undefined,
  maxDate: string | undefined,
): boolean {
  return (
    (minDate !== undefined && date < minDate) ||
    (maxDate !== undefined && date > maxDate)
  );
}

function clampDateToRange(
  date: string,
  minDate: string | undefined,
  maxDate: string | undefined,
): string {
  if (minDate !== undefined && date < minDate) return minDate;
  if (maxDate !== undefined && date > maxDate) return maxDate;
  return date;
}

/** Grid cells are built with the local Date constructor - this is calendar-
 * layout arithmetic (which weekday a given month starts on), not parsing of
 * a stored date string, so the UTC-anchoring rule above does not apply here. */
function buildMonthGrid(
  year: number,
  month: number,
  weekStartsOn: 0 | 1,
): string[] {
  const firstOfMonth = new Date(year, month, 1);
  const startOffset = (firstOfMonth.getDay() - weekStartsOn + 7) % 7;
  const gridStart = new Date(year, month, 1 - startOffset);
  return Array.from({ length: 42 }, (_, i) => {
    const d = new Date(gridStart);
    d.setDate(gridStart.getDate() + i);
    return toISODate(d.getFullYear(), d.getMonth(), d.getDate());
  });
}

function weekdayLabels(
  locale: string | undefined,
  weekStartsOn: 0 | 1,
): string[] {
  const formatter = new Intl.DateTimeFormat(locale, { weekday: "short" });
  // 1970-01-04 was a Sunday - a fixed, DST-free reference week to read labels from.
  return Array.from({ length: 7 }, (_, i) =>
    formatter.format(new Date(1970, 0, 4 + weekStartsOn + i)),
  );
}

export interface CalendarLabels {
  previousMonth?: string;
  nextMonth?: string;
  today?: string;
}

export interface CalendarProps {
  value?: string;
  onChange: (date: string) => void;
  minDate?: string;
  maxDate?: string;
  weekStartsOn?: 0 | 1;
  locale?: string;
  labels?: CalendarLabels;
  className?: string;
}

const Calendar = ({
  value,
  onChange,
  minDate,
  maxDate,
  weekStartsOn = 1,
  locale,
  labels,
  className,
}: CalendarProps): React.JSX.Element => {
  const initialDate = normalizeDateInput(value);
  const anchor = parseISODate(initialDate);
  const [view, setView] = useState(() => ({
    year: anchor.getUTCFullYear(),
    month: anchor.getUTCMonth(),
  }));
  const [focused, setFocused] = useState(() =>
    clampDateToRange(initialDate, minDate, maxDate),
  );
  const cellRefs = useRef<Record<string, HTMLButtonElement | null>>({});
  const focusRequestRef = useRef<string | null>(null);

  const grid = useMemo(
    () => buildMonthGrid(view.year, view.month, weekStartsOn),
    [view.year, view.month, weekStartsOn],
  );
  const enabledGrid = useMemo(
    () => grid.filter((iso) => !isDateDisabled(iso, minDate, maxDate)),
    [grid, minDate, maxDate],
  );
  const effectiveFocused = useMemo(() => {
    const preferred = focused || initialDate;
    return enabledGrid.includes(preferred)
      ? preferred
      : (enabledGrid[0] ?? preferred);
  }, [enabledGrid, focused, initialDate]);
  const weekdays = useMemo(
    () => weekdayLabels(locale, weekStartsOn),
    [locale, weekStartsOn],
  );
  const monthLabel = useMemo(
    () =>
      new Intl.DateTimeFormat(locale, {
        month: "long",
        year: "numeric",
        timeZone: "UTC",
      }).format(new Date(Date.UTC(view.year, view.month, 1))),
    [locale, view.year, view.month],
  );
  const dateFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat(locale, { dateStyle: "full", timeZone: "UTC" }),
    [locale],
  );
  const today = todayISODate();

  const isDisabled = (iso: string): boolean =>
    isDateDisabled(iso, minDate, maxDate);

  const syncedValueRef = useRef<string | null>(null);
  useEffect(() => {
    // The visible month follows the controlled value only. Clamping the view to
    // min/max would jump the grid forward the moment a constraint changes,
    // discarding the month the user is browsing; clamping applies to focus only.
    const next = normalizeDateInput(value);
    if (next === syncedValueRef.current) return;
    syncedValueRef.current = next;
    const nextDate = parseISODate(next);
    setView({ year: nextDate.getUTCFullYear(), month: nextDate.getUTCMonth() });
    setFocused(clampDateToRange(next, minDate, maxDate));
  }, [value, minDate, maxDate]);

  useEffect(() => {
    const target = focusRequestRef.current;
    if (!target) return;
    focusRequestRef.current = null;
    cellRefs.current[target]?.focus();
  }, [focused, view.year, view.month]);

  const goToMonth = (delta: number): void => {
    const currentDate = parseISODate(effectiveFocused);
    const target = addMonths(view.year, view.month, delta);
    const daysInTargetMonth = new Date(
      Date.UTC(target.year, target.month + 1, 0),
    ).getUTCDate();
    let next = clampDateToRange(
      toISODate(
        target.year,
        target.month,
        Math.min(currentDate.getUTCDate(), daysInTargetMonth),
      ),
      minDate,
      maxDate,
    );
    const targetGrid = buildMonthGrid(target.year, target.month, weekStartsOn);
    const targetEnabled = targetGrid.find(
      (iso) => !isDateDisabled(iso, minDate, maxDate),
    );
    if (!targetGrid.includes(next) || isDateDisabled(next, minDate, maxDate)) {
      next = targetEnabled ?? next;
    }
    focusRequestRef.current = next;
    setView(target);
    setFocused(next);
  };

  const moveFocus = (delta: number): void => {
    const base = effectiveFocused || focused || todayISODate();
    let next = addDaysISO(base, delta);
    if (isDateDisabled(next, minDate, maxDate)) {
      const step = delta < 0 ? -Math.abs(delta) : Math.abs(delta);
      let candidate = next;
      for (
        let attempts = 0;
        attempts < 3660 && isDateDisabled(candidate, minDate, maxDate);
        attempts++
      ) {
        candidate = addDaysISO(candidate, step);
      }
      if (!isDateDisabled(candidate, minDate, maxDate)) next = candidate;
    }
    focusRequestRef.current = next;
    setFocused(next);
    const nextDate = parseISODate(next);
    setView({ year: nextDate.getUTCFullYear(), month: nextDate.getUTCMonth() });
  };

  const handleKeyDown = (
    event: React.KeyboardEvent<HTMLButtonElement>,
  ): void => {
    switch (event.key) {
      case "ArrowLeft":
        event.preventDefault();
        moveFocus(-1);
        break;
      case "ArrowRight":
        event.preventDefault();
        moveFocus(1);
        break;
      case "ArrowUp":
        event.preventDefault();
        moveFocus(-7);
        break;
      case "ArrowDown":
        event.preventDefault();
        moveFocus(7);
        break;
      case "PageUp":
        event.preventDefault();
        goToMonth(-1);
        break;
      case "PageDown":
        event.preventDefault();
        goToMonth(1);
        break;
      case "Home": {
        event.preventDefault();
        const offsetFromWeekStart =
          (parseISODate(effectiveFocused).getUTCDay() - weekStartsOn + 7) % 7;
        moveFocus(-offsetFromWeekStart);
        break;
      }
      case "End": {
        event.preventDefault();
        const offsetToWeekEnd =
          6 -
          ((parseISODate(effectiveFocused).getUTCDay() - weekStartsOn + 7) % 7);
        moveFocus(offsetToWeekEnd);
        break;
      }
    }
  };

  return (
    <div className={cn("w-[280px] p-3", className)}>
      <div className="flex items-center justify-between pb-2">
        <button
          type="button"
          onClick={() => goToMonth(-1)}
          aria-label={labels?.previousMonth ?? "Previous month"}
          className="flex h-8 w-8 items-center justify-center rounded-md text-muted hover:bg-muted/15 hover:text-inherit"
        >
          <ChevronLeft size={16} className="rtl:rotate-180" />
        </button>
        <span className="text-sm font-semibold">{monthLabel}</span>
        <button
          type="button"
          onClick={() => goToMonth(1)}
          aria-label={labels?.nextMonth ?? "Next month"}
          className="flex h-8 w-8 items-center justify-center rounded-md text-muted hover:bg-muted/15 hover:text-inherit"
        >
          <ChevronRight size={16} className="rtl:rotate-180" />
        </button>
      </div>
      <div className="grid grid-cols-7 gap-y-0.5 text-center">
        {weekdays.map((label, i) => (
          <span key={i} className="py-1 text-[11px] font-medium text-muted">
            {label}
          </span>
        ))}
        {grid.map((iso) => {
          const date = parseISODate(iso);
          const inMonth = date.getUTCMonth() === view.month;
          const selected = iso === value;
          const isToday = iso === today;
          const disabled = isDisabled(iso);
          return (
            <button
              key={iso}
              ref={(el) => {
                cellRefs.current[iso] = el;
              }}
              type="button"
              role="gridcell"
              disabled={disabled}
              tabIndex={!disabled && iso === effectiveFocused ? 0 : -1}
              aria-current={isToday ? "date" : undefined}
              aria-selected={selected}
              aria-label={dateFormatter.format(date)}
              onFocus={() =>
                setFocused((current) => (current === iso ? current : iso))
              }
              onKeyDown={handleKeyDown}
              onClick={() => onChange(iso)}
              className={cn(
                "m-auto flex h-8 w-8 items-center justify-center rounded-full text-[13px] tabular-nums transition-colors",
                "disabled:cursor-not-allowed disabled:opacity-35",
                !inMonth && "text-muted/50",
                inMonth && !selected && "text-inherit hover:bg-muted/15",
                isToday && !selected && "font-semibold text-accent",
                selected &&
                  "bg-accent font-semibold text-accent-foreground hover:opacity-90",
              )}
            >
              {date.getUTCDate()}
            </button>
          );
        })}
      </div>
      {labels?.today && (
        <button
          type="button"
          onClick={() => !isDisabled(today) && onChange(today)}
          disabled={isDisabled(today)}
          className="mt-2 w-full rounded-md py-1.5 text-center text-[12px] font-medium text-accent hover:bg-muted/15 disabled:cursor-not-allowed disabled:opacity-35"
        >
          {labels.today}
        </button>
      )}
    </div>
  );
};

export { Calendar, toISODate, todayISODate };
