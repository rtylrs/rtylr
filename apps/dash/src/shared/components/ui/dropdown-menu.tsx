import React, { createContext, useContext, useState, useRef, useEffect } from 'react';
import { cn } from '../../lib/utils';

interface DropdownContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const DropdownContext = createContext<DropdownContextValue>({ open: false, setOpen: () => {} });

const DropdownMenu: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [open]);

  return (
    <DropdownContext.Provider value={{ open, setOpen }}>
      <div ref={ref} className="relative w-full">
        {children}
      </div>
    </DropdownContext.Provider>
  );
};

const DropdownMenuTrigger: React.FC<{
  children: React.ReactElement<{ onClick?: (e: React.MouseEvent) => void }>;
  asChild?: boolean;
}> = ({ children }) => {
  const { open, setOpen } = useContext(DropdownContext);
  return React.cloneElement(children, {
    onClick: (e: React.MouseEvent) => {
      e.stopPropagation();
      children.props.onClick?.(e);
      setOpen(!open);
    },
  });
};

interface DropdownMenuContentProps extends React.HTMLAttributes<HTMLDivElement> {
  align?: 'start' | 'end' | 'center';
  side?: 'top' | 'bottom';
}

const DropdownMenuContent: React.FC<DropdownMenuContentProps> = ({
  children,
  className,
  align = 'start',
  side = 'bottom',
}) => {
  const { open } = useContext(DropdownContext);
  if (!open) return null;

  const alignClasses = { start: 'start-0', end: 'end-0', center: 'left-1/2 -translate-x-1/2' };
  const sideClasses = { bottom: 'top-full mt-1', top: 'bottom-full mb-1' };

  return (
    <div
      role="menu"
      className={cn(
        'absolute z-50 min-w-[160px] rounded-xl py-1',
        alignClasses[align],
        sideClasses[side],
        className
      )}
      style={{
        background: 'var(--dash-card)',
        border: '1px solid var(--dash-border)',
        boxShadow: 'var(--dash-shadow)',
      }}
    >
      {children}
    </div>
  );
};

interface DropdownMenuItemProps extends React.HTMLAttributes<HTMLDivElement> {
  destructive?: boolean;
}

const DropdownMenuItem: React.FC<DropdownMenuItemProps> = ({
  className,
  destructive,
  onClick,
  onKeyDown,
  style,
  ...props
}) => {
  const { setOpen } = useContext(DropdownContext);
  return (
    <div
      role="menuitem"
      tabIndex={0}
      className={cn(
        'flex items-center gap-2 px-3 py-2 text-sm cursor-pointer transition-colors select-none rounded-lg mx-1 focus:outline-none focus:bg-[var(--dash-hover)]',
        destructive
          ? 'text-red-500 hover:bg-red-500/8 focus:bg-red-500/8'
          : 'hover:bg-[var(--dash-hover)]',
        className
      )}
      style={{ color: destructive ? undefined : 'var(--dash-text)', ...style }}
      onClick={(e) => {
        onClick?.(e);
        if (!e.defaultPrevented) {
          setOpen(false);
        }
      }}
      onKeyDown={(e) => {
        onKeyDown?.(e);
        if ((e.key === 'Enter' || e.key === ' ') && !e.defaultPrevented) {
          e.preventDefault();
          onClick?.(e as unknown as React.MouseEvent<HTMLDivElement>);
          setOpen(false);
        }
      }}
      {...props}
    />
  );
};

const DropdownMenuSeparator: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  ...props
}) => (
  <div
    className={cn('h-px my-1 mx-1', className)}
    style={{ background: 'var(--dash-border)' }}
    {...props}
  />
);

const DropdownMenuLabel: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  ...props
}) => (
  <div
    className={cn('px-3 py-1.5 text-xs font-semibold uppercase tracking-wider', className)}
    style={{ color: 'var(--dash-text-dim)' }}
    {...props}
  />
);

export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
};
