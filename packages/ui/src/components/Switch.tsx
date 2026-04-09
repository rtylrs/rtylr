import React from 'react';
import * as SwitchPrimitive from '@radix-ui/react-switch';
import { cn } from '../lib/cn';

interface SwitchProps extends SwitchPrimitive.SwitchProps {
  label?: string;
}

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitive.Root>,
  SwitchProps
>(({ className, label, id, ...props }, ref) => (
  <div className="flex items-center gap-2">
    <SwitchPrimitive.Root
      ref={ref}
      id={id}
      className={cn(
        'peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
        'disabled:cursor-not-allowed disabled:opacity-50',
        'data-[state=checked]:bg-accent data-[state=unchecked]:bg-black/15 dark:data-[state=unchecked]:bg-white/15',
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        className={cn(
          'pointer-events-none block h-4 w-4 rounded-full bg-white shadow-sm ring-0 transition-transform',
          'data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0'
        )}
      />
    </SwitchPrimitive.Root>
    {label && (
      <label htmlFor={id} className="text-sm font-medium cursor-pointer select-none">
        {label}
      </label>
    )}
  </div>
));

Switch.displayName = 'Switch';

export { Switch };
