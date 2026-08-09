import React from 'react';
import { cn } from '../../lib/utils';

export type LabelProps = React.LabelHTMLAttributes<HTMLLabelElement>;

const Label = React.forwardRef<HTMLLabelElement, LabelProps>(({ className, ...props }, ref) => (
  <label
    ref={ref}
    className={cn('mb-1.5 block text-sm font-medium leading-none', className)}
    style={{ color: 'var(--dash-text)' }}
    {...props}
  />
));

Label.displayName = 'Label';

export { Label };
