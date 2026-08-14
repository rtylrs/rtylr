import React, { createContext, useContext, useState } from 'react';
import { cn } from '../../lib/utils';

interface TabsContextValue {
  value: string;
  onValueChange: (value: string) => void;
}

const TabsContext = createContext<TabsContextValue>({ value: '', onValueChange: () => {} });

interface TabsProps extends React.HTMLAttributes<HTMLDivElement> {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
}

const Tabs: React.FC<TabsProps> = ({
  value: controlledValue,
  defaultValue = '',
  onValueChange,
  children,
  className,
  ...props
}) => {
  const [internalValue, setInternalValue] = useState(defaultValue);
  const isControlled = controlledValue !== undefined;
  const value = isControlled ? controlledValue : internalValue;

  const handleChange = (v: string) => {
    if (!isControlled) setInternalValue(v);
    onValueChange?.(v);
  };

  return (
    <TabsContext.Provider value={{ value, onValueChange: handleChange }}>
      <div className={cn('', className)} {...props}>
        {children}
      </div>
    </TabsContext.Provider>
  );
};

const TabsList: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  style,
  ...props
}) => (
  <div
    role="tablist"
    className={cn('inline-flex items-center rounded-lg p-1 gap-0.5', className)}
    style={{ background: 'var(--dash-hover)', ...style }}
    {...props}
  />
);

interface TabsTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  value: string;
}

const TabsTrigger: React.FC<TabsTriggerProps> = ({
  value,
  className,
  style,
  children,
  ...props
}) => {
  const { value: activeValue, onValueChange } = useContext(TabsContext);
  const isActive = activeValue === value;

  return (
    <button
      type="button"
      role="tab"
      aria-selected={isActive}
      className={cn(
        'inline-flex items-center justify-center px-3 py-1.5 text-sm font-medium rounded-md transition-all',
        isActive ? 'shadow-sm' : 'text-gray-500 hover:text-primary',
        className
      )}
      style={
        isActive ? { background: 'var(--dash-card)', color: 'var(--dash-text)', ...style } : style
      }
      onClick={() => onValueChange(value)}
      {...props}
    >
      {children}
    </button>
  );
};

interface TabsContentProps extends React.HTMLAttributes<HTMLDivElement> {
  value: string;
}

const TabsContent: React.FC<TabsContentProps> = ({ value, children, className, ...props }) => {
  const { value: activeValue } = useContext(TabsContext);
  if (activeValue !== value) return null;
  return (
    <div role="tabpanel" className={cn('mt-4', className)} {...props}>
      {children}
    </div>
  );
};

export { Tabs, TabsList, TabsTrigger, TabsContent };
