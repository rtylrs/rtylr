import React from 'react';
import { Switch } from '@voxire/ui';
import { cn } from '../lib/utils';

type FlowSwitchProps = React.ComponentProps<typeof Switch>;

/**
 * `@voxire/ui`'s Switch, given colours that resolve in this app.
 *
 * The component styles its track with `data-[state=checked]:bg-accent`, which
 * needs a Tailwind theme colour named `accent`. No rtylr app defines one: there
 * is no `@theme` block in any `globals.css`, despite the comment at the top of
 * each claiming the shared preset was ported into one. `bg-accent` therefore
 * computes to `rgba(0,0,0,0)` and a checked switch is indistinguishable from an
 * unchecked one - measured, not assumed.
 *
 * These are arbitrary-value utilities, which resolve without a theme block, and
 * they point at the same custom properties the rest of Flow uses. This is a local
 * workaround, not a fix: the real fix is a theme mapping in `@voxire/ui`'s
 * consumers, which also affects Alert, Badge, Avatar and Button's accent variant.
 * See docs/flow-time-tracking.md.
 */
const FlowSwitch = ({ className, ...props }: FlowSwitchProps): React.JSX.Element => (
  <Switch
    className={cn(
      'border-[var(--flow-border)] data-[state=checked]:bg-[var(--accent)] data-[state=unchecked]:bg-[var(--flow-active)]',
      className
    )}
    {...props}
  />
);

export default FlowSwitch;
