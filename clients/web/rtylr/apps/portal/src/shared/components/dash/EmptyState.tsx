import type { EmptyStateProps } from '@voxire/ui';
import { Button } from '../ui/button';
import { cn } from '../../lib/utils';

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  className,
  size = 'default',
}: EmptyStateProps) {
  return (
    <div className={cn('portal-empty-state', size === 'compact' && 'is-compact', className)}>
      <div className="portal-empty-state-icon" aria-hidden="true">
        <Icon size={22} strokeWidth={1.6} />
      </div>
      <h3>{title}</h3>
      {description && <p>{description}</p>}
      <Button onClick={action.onClick}>{action.label}</Button>
    </div>
  );
}
