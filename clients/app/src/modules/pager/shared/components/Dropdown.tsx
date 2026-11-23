import React, { useEffect, useRef, useState } from 'react';
import { Check, ChevronDown } from 'lucide-react';

export interface DropdownOption<T extends string> {
  value: T;
  label: string;
}

interface DropdownProps<T extends string> {
  /** The field's id, for its <label htmlFor>. */
  id: string;
  /** The <label>'s own id, which also names the open list. */
  labelId: string;
  value: T;
  options: readonly DropdownOption<T>[];
  onChange: (value: T) => void;
  disabled?: boolean;
  'aria-describedby'?: string;
}

/**
 * Rennili's select: a button showing the choice and a list under it, in the
 * board's own colours. It follows the select-only combobox pattern, so focus
 * stays on the button and the arrow keys move through the list.
 */
export function Dropdown<T extends string>({
  id,
  labelId,
  value,
  options,
  onChange,
  disabled = false,
  'aria-describedby': describedBy,
}: DropdownProps<T>): React.JSX.Element {
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);
  const root = useRef<HTMLDivElement>(null);
  const listId = `${id}-list`;
  const optionId = (index: number): string => `${id}-option-${index}`;
  const chosen = Math.max(
    options.findIndex((option) => option.value === value),
    0
  );

  useEffect(() => {
    if (!open) return undefined;
    const closeOutside = (event: PointerEvent): void => {
      if (event.target instanceof Node && !root.current?.contains(event.target)) setOpen(false);
    };
    document.addEventListener('pointerdown', closeOutside);
    return () => document.removeEventListener('pointerdown', closeOutside);
  }, [open]);

  const show = (): void => {
    setActive(chosen);
    setOpen(true);
  };

  const choose = (index: number): void => {
    setOpen(false);
    const option = options[index];
    if (option && option.value !== value) onChange(option.value);
  };

  const onKeyDown = (event: React.KeyboardEvent<HTMLButtonElement>): void => {
    if (!open) {
      if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(event.key)) {
        event.preventDefault();
        show();
      }
      return;
    }
    const last = options.length - 1;
    const moves: Record<string, () => void> = {
      ArrowDown: () => setActive((index) => Math.min(index + 1, last)),
      ArrowUp: () => setActive((index) => Math.max(index - 1, 0)),
      Home: () => setActive(0),
      End: () => setActive(last),
      Enter: () => choose(active),
      ' ': () => choose(active),
      Escape: () => setOpen(false),
    };
    const move = moves[event.key];
    if (move) {
      event.preventDefault();
      move();
    } else if (event.key === 'Tab') {
      setOpen(false);
    }
  };

  return (
    <div ref={root} className="relative w-full">
      <button
        type="button"
        id={id}
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listId}
        aria-activedescendant={open ? optionId(active) : undefined}
        aria-describedby={describedBy}
        disabled={disabled}
        onClick={() => (open ? setOpen(false) : show())}
        onKeyDown={onKeyDown}
        className="flex h-11 w-full items-center justify-between gap-2 rounded-lg ps-3 pe-2.5 text-start text-sm transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--dash-border-focus)] disabled:cursor-not-allowed disabled:opacity-50"
        style={{
          background: 'var(--dash-input-bg)',
          color: 'var(--dash-text)',
          border: `1px solid ${open ? 'var(--dash-text-muted)' : 'var(--dash-border)'}`,
        }}
      >
        <span className="truncate">{options[chosen]?.label}</span>
        <ChevronDown
          size={16}
          aria-hidden="true"
          className={`flex-none transition-transform ${open ? 'rotate-180' : ''}`}
          style={{ color: 'var(--dash-text-dim)' }}
        />
      </button>
      {open && (
        <ul
          id={listId}
          role="listbox"
          aria-labelledby={labelId}
          className="absolute inset-x-0 top-[calc(100%+6px)] z-30 max-h-64 overflow-auto rounded-xl p-1"
          style={{
            background: 'var(--dash-surface)',
            border: '1px solid var(--dash-border)',
            boxShadow: '0 16px 36px -18px rgba(20, 27, 23, 0.35)',
          }}
        >
          {options.map((option, index) => {
            const selected = index === chosen;
            return (
              <li
                key={option.value}
                id={optionId(index)}
                role="option"
                aria-selected={selected}
                // Keeps focus on the button, so the list stays open for the click.
                onMouseDown={(event) => event.preventDefault()}
                onMouseEnter={() => setActive(index)}
                onClick={() => choose(index)}
                className={`flex min-h-11 cursor-pointer items-center justify-between gap-3 rounded-lg px-3 py-2 text-sm ${selected ? 'font-semibold' : ''}`}
                style={{
                  color: 'var(--dash-text)',
                  background: index === active ? 'var(--dash-hover)' : 'transparent',
                }}
              >
                <span>{option.label}</span>
                {selected && <Check size={16} strokeWidth={2.4} aria-hidden="true" />}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
