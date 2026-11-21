import React, { useRef } from 'react';
import { Check, ChevronDown } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@voxire/ui';

export interface FlowSelectOption {
  value: string;
  label: string;
  /** Rendered before the label, in the trigger and in the list. */
  icon?: React.ReactNode;
}

interface FlowSelectProps {
  value: string;
  options: FlowSelectOption[];
  onChange: (value: string) => void;
  /** Accessible name of the control. The current value is appended for assistive tech. */
  label: string;
  disabled?: boolean;
  /** Shown when no option matches `value`. */
  placeholder?: string;
  /** Rendered before the trigger text, e.g. an avatar. Overrides the option icon. */
  leading?: React.ReactNode;
  /** Muted text colour for secondary controls. */
  muted?: boolean;
  className?: string;
  /** A command after the options, such as creating a new one. */
  action?: { label: string; icon?: React.ReactNode; onSelect: () => void };
}

/**
 * A styled replacement for the native `<select>` in Flow's properties rail.
 *
 * The OS draws a native select's open list, so it ignores the theme and cannot
 * be restyled. This keeps the same contract (a value, a list of options, one
 * change callback) on top of the shared DropdownMenu, which supplies keyboard
 * navigation, focus return, collision handling and the theme scope.
 */
export const FlowSelect: React.FC<FlowSelectProps> = ({
  value,
  options,
  onChange,
  label,
  disabled = false,
  placeholder,
  leading,
  muted = false,
  className,
  action,
}) => {
  const selected = options.find((option) => option.value === value);
  // The action runs once the menu has closed and given focus back, so a
  // dialog it opens does not fight the menu's focus trap.
  const actionChosen = useRef(false);
  const text = selected?.label ?? placeholder ?? '';

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        disabled={disabled}
        aria-label={text ? `${label}: ${text}` : label}
        className={`flow-select-trigger group flex w-full min-w-0 items-center gap-2 rounded-md px-2.5 py-1.5 text-start text-flow-meta transition-colors hover:bg-[var(--flow-hover)] disabled:cursor-default disabled:opacity-70 disabled:hover:bg-[var(--flow-input-bg)] ${className ?? ''}`}
        style={{
          color: muted || !selected ? 'var(--flow-text-muted)' : 'var(--flow-text)',
          background: 'var(--flow-input-bg)',
          border: '1px solid var(--flow-border-control)',
        }}
      >
        {leading ?? selected?.icon}
        <span className="min-w-0 flex-1 truncate">{text}</span>
        <ChevronDown
          size={13}
          aria-hidden="true"
          className="shrink-0 opacity-60 transition-transform group-data-[state=open]:rotate-180"
        />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        onCloseAutoFocus={(event) => {
          if (!actionChosen.current || !action) return;
          actionChosen.current = false;
          event.preventDefault();
          action.onSelect();
        }}
        align="start"
        className="max-h-[260px] min-w-[var(--radix-dropdown-menu-trigger-width)] overflow-y-auto"
        style={{
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        {options.map((option) => {
          const isSelected = option.value === value;
          return (
            <DropdownMenuItem
              key={option.value}
              aria-current={isSelected ? 'true' : undefined}
              onSelect={() => {
                if (!isSelected) onChange(option.value);
              }}
              className="text-flow-meta"
              style={{ color: 'var(--flow-text)' }}
            >
              {option.icon}
              <span className="min-w-0 flex-1 truncate">{option.label}</span>
              {isSelected && <Check size={13} aria-hidden="true" className="shrink-0" />}
            </DropdownMenuItem>
          );
        })}
        {action && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onSelect={() => {
                actionChosen.current = true;
              }}
              className="text-flow-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {action.icon}
              <span className="min-w-0 flex-1 truncate">{action.label}</span>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

interface FlowMultiSelectProps {
  values: readonly string[];
  options: FlowSelectOption[];
  onChange: (values: string[]) => void;
  label: string;
  disabled?: boolean;
  /** Shown when nothing is picked. */
  placeholder?: string;
  className?: string;
}

/**
 * FlowSelect for several values: the same trigger and menu, with a check per
 * picked option. Picking keeps the menu open so several can be toggled in one
 * go; the trigger lists what is picked, in option order.
 */
export const FlowMultiSelect: React.FC<FlowMultiSelectProps> = ({
  values,
  options,
  onChange,
  label,
  disabled = false,
  placeholder,
  className,
}) => {
  const picked = options.filter((option) => values.includes(option.value));
  const text = picked.length > 0 ? picked.map((option) => option.label).join(', ') : placeholder;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        disabled={disabled}
        aria-label={text ? `${label}: ${text}` : label}
        className={`flow-select-trigger group flex w-full min-w-0 items-center gap-2 rounded-md px-2.5 py-1.5 text-start text-flow-meta transition-colors hover:bg-[var(--flow-hover)] disabled:cursor-default disabled:opacity-70 disabled:hover:bg-[var(--flow-input-bg)] ${className ?? ''}`}
        style={{
          color: picked.length > 0 ? 'var(--flow-text)' : 'var(--flow-text-muted)',
          background: 'var(--flow-input-bg)',
          border: '1px solid var(--flow-border-control)',
        }}
      >
        {picked.length === 1 && picked[0]?.icon}
        <span className="min-w-0 flex-1 truncate">{text ?? ''}</span>
        <ChevronDown
          size={13}
          aria-hidden="true"
          className="shrink-0 opacity-60 transition-transform group-data-[state=open]:rotate-180"
        />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="start"
        className="max-h-[260px] min-w-[var(--radix-dropdown-menu-trigger-width)] overflow-y-auto"
        style={{
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        {options.map((option) => {
          const isPicked = values.includes(option.value);
          return (
            <DropdownMenuItem
              key={option.value}
              role="menuitemcheckbox"
              aria-checked={isPicked}
              onSelect={(event) => {
                event.preventDefault();
                onChange(
                  isPicked
                    ? values.filter((v) => v !== option.value)
                    : options
                        .map((o) => o.value)
                        .filter((v) => v === option.value || values.includes(v))
                );
              }}
              className="text-flow-meta"
              style={{ color: 'var(--flow-text)' }}
            >
              {option.icon}
              <span className="min-w-0 flex-1 truncate">{option.label}</span>
              {isPicked && <Check size={13} aria-hidden="true" className="shrink-0" />}
            </DropdownMenuItem>
          );
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export default FlowSelect;
