import { RefreshCw } from 'lucide-react';

export function refreshEmptyStateAction(label: string, onClick: () => void) {
  return {
    label,
    onClick,
    icon: RefreshCw,
    variant: 'outline' as const,
    size: 'md' as const,
  };
}
