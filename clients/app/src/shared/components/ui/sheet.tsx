import React, { createContext, useContext, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { createPortal } from 'react-dom';
import { X } from 'lucide-react';
import { useThemeScope } from '@voxire/ui';
import { cn } from '../../lib/utils';

interface SheetContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const SheetContext = createContext<SheetContextValue>({ open: false, setOpen: () => {} });

interface SheetProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children: React.ReactNode;
}

const Sheet: React.FC<SheetProps> = ({ open: controlledOpen, onOpenChange, children }) => {
  const [internalOpen, setInternalOpen] = useState(false);
  const isControlled = controlledOpen !== undefined;
  const open = isControlled ? controlledOpen : internalOpen;

  const setOpen = (value: boolean) => {
    if (!isControlled) setInternalOpen(value);
    onOpenChange?.(value);
  };

  return <SheetContext.Provider value={{ open, setOpen }}>{children}</SheetContext.Provider>;
};

const SheetTrigger: React.FC<{
  children: React.ReactElement<{ onClick?: (e: React.MouseEvent) => void }>;
}> = ({ children }) => {
  const { setOpen } = useContext(SheetContext);
  return React.cloneElement(children, {
    onClick: (e: React.MouseEvent) => {
      children.props.onClick?.(e);
      setOpen(true);
    },
  });
};

interface SheetContentProps extends React.HTMLAttributes<HTMLDivElement> {
  side?: 'right' | 'left';
}

const SheetContent: React.FC<SheetContentProps> = ({ children, className, side = 'right' }) => {
  const { t } = useTranslation();
  const { open, setOpen } = useContext(SheetContext);
  const scope = useThemeScope();

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    if (open) document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [open, setOpen]);

  if (!open) return null;

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex"
      style={scope}
      data-theme-scope={scope ? '' : undefined}
    >
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={() => setOpen(false)}
      />
      <div
        className={cn(
          'relative z-10 flex flex-col shadow-xl w-full max-w-md ml-auto h-full',
          side === 'left' && 'ml-0 mr-auto',
          className
        )}
        style={{
          background: 'var(--dash-card)',
          borderLeft: side !== 'left' ? '1px solid var(--dash-border)' : undefined,
          borderRight: side === 'left' ? '1px solid var(--dash-border)' : undefined,
        }}
      >
        <button
          type="button"
          aria-label={t('common.closePanel')}
          className="absolute right-4 top-4 rounded-lg p-1 text-gray-400 hover:text-gray-600 hover:bg-black/5 transition-colors"
          onClick={() => setOpen(false)}
        >
          <X size={16} />
        </button>
        {children}
      </div>
    </div>,
    document.body
  );
};

const SheetHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  style,
  ...props
}) => (
  <div
    className={cn('flex flex-col gap-1.5 p-6 pb-4', className)}
    style={{ borderBottom: '1px solid var(--dash-border)', ...style }}
    {...props}
  />
);

const SheetTitle: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  className,
  ...props
}) => <h2 className={cn('text-lg font-semibold text-primary', className)} {...props} />;

const SheetDescription: React.FC<React.HTMLAttributes<HTMLParagraphElement>> = ({
  className,
  ...props
}) => <p className={cn('text-sm text-gray-500', className)} {...props} />;

const SheetFooter: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  style,
  ...props
}) => (
  <div
    className={cn('flex items-center justify-end gap-2 p-6 pt-4 mt-auto', className)}
    style={{ borderTop: '1px solid var(--dash-border)', ...style }}
    {...props}
  />
);

const SheetClose: React.FC<{
  children: React.ReactElement<{ onClick?: (e: React.MouseEvent) => void }>;
}> = ({ children }) => {
  const { setOpen } = useContext(SheetContext);
  return React.cloneElement(children, {
    onClick: (e: React.MouseEvent) => {
      children.props.onClick?.(e);
      setOpen(false);
    },
  });
};

export {
  Sheet,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
  SheetClose,
};
